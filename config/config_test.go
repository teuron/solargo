package config

import (
	"fmt"
	"net"
	"reflect"
	"solargo/inverter"
	"solargo/persistence"
	"solargo/testutils"
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
		testname := fmt.Sprintf("%s", tt.path)
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
		testname := fmt.Sprintf("%s", tt.inverterName)
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
		testname := fmt.Sprintf("%s", tt.databaseName)
		t.Run(testname, func(t *testing.T) {
			ans := tt.config.GetDatabase()
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}
