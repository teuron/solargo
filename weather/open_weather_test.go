package weather

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

const ValidResponse = `{
	"coord": {
	  "lon": 1.0,
	  "lat": 2.0
	},
	"weather": [
	  {
		"id": 800,
		"main": "Clear",
		"description": "Really nice sky. Color is blue?!?",
		"icon": "01d"
	  }
	],
	"base": "stations",
	"main": {
	  "temp": 14.17,
	  "feels_like": 12.74,
	  "temp_min": 12.22,
	  "temp_max": 15.56,
	  "pressure": 1024,
	  "humidity": 66
	},
	"visibility": 10000,
	"wind": {
	  "speed": 1.35,
	  "deg": 3
	},
	"rain": {
	  "3h": 2.0
	},
	"snow": {
	  "3h": 1.0
	},
	"clouds": {
	  "all": 0
	},
	"dt": 1234,
	"sys": {
	  "type": 3,
	  "id": 34534,
	  "country": "AT",
	  "sunrise": 456456456,
	  "sunset": 456456456
	},
	"timezone": 45645,
	"id": 4564564564,
	"name": "XXX",
	"cod": 200
  }`

func TestValidResponse(t *testing.T) {
	var o OpenWeather

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, ValidResponse)
	}))
	defer ts.Close()

	o.URL = ts.URL

	actual, err := o.RetrieveForecast()

	if err != nil {
		t.Fatalf("Should not produce Error: %s", err)
	}

	rise, set := sunrise.SunriseSunset(
		o.Latitude, o.Longitude,
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
	)

	var expected Data
	expected.LocationName = "XXX"
	expected.Date = actual.Date
	expected.Sunrise = rise
	expected.Sunset = set
	expected.Humidity = 66
	expected.Temperature = 14.17
	expected.SkyDescription = "Really nice sky. Color is blue?!?"
	expected.WindSpeed = 1.35
	expected.CloudDensity = 0
	expected.WindDirection = 3
	expected.RainAmount = 2.0
	expected.SnowAmount = 1.0
	expected.Pressure = 1024

	if actual != expected {
		t.Errorf("Error actual = %v, and expected = %v.", actual, expected)
	}
}

func TestInvalidStatusCode(t *testing.T) {
	var o OpenWeather

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	o.URL = ts.URL

	_, err := o.RetrieveForecast()

	if err == nil {
		t.Fatalf("Should produce Error")
	}
}

func TestHTTPError(t *testing.T) {
	var o OpenWeather

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("HTTP ERROR")
	}))
	defer ts.Close()

	o.URL = ts.URL

	_, err := o.RetrieveForecast()

	if err == nil {
		t.Fatalf("Should produce Error")
	}
}
