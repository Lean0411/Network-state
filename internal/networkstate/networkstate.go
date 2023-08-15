package networkstate

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
	NetworkStatsFilePath = "/proc/net/dev"  //讀取檔案路徑
	MeasurementInterval  = 1 * time.Second  //幾秒讀一次網路上傳下載
)


//輸出bandwidth
func MeasureAndPrintNetworkSpeed(interfaceName string) bool {
	download1, upload1, err := getNetworkStats(interfaceName)
	if handleError("initial", interfaceName, err) {
		return false
	}

	time.Sleep(MeasurementInterval)

	download2, upload2, err := getNetworkStats(interfaceName)
	if handleError("subsequent", interfaceName, err) {
		return false
	}
	//若後面取出來的值小於前面的值，代表錯誤
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

//讀取網路數據
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

// 測試網路的 jitter
func TestJitter(host string) {
	jitter, err := jitter(host)
	if err != nil {
		fmt.Println("Jitter test error:", err)
	} else {
		fmt.Printf("Jitter: %.2f ms\n", jitter)
	}
}

// 測試網路的延遲
func TestDelay(host string) {
	delay, err := delay(host)
	if err != nil {
		fmt.Println("delay test error:", err)
	} else {
		fmt.Printf("delay: %.2f ms\n", delay)
	}
}

// 測試網路的封包丟失率
func TestPacketLossRate(host string) {
	packetLossRate, err := packetLossRate(host)
	if err != nil {
		fmt.Println("packet loss test error:", err)
	} else {
		fmt.Printf("Pack loss rate: %.2f%%\n", packetLossRate)
	}
}

// jitter 函數，測量網路的 jitter
func jitter(host string) (float64, error) {
	pingResults, err := pingTest(host, 3)
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

// delay 函數，測量網路的平均延遲
func delay(host string) (float64, error) {
	pingResults, err := pingTest(host, 3)
	if err != nil {
		return 0, err
	}

	var sum float64
	for _, time := range pingResults {
		sum += time
	}

	return sum / float64(len(pingResults)), nil
}

// packetLossRate 函數，測量每秒的網路封包丟失率
func packetLossRate(host string) (float64, error) {
	// 在1秒內以200ms的時間間隔發送5個最小封包
	totalPackets := 5
	lostPackets := 0

	for i := 0; i < totalPackets; i++ {
		cmd := exec.Command("ping", "-c", "1", "-s", "1", host)
		err := cmd.Run()

		if err != nil {
			// 增加丟失的封包數
			lostPackets++
		}

		// 如果不是最後一次迴圈，等待200ms再發送下一個封包
		if i < totalPackets-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// 計算丟失率：(丟失的封包數/總封包數) * 100
	lossRate := (float64(lostPackets) / float64(totalPackets)) * 100
	return lossRate, nil
}


const maxRetries = 3
// pingTest 函數，執行 ping 測試
func pingTest(host string, count int) ([]float64, error) {
	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command("ping", "-c", strconv.Itoa(count), host)
		output, err := cmd.CombinedOutput() // 獲取 STDOUT 和 STDERR

		if err == nil {
			lines := strings.Split(string(output), "\n")
			var pingResults []float64
			for _, line := range lines {
				if strings.Contains(line, "time=") {
					timePart := strings.Split(line, "time=")[1]
					timePart = strings.Split(timePart, " ")[0]
					timeVal, err := strconv.ParseFloat(timePart, 64)
					if err != nil {
						return nil, fmt.Errorf("ping 輸出時發生錯誤: %v", err)
					}
					pingResults = append(pingResults, timeVal)
				}
			}
			if len(pingResults) == 0 {
				return nil, errors.New("找不到 ping 結果")
			}
			return pingResults, nil
		} else {
			fmt.Printf("Ping 測試失敗: %v\n", err) // 輸出錯誤訊息
			if i < maxRetries-1 {
				fmt.Println("重新ping", host, "中") 
				// 等一秒後再重試
				time.Sleep(1 * time.Second)
			}
		}
	}

	return nil, fmt.Errorf("ping 命令失敗達到上限")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

//錯誤資訊可以寫在這
func handleError(stage, interfaceName string, err error) bool {
	if err != nil {
		fmt.Printf("在獲取 %s 介面的 %s 網絡統計時出錯: %s\n", interfaceName, stage, err)
		return true
	}
	return false
}
