package testutils

import (
	"fmt"
	"solargo/inverter"
	"solargo/persistence"
	"solargo/weather"
	"solargo/yield_forecast"
	"testing"
)

//AssertPanic tests if the provided function panics
func AssertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The function did not panic")
		}
	}()
	f()
}

//SuccessInverter used for testing
type SuccessInverter struct {
}

//GetInverterStatistics always returns the standard statistics
func (f *SuccessInverter) GetInverterStatistics() (inverter.DailyStatistics, error) {
	var statistics inverter.DailyStatistics
	return statistics, nil

}

//RetrieveData always returns the standard data
func (f *SuccessInverter) RetrieveData() (inverter.Data, error) {
	var data inverter.Data

	return data, nil
}

//ErrorInverter used for testing
type ErrorInverter struct {
}

//GetInverterStatistics always produces an error
func (f *ErrorInverter) GetInverterStatistics() (inverter.DailyStatistics, error) {
	var statistics inverter.DailyStatistics
	return statistics, fmt.Errorf("Error")

}

//RetrieveData always produces an error
func (f *ErrorInverter) RetrieveData() (inverter.Data, error) {
	var data inverter.Data

	return data, fmt.Errorf("Error")
}

//SuccessDatabase used for testing
type SuccessDatabase struct{}

//SendData of the inverter to nowhere
func (db *SuccessDatabase) SendData(data inverter.Data) {}

//SendWeather updates nothing
func (db *SuccessDatabase) SendWeather(data weather.Data) {}

//SendYieldForecast updates nothing
func (db *SuccessDatabase) SendYieldForecast(data []yield_forecast.Data) {}

//GetTodaysProduction from nothing
func (db *SuccessDatabase) GetTodaysProduction() ([]persistence.ProductionStamps, error) {
	var ps []persistence.ProductionStamps
	return ps, nil
}
