package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type VolumeShockerStock struct {
	Name           string
	LTp            float64
	ChangePercent  float64
	VolumeMultiple float64
}

func main() {
	//TODO: url needs to be given in config
	volumeShockers, err := scrapeVolumeShockers("https://trendlyne.com/stock-screeners/volume-based/high-volume-stocks/top-gainers/today/index/BSE500/")
	if err != nil {
		fmt.Println("Scrape error:", err)
		return
	}

	// Sorting by ChangePercent
	sort.Slice(volumeShockers, func(i, j int) bool {
		return volumeShockers[i].ChangePercent > volumeShockers[j].ChangePercent
	})

	//TODO: add limit in config
	limitTop(&volumeShockers, 3)

	sort.Slice(volumeShockers, func(i, j int) bool {
		return volumeShockers[i].VolumeMultiple > volumeShockers[j].VolumeMultiple
	})

	fmt.Println("Sorted by ChangePercent:", volumeShockers)
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
			ltp, errLTP := parseLTP(columns[1])
			changePercent, errCP := parseChangePercent(columns[2])
			volumeMultiple, errVM := parseVolumeMultiple(columns[5])

			if errCP == nil && errVM == nil && errLTP == nil {
				volumeShockers = append(volumeShockers, VolumeShockerStock{
					Name:           columns[0],
					ChangePercent:  changePercent,
					VolumeMultiple: volumeMultiple,
					LTp:            ltp,
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

func parseLTP(text string) (float64, error) {
	ltp, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("parse error: could not parse ltp in text err: %v", err)
	}
	return ltp, nil
}

// limitTop modifies the slice pointer to keep only the top 5 elements.
func limitTop(stocks *[]VolumeShockerStock, limit int8) {
	if len(*stocks) > int(limit) {
		*stocks = (*stocks)[:limit]
	}
}
