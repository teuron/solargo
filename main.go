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

func updateWeather() {
	log.Info("Update weather: ", time.Now().String())
}

func readController(config *config.Config) {
	rise, set := sunriseSunset(config)
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
	_, set := sunriseSunset(config)
	//If sunset is in less then 30 minutes, we send the summary
	in30min := time.Now().Add(thirtyMinutes)
	if in30min.After(set) && in30min.Before(set.Add(thirtyMinutes)) {
		log.Info("Send summary: ", time.Now().String())
		summary.SendSummary(config, config.GetInverter())
	}
}

func sunriseSunset(config *config.Config) (time.Time, time.Time) {
	return sunrise.SunriseSunset(
		config.Latitude, config.Longitude,
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
	)
}

func initializeLogger(config config.Config) {
	if !config.Logging.Enabled {
		log.SetOutput(ioutil.Discard)
	} else {
		os.MkdirAll(filepath.Dir(config.Logging.Filename), os.ModePerm)
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

	//Every 30 seconds read the controller
	c.AddFunc("@every 0h0m30s", func() { readController(&config) })

	//Update the weather every half an hour
	c.AddFunc("15,45 * * * *", updateWeather)

	//Send the summary always around 30 minutes before sunset
	c.AddFunc("@every 0h30m0s", func() { sendSummary(&config) })
	c.Start()

	//On startup, run every function once
	sendSummary(&config)
	readController(&config)
	updateWeather()

	//Sleep forever
	select {}
}
