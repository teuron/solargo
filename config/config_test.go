package config

import (
	"net"
	"reflect"
	"solargo/inverter"
	"solargo/persistence"
	"solargo/testutils"
	"solargo/weather"
	"solargo/yield_forecast"
	"testing"
)

func TestReadConfig(t *testing.T) {
	var config Config
	config.Debug = true
	var emptyConfig Config
	var tests = []struct {
		path   string
		errors bool
		want   Config
	}{
		{"../testutils/config/not_there.yaml", true, emptyConfig},
		{"../testutils/config/not_valid.json", true, emptyConfig},
		{"../testutils/config/sample.yaml", false, config},
	}

	for _, tt := range tests {
		testname := tt.path
		t.Run(testname, func(t *testing.T) {
			if tt.errors {
				testutils.AssertPanic(t, func() { ReadConfig(tt.path) })
			} else {
				ans := ReadConfig(tt.path)
				if reflect.DeepEqual(ans, tt.want) {
					t.Errorf("got %v, want %v", ans, tt.want)
				}
			}
		})
	}
}

func TestGetInverter(t *testing.T) {
	var config Config
	config.Inverter.IP = net.IPv4(1, 2, 3, 4)
	config.Inverter.Port = 5678
	config.Inverter.DeviceID = "9"

	var fronius inverter.FroniusSymo
	fronius.IP = net.IPv4(1, 2, 3, 4)
	fronius.Port = 5678
	fronius.DeviceID = "9"

	var tests = []struct {
		inverterName string
		config       Config
		want         inverter.GenericInverter
	}{
		{"Fronius Symo", config, &fronius},
	}

	for _, tt := range tests {
		testname := tt.inverterName
		t.Run(testname, func(t *testing.T) {
			ans := tt.config.GetInverter()
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}

func TestGetDatabase(t *testing.T) {
	var config Config
	config.Persistence.URL = "http://localhost"
	config.Persistence.DatabaseName = "dbname"
	config.Persistence.User = "user"
	config.Persistence.Password = "pw"

	var influx persistence.Influx
	influx.URL = "http://localhost"
	influx.DatabaseName = "dbname"
	influx.User = "user"
	influx.Password = "pw"

	var tests = []struct {
		databaseName string
		config       Config
		want         persistence.GenericDatabase
	}{
		{"Influx DB", config, &influx},
	}

	for _, tt := range tests {
		testname := tt.databaseName
		t.Run(testname, func(t *testing.T) {
			ans := tt.config.GetDatabase()
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}

func TestGetWeatherService(t *testing.T) {
	var config Config
	config.Weather.Enabled = true
	config.Weather.Token = "token"
	config.Weather.City = "123"
	config.Weather.LanguageCode = "de"

	var w weather.OpenWeather
	w.Token = "token"
	w.City = "123"
	w.LanguageCode = "de"
	w.URL = weather.OpenWeatherURL

	var tests = []struct {
		weatherService string
		config         Config
		want           weather.GenericWeather
	}{
		{"OpenWeather", config, &w},
	}

	for _, tt := range tests {
		testname := tt.weatherService
		t.Run(testname, func(t *testing.T) {
			ans := tt.config.GetWeatherService()
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}

func TestGetYieldService(t *testing.T) {
	var config Config
	config.Yield.Enabled = true
	config.Yield.Token = "token"
	config.Yield.Type = "inverter"
	config.Yield.ID = "1234"
	config.Yield.Algorithm = "algo"

	var s yield_forecast.SolarPrognose
	s.Token = "token"
	s.Type = "inverter"
	s.ID = "1234"
	s.Algorithm = "algo"
	s.URL = yield_forecast.SolarPrognoseURL

	var tests = []struct {
		testName string
		config   Config
		want     yield_forecast.GenericYieldForecast
	}{
		{"Solarprognose", config, &s},
	}

	for _, tt := range tests {
		testname := tt.testName
		t.Run(testname, func(t *testing.T) {
			ans := tt.config.GetYieldForecastService()
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}
