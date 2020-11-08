package config

import (
	"fmt"
	"io/ioutil"
	"net"
	"solargo/inverter"
	"solargo/persistence"
	"solargo/weather"
	"solargo/yield_forecast"

	"gopkg.in/yaml.v2"
)

//Config of SolarGo
type Config struct {
	Debug     bool    `yaml:"debug"`
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
	Summary   struct {
		TelegramURL    string `yaml:"telegram_url"`
		BotToken       string `yaml:"bot_token"`
		ChatID         string `yaml:"chat_id"`
		SendStatistics bool   `yaml:"send_statistics"`
	} `yaml:"summary"`
	Inverter struct {
		IP       net.IP `yaml:"ip"`
		Port     uint16 `yaml:"port"`
		DeviceID string `yaml:"device_id"`
	} `yaml:"inverter"`
	Logging struct {
		Enabled  bool   `yaml:"enabled"`
		Filename string `yaml:"file_name"`
	} `yaml:"logging"`
	Persistence struct {
		URL          string `yaml:"url"`
		DatabaseName string `yaml:"database_name"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
	} `yaml:"persistence"`
	Weather struct {
		Enabled      bool   `yaml:"enabled"`
		Token        string `yaml:"api_token"`
		City         string `yaml:"city_code"`
		LanguageCode string `yaml:"language_code"`
	} `yaml:"weather"`
	Yield struct {
		Enabled   bool   `yaml:"enabled"`
		Token     string `yaml:"api_token"`
		Type      string `yaml:"type"`
		ID        string `yaml:"id"`
		Algorithm string `yaml:"algorithm"`
	} `yaml:"yield_forecast"`
}

//ReadConfig reads the provided config yaml
func ReadConfig(path string) Config {
	config := Config{}
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Can not read config file. Error: %s", err))
	}
	err = yaml.UnmarshalStrict([]byte(dat), &config)
	if err != nil {
		panic(fmt.Sprintf("Can not read config file. Error %s", err))
	}
	return config
}

//GetInverter from a config
func (config *Config) GetInverter() inverter.GenericInverter {
	var inverter inverter.FroniusSymo
	inverter.IP = config.Inverter.IP
	inverter.Port = config.Inverter.Port
	inverter.DeviceID = config.Inverter.DeviceID
	return &inverter
}

//GetDatabase from a config
func (config *Config) GetDatabase() persistence.GenericDatabase {
	var database persistence.Influx
	database.URL = config.Persistence.URL
	database.DatabaseName = config.Persistence.DatabaseName
	database.User = config.Persistence.User
	database.Password = config.Persistence.Password
	return &database
}

//GetWeatherService from a config
func (config *Config) GetWeatherService() weather.GenericWeather {
	var w weather.OpenWeather
	w.Token = config.Weather.Token
	w.LanguageCode = config.Weather.LanguageCode
	w.City = config.Weather.City
	w.URL = weather.OpenWeatherURL
	return &w
}

//GetYieldForecastService from a config
func (config *Config) GetYieldForecastService() yield_forecast.GenericYieldForecast {
	var s yield_forecast.SolarPrognose
	s.Token = config.Yield.Token
	s.Type = config.Yield.Type
	s.ID = config.Yield.ID
	s.Algorithm = config.Yield.Algorithm
	s.URL = yield_forecast.SolarPrognoseURL
	return &s
}
