package main

import (
	"time"
	"Networkstate/internal/networkstate"
)

const(
	interfaceName        = "ens4"
	host                 = "www.google.com"
	Period               = int64(1 * time.Second) // 週期性的時間
)

func main() {
    for {
        networkstate.TestJitter(host)
        networkstate.TestDelay(host)
        networkstate.TestPacketLossRate(host)

		
        for {
            if !networkstate.MeasureAndPrintNetworkSpeed(interfaceName) {
                time.Sleep(time.Duration(Period)) // 使用time.Duration進行轉換
            } else {
                break
            }
        }

        time.Sleep(time.Duration(Period)) 
    }
}