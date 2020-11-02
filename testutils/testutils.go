package testutils

import (
	"fmt"
	"solargo/inverter"
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
