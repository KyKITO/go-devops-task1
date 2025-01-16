package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	loadAverageThreshold  = 30
	memoryUsageThreshold  = 80
	diskUsageThreshold    = 90
	networkUsageThreshold = 90
)

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"

	errorCount := 0

	response, err := http.Get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		errorCount++
		fmt.Println("Unable to fetch server statistic")
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errorCount++
		fmt.Println("Unable to fetch server statistic")
		return
	}

	data := strings.TrimSpace(string(body))
	values := strings.Split(data, ",")

	if len(values) != 6 {
		errorCount++
		fmt.Println("Unable to fetch server statistic")
		return
	}

	loadAverage, _ := strconv.ParseFloat(values[0], 64)
	memoryTotal, _ := strconv.ParseInt(values[1], 10, 64)
	memoryUsed, _ := strconv.ParseInt(values[2], 10, 64)
	diskTotal, _ := strconv.ParseInt(values[3], 10, 64)
	diskUsed, _ := strconv.ParseInt(values[4], 10, 64)
	networkCapacity, _ := strconv.ParseInt(values[5], 10, 64)

	checkLoadAverage(loadAverage)
	checkMemoryUsage(memoryTotal, memoryUsed)
	checkDiskUsage(diskTotal, diskUsed)
	checkNetworkUsage(networkCapacity)

	if errorCount >= 3 {
		fmt.Println("Unable to fetch server statistic")
	}
}

func checkLoadAverage(loadAverage float64) {
	if loadAverage > loadAverageThreshold {
		fmt.Printf("Load Average is too high: %.2f\n", loadAverage)
	}
}

func checkMemoryUsage(total int64, used int64) {
	memoryUsagePercent := (float64(used) / float64(total)) * 100
	if memoryUsagePercent > float64(memoryUsageThreshold) {
		fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsagePercent)
	}
}

func checkDiskUsage(total int64, used int64) {
	diskFreeSpaceMb := (total - used) / (1024 * 1024) // переводим байты в мегабайты
	diskUsagePercent := (float64(used) / float64(total)) * 100
	if diskUsagePercent > float64(diskUsageThreshold) {
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeSpaceMb)
	}
}

func checkNetworkUsage(capacity int64) {
	currentLoadBps := capacity                          // Здесь вы должны вставить вашу логику для получения текущей загруженности сети в байтах в секунду.
	freeBandwidthMbit := (capacity * 8) / (1024 * 1024) // переводим байты в мегабиты

	if freeBandwidthMbit < float64(networkUsageThreshold) {
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeBandwidthMbit)
	}
}
