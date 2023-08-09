package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	interfaceName := "lo"
	for {
		download1, upload1 := getNetworkStats(interfaceName)
		time.Sleep(5 * time.Second)
		download2, upload2 := getNetworkStats(interfaceName)

		downloadRate := float64(download2-download1) / 5
		uploadRate := float64(upload2-upload1) / 5

		fmt.Printf("Interface: %s\n", interfaceName)
		fmt.Printf("Download rate: %f MB/s\n", downloadRate/1024.0/1024.0)
		fmt.Printf("Upload rate: %f MB/s\n", uploadRate/1024.0/1024.0)
	}
}

func getNetworkStats(interfaceName string) (download, upload uint64) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		if i < 2 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		iface := fields[0]
		if iface == interfaceName+":" {
			download, _ = strconv.ParseUint(fields[1], 10, 64)
			upload, _ = strconv.ParseUint(fields[9], 10, 64)
			break
		}
	}

	return
}
