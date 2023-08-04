package ping

import (
	"fmt"
	"os/exec"
	"strings"
)

// Delay 是可被其他 package 調用的函式，首字母要大寫
func Delay() {
	// 目標網站的 URL
	url := "www.google.com"

	// 執行 ping 測試，獲得延遲數據
	delay, err := pingTest(url)
	if err != nil {
		fmt.Println("無法進行 ping 測試:", err)
		return
	}

	fmt.Printf("Google 網站的延遲為: %s\n", delay)
	fmt.Println("URL:", url) // 輸出 URL 值
}

func pingTest(url string) (string, error) { //delay ping
	// 執行 ping 測試命令
	cmd := exec.Command("ping", "-c", "5", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// 將 ping 回應轉換成字串
	pingResponse := string(output)

	// 提取延遲數據
	// 注意：這只是一個簡單的示範，實際應用中需要更完善的處理方式
	lines := strings.Split(pingResponse, "\n")
	for _, line := range lines {
		if strings.Contains(line, "time=") {
			delay := strings.Split(line, "time=")[1]
			return delay, nil
		}
	}

	return "", fmt.Errorf("無法獲得延遲數據")
}
