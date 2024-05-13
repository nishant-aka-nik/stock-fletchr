package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	ScrapeURL string `json:"scrapeURL"`
	Colly     Colly  `json:"colly"`
}

type Colly struct {
	UserAgent          string `json:"userAgent"`
	DelaySeconds       int    `json:"delaySeconds"`
	RandomDelaySeconds int    `json:"randomDelaySeconds"`
}

var AppConfig *Config

func LoadConfig() error {
	CONFIG_PATH, filepathErr := filepath.Abs("./config/config.json")
	if filepathErr != nil {
		return fmt.Errorf("could not find config file: %v", filepathErr)
	}

	file, err := os.Open(CONFIG_PATH)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	AppConfig = &Config{}
	err = json.Unmarshal(bytes, AppConfig)
	if err != nil {
		return err
	}

	return nil
}
