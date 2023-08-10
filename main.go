package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	host := "www.google.com"

	// 測試網路的 jitter
	TestJitter(host)

	// 測試網路的延遲
	TestDelay(host)

	// 測試網路的封包丟失率
	TestPacketLossRate(host)
}

// 測試網路的 jitter
func TestJitter(host string) {
	jitter, err := jitter(host)
	if err != nil {
		fmt.Println("Jitter 測試失敗:", err)
	} else {
		fmt.Printf("Jitter: %.2f ms\n", jitter)
	}
}

// 測試網路的延遲
func TestDelay(host string) {
	delay, err := delay(host)
	if err != nil {
		fmt.Println("延遲測試失敗:", err)
	} else {
		fmt.Printf("延遲: %.2f ms\n", delay)
	}
}

// 測試網路的封包丟失率
func TestPacketLossRate(host string) {
	packetLossRate, err := packetLossRate(host)
	if err != nil {
		fmt.Println("封包丟失率測試失敗:", err)
	} else {
		fmt.Printf("封包丟失率: %.2f%%\n", packetLossRate)
	}
}

// jitter 函數，測量網路的 jitter
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

// delay 函數，測量網路的平均延遲
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

// packetLossRate 函數，測量網路的封包丟失率
func packetLossRate(host string) (float64, error) {
	cmd := exec.Command("ping", "-c", "10", host)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "packet loss") {
			splitLine := strings.Split(line, ", ")
			lossPart := splitLine[2]
			lossPart = strings.TrimSuffix(lossPart, "% packet loss")
			lossRate, err := strconv.ParseFloat(lossPart, 64)
			if err != nil {
				return 0, err
			}
			return lossRate, nil
		}
	}

	return 0, errors.New("找不到封包丟失率")
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
