package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	NetworkStatsFilePath = "/proc/net/dev"
	MeasurementInterval  = 5 * time.Second
)

func main() {
	interfaceName := "lo"
	for {
		if !measureAndPrintNetworkSpeed(interfaceName) {
			// 這邊可以根據函數的返回值決定是否重試或退出
			time.Sleep(1 * time.Second)  // 遇到錯誤，等待1秒，重試
		}
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

	fmt.Printf("Interface: %s\n", interfaceName)
	fmt.Printf("Download rate: %f MB/s\n", downloadRate/1024.0/1024.0)
	fmt.Printf("Upload rate: %f MB/s\n", uploadRate/1024.0/1024.0)

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
