package main

import (
	"io"
	"io/ioutil"
	"os"
	"time"

	"path/filepath"
	"solargo/config"
	"solargo/summary"

	"github.com/nathan-osman/go-sunrise"
	log "github.com/sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
)

const thirtyMinutes = 30 * time.Minute

var rise, set time.Time

func updateWeather(config *config.Config) {
	//If the time is between 30 min of sunrise and sunset, we can update the weather
	if time.Now().Add(-thirtyMinutes).After(rise) && time.Now().Add(thirtyMinutes).Before(set) && config.Weather.Enabled {
		log.Info("Update weather: ", time.Now().String())
		weather := config.GetWeatherService()
		database := config.GetDatabase()
		data, err := weather.RetrieveForecast()

		if err != nil {
			log.Error("Cannot read weather data: ", err)
			return
		}

		database.SendWeather(data)
	}
}

func readController(config *config.Config) {
	//If the time is between sunrise and sunset, we can read out the controller
	if time.Now().After(rise) && time.Now().Before(set) {
		log.Info("Reading the controller: ", time.Now().String())

		inverter := config.GetInverter()
		database := config.GetDatabase()

		data, err := inverter.RetrieveData()

		if err != nil {
			log.Error("Cannot read inverter data: ", err)
			return
		}

		database.SendData(data)
	}
}

func sendSummary(config *config.Config) {
	//If sunset is in less then 30 minutes, we send the summary
	in30min := time.Now().Add(thirtyMinutes)
	if in30min.After(set) && in30min.Before(set.Add(thirtyMinutes)) {
		log.Info("Send summary: ", time.Now().String())
		summary.SendSummary(config, config.GetInverter(), config.GetDatabase())
	}
}

func updateYieldForecast(config *config.Config) {
	//If sunrise is in less then 30 minutes and it is before sunset, we start to get the yield forecast
	in30min := time.Now().Add(thirtyMinutes)
	if in30min.After(rise) && in30min.Before(set) && config.Yield.Enabled {
		log.Info("Update Yield Forecast: ", time.Now().String())
		yield := config.GetYieldForecastService()
		database := config.GetDatabase()
		data, err := yield.RetrieveForecast()

		if err != nil {
			log.Error("Cannot read yield forecast data: ", err)
			return
		}

		database.SendYieldForecast(data)
	}
}

func sunriseSunset(config *config.Config) {
	rise, set = sunrise.SunriseSunset(
		config.Latitude, config.Longitude,
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
	)
	log.Info("Updated sunrise ", rise, "and sunset ", set)
}

func initializeLogger(config config.Config) {
	if !config.Logging.Enabled {
		log.SetOutput(ioutil.Discard)
	} else {
		_ = os.MkdirAll(filepath.Dir(config.Logging.Filename), os.ModePerm)
		logFile, err := os.OpenFile(config.Logging.Filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic("Could not open log file!")
		}
		if config.Debug {
			mw := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(mw)
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetOutput(logFile)
			log.SetLevel(log.ErrorLevel)
		}
	}

}

func main() {
	//Read config
	config := config.ReadConfig("config.yaml")

	//Initialize logger
	initializeLogger(config)

	//Start cron jobs
	c := cron.New()
	defer c.Stop()

	//At 0:10 update sunrise, sunset
	_, _ = c.AddFunc("10 0 * * *", func() { sunriseSunset(&config) })

	//Every 30 seconds read the controller
	_, _ = c.AddFunc("@every 0h0m30s", func() { readController(&config) })

	//Update the weather every half an hour
	_, _ = c.AddFunc("15,45 * * * *", func() { updateWeather(&config) })

	//Send the summary always around 30 minutes before sunset
	_, _ = c.AddFunc("@every 0h30m0s", func() { sendSummary(&config) })

	//Update the yield forecast every half an hour
	_, _ = c.AddFunc("@every 0h30m0s", func() { updateYieldForecast(&config) })

	c.Start()

	//On startup, run every function once
	sunriseSunset(&config)
	sendSummary(&config)
	readController(&config)
	updateWeather(&config)
	updateYieldForecast(&config)

	//Sleep forever
	select {}
}
