package persistence

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"solargo/inverter"
	"solargo/weather"
	"solargo/yield_forecast"
	"testing"
	"time"
)

var wantedInfluxString = `Info Firmware=Firmware,Product="Product",Object="Object",Date="2009-11-10 23:00:00 +0000 UTC"
AC Voltage=1.000000,Current=2.000000,Frequency=3.000000,Power=4.000000
PV Voltage=5.000000,Current=6.000000,Power=7.000000,Voltage_String_1=8.000000,Current_String_1=9.000000,Voltage_String_2=10.000000,Current_String_2=11.000000
Service Status=12,Temperature=13.000000,ErrorCode=14,PVPower=15.000000,MeterLocation="unknown",Mode=battery,Autonomy=16.000000,SelfConsumption=17.000000
Statistics Date="2009-11-10 23:00:00 +0000 UTC",Week=19,Month=11,Production=18.000000,WeekDay="Wednesday"
Cummulations ProductionToday=20.000000,ProductionTotal=21.000000,ProductionYear=22.000000,SumProdToday=23.000000,SumProdTotal=24.000000,SumProdYear=25.000000,SumPowerGrid=26.000000,SumPowerLoad=27.000000,SumPowerBattery=28.000000,SumPowerPV=29.000000
Meter Production=30.000000,ApparentPower=32.000000,BlindPower=33.000000,EnergyProduction=34.000000,EnergyUsed=35.000000,Feed=36.000000,Purchase=37.000000,Usage=38.000000
`

var wantedWeatherString = `weather time="2009-11-10 23:00:00 +0000 UTC",location="XXX",sunrise="2010-11-10 23:00:00 +0000 UTC",sunset="2011-11-10 23:00:00 +0000 UTC",humidity=66.000000,temperature=14.170000,sky_description="Really nice sky. Color is blue?!?",wind_speed=1.350000,cloud_density=0.000000,wind_direction=3.000000,rain_amount=2.000000,snow_amount=1.000000,pressure=1024.000000
`

var wantedYieldForecastString = `yieldforecast date="10.11.2009",current_production=0.000000,cummulated_production=1.000000 1257894000
yieldforecast date="10.11.2010",current_production=2.000000,cummulated_production=3.000000 1289430000
`

var validProduction = `{"results":[{"statement_id":0,"series":[{"name":"AC","columns":["time","cumulative_sum"],"values":[["2020-11-21T12:32:00Z",3.3],["2020-11-21T12:33:00Z",4.4],["2020-11-21T12:34:00Z",5.5],["2020-11-21T12:35:00Z",6.6]]}]}]}`

var sampleWeather = weather.Data{
	LocationName:   "XXX",
	Date:           time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	Sunrise:        time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	Sunset:         time.Date(2011, time.November, 10, 23, 0, 0, 0, time.UTC),
	Humidity:       66,
	Temperature:    14.17,
	SkyDescription: "Really nice sky. Color is blue?!?",
	WindSpeed:      1.35,
	CloudDensity:   0,
	WindDirection:  3,
	RainAmount:     2.0,
	SnowAmount:     1.0,
	Pressure:       1024,
}

func getSampleInverterData() inverter.Data {
	now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	var data inverter.Data
	data.Info.FirmWare = "Firmware"
	data.Info.Product = "Product"
	data.Info.Object = "Object"
	data.Info.Date = now
	data.AC.Voltage = 1.0
	data.AC.Current = 2.0
	data.AC.Frequency = 3.0
	data.AC.Power = 4.0
	data.PV.Voltage = 5.0
	data.PV.Current = 6.0
	data.PV.Power = 7.0
	data.PV.String1.Voltage = 8.0
	data.PV.String1.Current = 9.0
	data.PV.String2.Voltage = 10.0
	data.PV.String2.Current = 11.0
	data.Service.DeviceStatus = 12
	data.Service.Temperature = 13.0
	data.Service.ErrorCode = 14
	data.Service.PVPower = 15.0
	data.Service.MeterLocation = "unknown"
	data.Service.Mode = "battery"
	data.Service.Autonomy = 16.0
	data.Service.SelfConsumption = 17.0
	data.Statistics.Date = now
	data.Statistics.Week = 19
	data.Statistics.Month = 11
	data.Statistics.Production = 18.0
	data.Statistics.WeekDay = "Wednesday"
	data.Sums.ProductionToday = 20.0
	data.Sums.ProductionTotal = 21.0
	data.Sums.ProductionYear = 22.0
	data.Sums.SumProdToday = 23.0
	data.Sums.SumProdTotal = 24.0
	data.Sums.SumProdYear = 25.0
	data.Sums.SumPowerGrid = 26.0
	data.Sums.SumPowerLoad = 27.0
	data.Sums.SumPowerBattery = 28.0
	data.Sums.SumPowerPv = 29.0
	data.Meter.Production = 30.0
	data.Meter.ApparentPower = 32.0
	data.Meter.BlindPower = 33.0
	data.Meter.EnergyProduction = 34.0
	data.Meter.EnergyUsed = 35.0
	data.Meter.Feed = 36.0
	data.Meter.Purchased = 37.0
	data.Meter.Used = 38.0
	return data
}

func getSampleYieldForecastData() []yield_forecast.Data {
	var data [2]yield_forecast.Data
	data[0].Date = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	data[0].CurrentProduction = inverter.WattHour(0)
	data[0].CummulatedProduction = inverter.WattHour(1)
	data[1].Date = time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	data[1].CurrentProduction = inverter.WattHour(2)
	data[1].CummulatedProduction = inverter.WattHour(3)
	return data[:]
}

func influxFromURL(url string) Influx {
	var db Influx
	db.URL = url
	db.DatabaseName = "dbname"
	db.User = "user"
	db.Password = "pw"
	return db
}

func TestInverterDataToInfluxData(t *testing.T) {

	var tests = []struct {
		testName string
		data     inverter.Data
		want     string
	}{
		{"Example Success", getSampleInverterData(), wantedInfluxString},
	}

	for _, tt := range tests {
		testname := tt.testName
		t.Run(testname, func(t *testing.T) {
			ans := inverterDataToInfluxData(tt.data)
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestWeatherDataToInfluxData(t *testing.T) {
	var tests = []struct {
		testName string
		data     weather.Data
		want     string
	}{
		{"Example Success", sampleWeather, wantedWeatherString},
	}

	for _, tt := range tests {
		testname := tt.testName
		t.Run(testname, func(t *testing.T) {
			ans := weatherToInfluxData(tt.data)
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestSendingWeather(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fail()
		}
		bodyString := string(bodyBytes)
		if bodyString != wantedWeatherString {
			t.Fail()
		}
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.SendWeather(sampleWeather)
}

func TestSendingInverter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fail()
		}
		bodyString := string(bodyBytes)
		if bodyString != wantedInfluxString {
			t.Fail()
		}
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.SendData(getSampleInverterData())
}

func TestSendingYieldForecast(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fail()
		}
		bodyString := string(bodyBytes)
		if bodyString != wantedYieldForecastString {
			print(bodyString)
			t.Fail()
		}
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.SendYieldForecast(getSampleYieldForecastData())
}

func TestSendingUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	err := db.persist("")

	if err.Error() != "Could not save data, because Username or Password is wrong" {
		t.Fatalf("Expected Error: %s", err)
	}
}

func TestSendingHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("HTTP ERROR")
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	err := db.persist("")

	if err.Error() != fmt.Sprintf("Post \"%s/write?db=dbname&precision=s\": EOF", ts.URL) {
		t.Fatalf("Expected Error: %s", err)
	}
}

func TestAuthThere(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val, ok := r.Header["Authorization"]
		if !ok || val[0] != "Basic dXNlcjpwdw==" {
			t.Fail()
		}
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.SendData(getSampleInverterData())
}

func TestAuthNotThereIfUsernameIsEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Header["Authorization"]
		if ok {
			t.Fail()
		}
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.User = ""
	db.SendData(getSampleInverterData())
}

func TestAuthNotThereIfPasswordIsEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Header["Authorization"]
		if ok {
			t.Fail()
		}
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.Password = ""
	db.SendData(getSampleInverterData())
}

func TestSendingInvalidRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	db.URL = "http://t . c o m"
	err := db.persist("")

	if err != nil && err.Error() != "parse \"http://t . c o m/write?db=dbname&precision=s\": invalid character \" \" in host name" {
		t.Fatalf("Expected Error: %s", err)
	}
}

func TestRetrieveProduction(t *testing.T) {
	var expected [4]ProductionStamps
	d, _ := time.Parse(time.RFC3339, "2020-11-21T12:32:00Z")
	expected[0].Date = d
	expected[0].Value = 3.3
	d, _ = time.Parse(time.RFC3339, "2020-11-21T12:33:00Z")
	expected[1].Date = d
	expected[1].Value = 4.4
	d, _ = time.Parse(time.RFC3339, "2020-11-21T12:34:00Z")
	expected[2].Date = d
	expected[2].Value = 5.5
	d, _ = time.Parse(time.RFC3339, "2020-11-21T12:35:00Z")
	expected[3].Date = d
	expected[3].Value = 6.6

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validProduction)
	}))
	defer ts.Close()

	db := influxFromURL(ts.URL)
	actual, err := db.GetTodaysProduction()

	if err != nil {
		t.Errorf("RetrieveProduction should not produce error %s", err)
	}
	for i := range actual {
		if !reflect.DeepEqual(actual[i], expected[i]) {
			t.Errorf("Error actual = %v\n, and expected = %v\n.", actual, expected)
		}
	}
}
