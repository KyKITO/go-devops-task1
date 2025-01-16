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
	poll := CreateServerPoller(ServerURL, RequestTimeout, RequestsFrequency, ErrorThreshold)
	analyze := CreateStatsAnalyzer(LoadThreshold, MemoryThreshold, DiskThreshold, NetworkThreshold)
	for response := range poll() {
		stats, err := ParseStats(response)
		if err == nil {
			analyze(stats)
		}
	}
}

func CreateStatsAnalyzer(loadThreshold, memoryThreshold, diskThreshold, networkThreshold int) func(ServerStats) {
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
			availableBandwidth := (stats.NetworkCapacity - stats.NetworkUsage) * 8 / 1024
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", availableBandwidth)
		}
	}
}

func ParseStats(rawStats []byte) (ServerStats, error) {
	parts := strings.Split(strings.TrimSpace(string(rawStats)), ",")
	if len(parts) != 7 {
		return ServerStats{}, fmt.Errorf("invalid data format")
	}
	stats := ServerStats{}
	for i, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return ServerStats{}, err
		}
		switch i {
		case 0:
			stats.LoadAverage = value
		case 1:
			stats.MemoryCapacity = value
		case 2:
			stats.MemoryUsage = value
		case 3:
			stats.DiskCapacity = value
		case 4:
			stats.DiskUsage = value
		case 5:
			stats.NetworkCapacity = value
		case 6:
			stats.NetworkUsage = value
		default:
		}
	}
	return stats, nil
}

func CreateServerPoller(url string, reqTimeout, reqFreq time.Duration, errorThreshold int) func() chan []byte {
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
