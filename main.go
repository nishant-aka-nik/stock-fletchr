package main

import (
	"fletcher/config"
	"fletcher/scrapper"
	"fmt"
	"sort"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	volumeShockers, err := scrapper.ScrapeVolumeShockers(config.AppConfig.ScrapeURL)
	if err != nil {
		fmt.Println("Scrape error:", err)
		return
	}

	sort.Slice(volumeShockers, func(i, j int) bool {
		return volumeShockers[i].VolumeMultiple > volumeShockers[j].VolumeMultiple
	})

	limitTop(&volumeShockers, config.AppConfig.Limit)

	// Sorting by ChangePercent
	sort.Slice(volumeShockers, func(i, j int) bool {
		return volumeShockers[i].ChangePercent > volumeShockers[j].ChangePercent
	})

	for _, v := range volumeShockers {
		fmt.Printf("Name: %v,\nLTP: %v,\nChangePercent: %v,\nVolumeMultiple: %v \n\n", v.Name, v.LTp, v.ChangePercent, v.VolumeMultiple)
	}

	fmt.Println("tata bye bye")
}

// limitTop modifies the slice pointer to keep only the top 5 elements.
func limitTop(stocks *[]scrapper.VolumeShockerStock, limit int8) {
	if len(*stocks) > int(limit) {
		*stocks = (*stocks)[:limit]
	}
}
