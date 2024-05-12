package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Response struct {
	Name        string
	PageUrl     string
	DownloadUrl string
}

const dirPath = "./torFiles/"

func main() {
	responses := map[string]*Response{}

	torMeta := &Response{}
	// Create a new collector with options to mimic a browser
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"),
		colly.Async(false), // Enable asynchronous requests
	)

	listCollector := c.Clone()
	linkCollector := c.Clone()

	// Limit the rate of requests
	c.Limit(&colly.LimitRule{
		Delay:       2 * time.Second, // time between requests to the same domain
		RandomDelay: 2 * time.Second, // random delay added to the delay
	})

	// On every <a> element
	listCollector.OnHTML("a", func(e *colly.HTMLElement) {
		linkHref := e.Request.AbsoluteURL(e.Attr("href")) // Convert relative URL to absolute URL
		// Check if the link contains the specific substring
		if strings.Contains(linkHref, "/torrents/details/") && strings.Contains(e.Text, "720p") {
			linkText := strings.Replace(e.Text, " ", "_", -1)

			fmt.Printf("Target Link found: %q -> %s\n", linkText, linkHref)
			torMeta = &Response{
				Name:    linkText,
				PageUrl: linkHref,
			}
			responses[linkText] = torMeta
		}
	})

	// On every <a> element with the title "Click here to download torrent"

	// Before making a request print "Visiting ..."
	listCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	listCollector.OnScraped(func(r *colly.Response) {
		fmt.Println("Scraped", r.Request.URL.String())
		for key := range responses {
			x := responses[key].PageUrl
			torMeta = responses[key]
			filepath, _ := filepath.Abs(dirPath + torMeta.Name + ".torrent")
			_, err := os.Stat(filepath)

			if err == nil {
				fmt.Printf("File %s already exists. Skipping...\n", torMeta.Name)
				continue
			}
			linkCollector.Visit(x)
		}
	})

	linkCollector.OnHTML(`a[title="Click here to download torrent"]`, func(e *colly.HTMLElement) {
		linkHref := e.Request.AbsoluteURL(e.Attr("href"))
		torMeta.DownloadUrl = linkHref
		torFileName := strings.Replace(torMeta.Name, " ", "_", -1)
		responses[torFileName] = torMeta
		fmt.Printf("Link for download: %s\n", linkHref)
	})

	linkCollector.OnScraped(func(r *colly.Response) {
		fmt.Println("Scraped", r.Request.URL.String())
		fmt.Println("Download URL: ", torMeta.DownloadUrl)

		client := &http.Client{}

		// Create a new request
		req, err := http.NewRequest("GET", torMeta.DownloadUrl, nil)
		if err != nil {
			panic(err)
		}

		// Set User-Agent header
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

		// Send the request
		response, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			panic("Failed to download file: " + response.Status)
		}

		torFileName := strings.Replace(torMeta.Name, " ", "_", -1)
		filePath := fmt.Sprintf("%s%s.torrent", dirPath, torFileName)

		// Check if the directory exists, create if not
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
				panic(err)
			}
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, response.Body)
		if err != nil {
			panic(err)
		}

		println("File downloaded successfully.")
	})

	// Handle any errors
	listCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Start scraping on the desired page
	listCollector.Visit("")

	// Wait for asynchronous tasks to complete
	listCollector.Wait()
}
