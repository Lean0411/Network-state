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
		download1, upload1, err := getNetworkStats(interfaceName)
		if err != nil {
			fmt.Printf("Error getting network stats for %s: %s\n", interfaceName, err)
			os.Exit(1)
		}

		time.Sleep(5 * time.Second)

		download2, upload2, err := getNetworkStats(interfaceName)
		if err != nil {
			fmt.Printf("Error getting network stats for %s: %s\n", interfaceName, err)
			os.Exit(1)
		}

		downloadRate := float64(download2-download1) / 5
		uploadRate := float64(upload2-upload1) / 5

		fmt.Printf("Interface: %s\n", interfaceName)
		fmt.Printf("Download rate: %f MB/s\n", downloadRate/1024.0/1024.0)
		fmt.Printf("Upload rate: %f MB/s\n", uploadRate/1024.0/1024.0)
	}
}

func getNetworkStats(interfaceName string) (download, upload uint64, err error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read /proc/net/dev: %w", err)
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
				return 0, 0, fmt.Errorf("failed to parse download data: %w", err)
			}

			upload, err = strconv.ParseUint(fields[9], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to parse upload data: %w", err)
			}
			return download, upload, nil
		}
	}

	return 0, 0, fmt.Errorf("interface %s not found", interfaceName)
}
