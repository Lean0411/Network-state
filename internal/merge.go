package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	NetworkStatsFilePath = "/proc/net/dev"
	MeasurementInterval  = 1 * time.Second
	interfaceName        = "ens4"
	host                 = "www.google.com"
)

func main() {
	//go networkSpeedMonitor()
	//go networkQualityMonitor()

	select {} // 無窮迴圈，讓主 goroutine 不停止
}

func networkSpeedMonitor() {
	for {
		if !measureAndPrintNetworkSpeed(interfaceName) {
			time.Sleep(1 * time.Second) // 遇到錯誤，等待1秒，重試
		}
	}
}

func networkQualityMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	counter := 0

	for range ticker.C {
		if counter%5 == 0 { // 每5秒
			TestJitter(host)
			TestDelay(host)
		}
		TestPacketLossRate(host)
		counter++
	}
}

func measureAndPrintNetworkSpeed(interfaceName string) bool {
	download1, upload1, err := getNetworkStats(interfaceName)
	if handleError("initial", interfaceName, err) {
		return false
	}

	time.Sleep(MeasurementInterval)

	download2, upload2, err := getNetworkStats(interfaceName)
	if handleError("subsequent", interfaceName, err) {
		return false
	}

	if download1 > download2 || upload1 > upload2 {
		fmt.Printf("Error: data for %s seems to have wrapped around or reset\n", interfaceName)
		return false
	}

	downloadRate := float64(download2-download1) / float64(MeasurementInterval.Seconds())
	uploadRate := float64(upload2-upload1) / float64(MeasurementInterval.Seconds())
	totalBandwidthRate := downloadRate + uploadRate

	fmt.Printf("Interface: %s\n", interfaceName)
	fmt.Printf("Download rate: %f MB/s\n", downloadRate/1024.0/1024.0)
	fmt.Printf("Upload rate: %f MB/s\n", uploadRate/1024.0/1024.0)
	fmt.Printf("Total Bandwidth: %f MB/s\n", totalBandwidthRate/1024.0/1024.0)

	return true
}

func getNetworkStats(interfaceName string) (download, upload uint64, err error) {
	data, err := os.ReadFile(NetworkStatsFilePath)
	if err != nil {
		return 0, 0, fmt.Errorf("讀取 %s 失敗: %w", NetworkStatsFilePath, err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		iface := fields[0]
		if iface == interfaceName+":" {
			download, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("解析 %s 介面的下載數據失敗: %w", interfaceName, err)
			}

			upload, err = strconv.ParseUint(fields[9], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("解析 %s 介面的上傳數據失敗: %w", interfaceName, err)
			}
			return download, upload, nil
		}
	}
	return 0, 0, fmt.Errorf("未找到 %s 介面", interfaceName)
}

func handleError(stage, interfaceName string, err error) bool {
	if err != nil {
		fmt.Printf("在獲取 %s 介面的 %s 網絡統計時出錯: %s\n", interfaceName, stage, err)
		return true
	}
	return false
}

func TestJitter(host string) {
	jitter, err := jitter(host)
	if err != nil {
		fmt.Println("Jitter test error:", err)
	} else {
		fmt.Printf("Jitter: %.2f ms\n", jitter)
	}
}

func TestDelay(host string) {
	delay, err := delay(host)
	if err != nil {
		fmt.Println("delay test error:", err)
	} else {
		fmt.Printf("delay: %.2f ms\n", delay)
	}
}

func TestPacketLossRate(host string) {
	packetLossRate, err := packetLossRate(host)
	if err != nil {
		fmt.Println("packet loss test error:", err)
	} else {
		fmt.Printf("Pack loss rate: %.2f%%\n", packetLossRate)
	}
}

func jitter(host string) (float64, error) {
	pingResults, err := pingTest(host, 10)
	if err != nil {
		return 0, err
	}

	var sum float64
	var prevTime float64
	for i, time := range pingResults {
		if i > 0 {
			sum += abs(time - prevTime)
		}
		prevTime = time
	}

	return sum / float64(len(pingResults)-1), nil
}

func delay(host string) (float64, error) {
	pingResults, err := pingTest(host, 5)
	if err != nil {
		return 0, err
	}

	var sum float64
	for _, time := range pingResults {
		sum += time
	}

	return sum / float64(len(pingResults)), nil
}

func packetLossRate(host string) (float64, error) {
	totalPackets := 5
	lostPackets := 0

	for i := 0; i < totalPackets; i++ {
		cmd := exec.Command("ping", "-c", "1", "-s", "1", host)
		err := cmd.Run()

		if err != nil {
			lostPackets++
		}

		if i < totalPackets-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	lossRate := (float64(lostPackets) / float64(totalPackets)) * 100
	return lossRate, nil
}

func pingTest(host string, count int) ([]float64, error) {
	cmd := exec.Command("ping", "-c", strconv.Itoa(count), host)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("ping command failed")
	}

	return parsePingOutput(string(output)), nil
}

func parsePingOutput(output string) []float64 {
	lines := strings.Split(output, "\n")
	var results []float64
	for _, line := range lines {
		if strings.Contains(line, "time=") {
			parts := strings.Split(line, "time=")
			if len(parts) > 1 {
				timePart := parts[1]
				timeValue := strings.Split(timePart, " ")[0]
				timeFloat, err := strconv.ParseFloat(timeValue, 64)
				if err == nil {
					results = append(results, timeFloat)
				}
			}
		}
	}
	return results
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
