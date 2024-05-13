package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type VolumeShockerStock struct {
	Name           string
	ChangePercent  float64
	VolumeMultiple float64
}

func main() {
	volumeShockers, err := scrapeVolumeShockers("https://trendlyne.com/stock-screeners/volume-based/high-volume-stocks/top-gainers/today/index/BSE500/")
	if err != nil {
		fmt.Println("Scrape error:", err)
		return
	}
	for _, shocker := range volumeShockers {
		fmt.Printf("Name: %s, Change: %.2f%%, Volume Multiple: %.2f\n", shocker.Name, shocker.ChangePercent, shocker.VolumeMultiple)
	}
}

func scrapeVolumeShockers(url string) ([]VolumeShockerStock, error) {
	var volumeShockers []VolumeShockerStock

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"),
		colly.Async(false),
		colly.AllowURLRevisit(),
	)

	c.Limit(&colly.LimitRule{
		Delay:       2 * time.Second,
		RandomDelay: 2 * time.Second,
	})

	c.OnHTML("table tr", func(e *colly.HTMLElement) {
		var columns []string
		e.ForEach("td", func(_ int, el *colly.HTMLElement) {
			columns = append(columns, strings.TrimSpace(el.Text))
		})

		if len(columns) > 5 {
			changePercent, errCP := parseChangePercent(columns[2])
			volumeMultiple, errVM := parseVolumeMultiple(columns[5])
			if errCP == nil && errVM == nil {
				volumeShockers = append(volumeShockers, VolumeShockerStock{
					Name:           columns[0],
					ChangePercent:  changePercent,
					VolumeMultiple: volumeMultiple,
				})
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "Error:", err)
	})

	err := c.Visit(url)
	if err != nil {
		return nil, err
	}
	c.Wait()

	return volumeShockers, nil
}

func parseChangePercent(text string) (float64, error) {
	re := regexp.MustCompile(`\((\d+\.\d+) %\)`)
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		return strconv.ParseFloat(match[1], 64)
	}
	return 0, fmt.Errorf("parse error: could not find change percent in text")
}

func parseVolumeMultiple(text string) (float64, error) {
	re := regexp.MustCompile(`(\d+\.\d+)`)
	match := re.FindStringSubmatch(text)
	if len(match) > 0 {
		return strconv.ParseFloat(match[0], 64)
	}
	return 0, fmt.Errorf("parse error: could not find volume multiple in text")
}
