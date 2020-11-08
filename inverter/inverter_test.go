package inverter

import (
	"fmt"
	"math"
	"testing"
)

func TestConversionToKWh(t *testing.T) {
	tolerance := 0.001
	var tests = []struct {
		testName string
		data     WattHour
		want     KWh
	}{
		{"Success", WattHour(1000.0), KWh(1.0)},
		{"Success", WattHour(1500.0), KWh(1.5)},
		{"Success", WattHour(1234.56), KWh(1.23456)},
		{"Success", WattHour(1000000.0), KWh(1000.0)},
		{"Success", WattHour(0.0), KWh(0.0)},
		{"Success", WattHour(1.0), KWh(0.001)},
		{"Success", WattHour(-1000.0), KWh(-1.0)},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s: %vWh to %vkWh", tt.testName, tt.data, tt.want)
		t.Run(testname, func(t *testing.T) {
			ans := tt.data.ToKWh()
			if diff := math.Abs(float64(ans) - float64(tt.want)); diff > tolerance {
				t.Errorf("got %v, want %v, difference %f", ans, tt.want, diff)
			}
		})
	}
}

func TestDailyProductionToString(t *testing.T) {
	var tests = []struct {
		testName string
		data     DailyStatistics
		want     string
	}{
		{"Simple Case", DailyStatistics{
			DailyProduction:  WattHour(1000),
			YearlyProduction: WattHour(2000),
			TotalProduction:  WattHour(3000),
		}, "Daily Production: 1.00 kWh\nYearly Production: 2.00 kWh\nTotal Production: 3.00 kWh"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s: %vWh to %vkWh", tt.testName, tt.data, tt.want)
		t.Run(testname, func(t *testing.T) {
			ans := tt.data.String()
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}
