package persistence

import (
	"solargo/inverter"
	"solargo/weather"
	"solargo/yield_forecast"
	"time"
)

//ProductionStamps contains a solar production at a given time
type ProductionStamps struct {
	Date  time.Time
	Value inverter.WattHour
}

//GenericDatabase provides an abstraction over a specific database
type GenericDatabase interface {
	//SendData of the inverter to the database
	SendData(data inverter.Data)

	//SendWeather updates to the database
	SendWeather(data weather.Data)

	//SendYieldForecast updates to the database
	SendYieldForecast(data []yield_forecast.Data)

	//GetTodaysProduction from the database
	GetTodaysProduction() ([]ProductionStamps, error)
}
