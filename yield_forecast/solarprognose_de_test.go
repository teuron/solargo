package yield_forecast

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"solargo/inverter"
	"strings"
	"testing"
	"time"
)

const (
	errorNotJSON    = `{not a json`
	errorJSONFormat = `{"data":{"XX1576735200":[0,0]}}`
	//Taken from https://www.solarprognose.de/web/de/solarprediction/page/api
	validResponse = `{"data":{"1576735200":[0,0],"1576738800":[0.064,0.064],"1576742400":[0.606,0.67]}}`
)

func TestSolarPrognoseJSONParsingError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, errorNotJSON)
	}))
	defer ts.Close()

	var s SolarPrognose
	s.URL = ts.URL

	tests := []struct {
		name string
		f    func() error
	}{
		{"RetrieveForecast", func() error { _, err := s.RetrieveForecast(); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if err == nil || !strings.HasPrefix(err.Error(), "Error while receiving yield forecast data: invalid character 'n' looking for beginning of object key string") {
				t.Errorf("SolarPrognose.de error = %v, want Prefix %s", err, "Error while receiving yield forecast data: invalid character 'n' looking for beginning of object key string")
			}
		})
	}
}

func TestSolarPrognoseInvalidJSONFormat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, errorJSONFormat)
	}))
	defer ts.Close()

	var s SolarPrognose
	s.URL = ts.URL

	tests := []struct {
		name string
		f    func() error
	}{
		{"RetrieveForecast", func() error { _, err := s.RetrieveForecast(); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if err == nil || !strings.HasPrefix(err.Error(), "Error trying to convert yield forecast data") {
				t.Errorf("SolarPrognose.de error = %v, want Prefix %s", err, "Error trying to convert yield forecast data")
			}
		})
	}
}

func TestSolarPrognoseInvalidURLError(t *testing.T) {
	var s SolarPrognose
	s.URL = "local host . com"

	tests := []struct {
		name string
		f    func() error
	}{
		{"RetrieveForecast", func() error { _, err := s.RetrieveForecast(); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f()
			if err == nil || !strings.HasPrefix(err.Error(), "Get \"local%20host%20.%20com/") {
				t.Errorf("SolarPrognose.de error = %v, want Prefix %s", err, "Get \"local%20host%20.%20com/")
			}
		})
	}
}

func TestSolarPrognoseSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validResponse)
	}))
	defer ts.Close()

	var s SolarPrognose
	s.URL = ts.URL

	actual, err := s.RetrieveForecast()

	if err != nil {
		t.Fatalf("Should not produce Error: %s", err)
	}

	var expected [3]Data
	expected[0].Date = time.Unix(1576735200, 0)
	expected[0].CurrentProduction = inverter.WattHour(0.0)
	expected[0].CummulatedProduction = inverter.WattHour(0.0)
	expected[1].Date = time.Unix(1576738800, 0)
	expected[1].CurrentProduction = inverter.WattHour(64)
	expected[1].CummulatedProduction = inverter.WattHour(64)
	expected[2].Date = time.Unix(1576742400, 0)
	expected[2].CurrentProduction = inverter.WattHour(606)
	expected[2].CummulatedProduction = inverter.WattHour(670)

	for i, v := range expected {
		if actual[i].Date.Unix() != v.Date.Unix() || actual[i].CurrentProduction != v.CurrentProduction || actual[i].CummulatedProduction != v.CummulatedProduction {
			t.Errorf("Error actual = %v\n, and expected = %v\n.", actual, expected)
			break
		}
	}
}
