package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	for {
		netStats1, err := getNetworkStats()
		if err != nil {
			fmt.Println("Error getting network stats:", err)
			os.Exit(1)
		}

		time.Sleep(5 * time.Second)

		netStats2, err := getNetworkStats()
		if err != nil {
			fmt.Println("Error getting network stats:", err)
			os.Exit(1)
		}

		for iface, s1 := range netStats1 {
			s2, ok := netStats2[iface]
			if !ok {
				continue
			}

			downloadRate := float64(s2.download-s1.download) / 5
			uploadRate := float64(s2.upload-s1.upload) / 5

			fmt.Printf("Interface: %s\n", iface)
			fmt.Printf("  Download rate: %f MB/s\n", downloadRate/1000.0/1000.0)
			fmt.Printf("  Upload rate: %f MB/s\n", uploadRate/1000.0/1000.0)
		}
	}
}

type stats struct {
	download, upload uint64
}

func getNetworkStats() (map[string]stats, error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	netStats := make(map[string]stats)

	for i, line := range lines {
		if i < 2 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		iface := fields[0][:len(fields[0])-1]

		download, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return nil, err
		}

		upload, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return nil, err
		}

		netStats[iface] = stats{download, upload}
	}

	return netStats, nil
}
