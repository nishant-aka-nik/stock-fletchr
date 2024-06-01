package scrapper

import (
	"fletcher/config"
	"fmt"
	"regexp"
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

func ScrapeVolumeShockers(url string) ([]VolumeShockerStock, error) {
	var volumeShockers []VolumeShockerStock

	c := colly.NewCollector(
		colly.UserAgent(config.AppConfig.Colly.UserAgent),
		colly.Async(false),
		colly.AllowURLRevisit(),
	)

	c.Limit(&colly.LimitRule{
		Delay:       time.Duration(config.AppConfig.Colly.DelaySeconds) * time.Second,
		RandomDelay: time.Duration(config.AppConfig.Colly.RandomDelaySeconds) * time.Second,
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

