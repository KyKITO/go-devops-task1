package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	RequestTimeout    = 2 * time.Second
	RequestsFrequency = 300 * time.Millisecond
	ErrorThreshold    = 3
	ServerURL         = "http://srv.msk01.gigacorp.local/_stats"
	LoadThreshold     = 30
	MemoryThreshold   = 80
	DiskThreshold     = 90
	NetworkThreshold  = 90
)

type ServerStats struct {
	LoadAverage     int
	MemoryCapacity  int
	MemoryUsage     int
	DiskCapacity    int
	DiskUsage       int
	NetworkCapacity int
	NetworkUsage    int
}

func main() {
	poll := createServerPoller(ServerURL, RequestTimeout, RequestsFrequency, ErrorThreshold)
	analyze := createStatsAnalyzer(LoadThreshold, MemoryThreshold, DiskThreshold, NetworkThreshold)
	for response := range poll() {
		stats, err := parseStats(response)
		if err == nil {
			analyze(stats)
		}
	}
}

func createStatsAnalyzer(loadThreshold, memoryThreshold, diskThreshold, networkThreshold int) func(serverStats ServerStats) {
	return func(stats ServerStats) {
		memoryUsagePercent := (stats.MemoryUsage * 100) / stats.MemoryCapacity
		diskUsagePercent := (stats.DiskUsage * 100) / stats.DiskCapacity
		networkUsagePercent := (stats.NetworkUsage * 100) / stats.NetworkCapacity

		if stats.LoadAverage > loadThreshold {
			fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
		}
		if memoryUsagePercent > memoryThreshold {
			fmt.Printf("Memory usage too high: %d%%\n", memoryUsagePercent)
		}
		if diskUsagePercent > diskThreshold {
			availableSpace := (stats.DiskCapacity - stats.DiskUsage) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", availableSpace)
		}
		if networkUsagePercent > networkThreshold {
			availableBandwidth := ((stats.NetworkCapacity - stats.NetworkUsage) * 8) / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", availableBandwidth)
		}
	}
}

func parseStats(rawStats []byte) (ServerStats, error) {
	parts := strings.Split(strings.TrimSpace(string(rawStats)), ",")
	if len(parts) != 7 {
		return ServerStats{}, fmt.Errorf("invalid data length")
	}

	stats := ServerStats{}
	var err error
	stats.LoadAverage, err = strconv.Atoi(parts[0])
	stats.MemoryCapacity, err = strconv.Atoi(parts[1])
	stats.MemoryUsage, err = strconv.Atoi(parts[2])
	stats.DiskCapacity, err = strconv.Atoi(parts[3])
	stats.DiskUsage, err = strconv.Atoi(parts[4])
	stats.NetworkCapacity, err = strconv.Atoi(parts[5])
	stats.NetworkUsage, err = strconv.Atoi(parts[6])

	if err != nil {
		return ServerStats{}, err
	}
	return stats, nil
}

func createServerPoller(url string, reqTimeout, reqFreq time.Duration, errorThreshold int) func() chan []byte {
	return func() chan []byte {
		responsesChan := make(chan []byte)
		client := http.Client{Timeout: reqTimeout}
		errorCounter := 0

		go func() {
			defer close(responsesChan)
			for {
				time.Sleep(reqFreq)

				if errorCounter >= errorThreshold {
					fmt.Println("Unable to fetch server statistic")
					break
				}

				resp, err := client.Get(url)
				if err != nil || resp.StatusCode != http.StatusOK {
					errorCounter++
					if err != nil {
						fmt.Printf("Failed to send request: %v\n", err)
					}
					continue
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					errorCounter++
					fmt.Printf("Failed to parse response: %v\n", err)
					continue
				}
				resp.Body.Close()

				responsesChan <- body
				errorCounter = 0
			}
		}()

		return responsesChan
	}
}
