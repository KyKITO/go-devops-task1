package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func fetchStats() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://srv.msk01.gigacorp.local/_stats", nil)
	if err != nil {
		return "", err
	}
	req.Host = "srv.msk01.gigacorp.local"

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func analyzeData(data string) error {
	values := strings.Split(strings.TrimSpace(data), ",")
	if len(values) != 7 {
		return fmt.Errorf("invalid data format")
	}

	loadAverage, _ := strconv.Atoi(values[0])
	totalMemory, _ := strconv.ParseFloat(values[1], 64)
	usedMemory, _ := strconv.ParseFloat(values[2], 64)
	totalDisk, _ := strconv.ParseFloat(values[3], 64)
	usedDisk, _ := strconv.ParseFloat(values[4], 64)
	totalBw, _ := strconv.ParseFloat(values[5], 64)
	usedBw, _ := strconv.ParseFloat(values[6], 64)

	if loadAverage > 30 {
		fmt.Printf("Load Average is too high: %d\n", loadAverage)
	}

	memoryUsagePercent := (usedMemory / totalMemory) * 100
	if memoryUsagePercent > 80 {
		fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsagePercent)
	}

	freeDiskSpaceMb := (totalDisk - usedDisk) / (1024 * 1024)
	diskUsagePercent := (usedDisk / totalDisk) * 100
	if diskUsagePercent > 90 {
		fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDiskSpaceMb)
	}

	freeBandwidthMbps := ((totalBw - usedBw) * 8) / (1024 * 1024)
	bandwidthUsagePercent := (usedBw / totalBw) * 100
	if bandwidthUsagePercent > 90 {
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeBandwidthMbps)
	}

	return nil
}

func main() {
	errorCount := 0
	for {
		data, err := fetchStats()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
		} else {
			if err := analyzeData(data); err != nil {
				errorCount++
				if errorCount >= 3 {
					fmt.Println("Unable to fetch server statistic")
				}
			} else {
				errorCount = 0
			}
		}

		time.Sleep(60 * time.Second)
	}
}
