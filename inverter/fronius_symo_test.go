package inverter

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

const ValidResponse = `{
	"Body":{
	   "Data":{
		  "DeviceStatus:":{
			 "InverterState":"Running"
		  },
		  "DAY_ENERGY":{
			 "Unit":"W",
			 "Value": 1.0
		  },
		  "YEAR_ENERGY":{
			"Unit":"W",
			"Value": 2.0
		 },
		 "TOTAL_ENERGY":{
			"Unit":"W",
			"Value": 3.0
		 }
	   }
	},
	"Head":{
	   "RequestArguments":{
		  "DataCollection":"CumulationInverterData",
		  "DeviceClass":"Inverter",
		  "DeviceId":"1",
		  "Scope":"Device"
	   },
	   "Status":{
		  "Code":0,
		  "Reason":"",
		  "UserMessage":""
	   },
	   "Timestamp":"2019-08-28T05:59:13+00:00"
	}
 }`

const wrongStatus = `{
	"Body":{
	   "Data":{
		  "DeviceStatus:":{
			 "InverterState":"Running"
		  },
		  "PAC":{
			 "Unit":"W",
			 "Value":8.4296154682294417e+252
		  }
	   }
	},
	"Head":{
	   "RequestArguments":{
		  "DataCollection":"CumulationInverterData",
		  "DeviceClass":"Inverter",
		  "DeviceId":"1",
		  "Scope":"Device"
	   },
	   "Status":{
		  "Code":1,
		  "Reason":"TestReason",
		  "UserMessage":""
	   },
	   "Timestamp":"2019-08-28T05:59:13+00:00"
	}
 }`

func inverterFromURL(url *url.URL) FroniusSymo {
	var inverter FroniusSymo

	ip := net.ParseIP(url.Hostname())
	p, _ := strconv.ParseInt(url.Port(), 10, 64)

	inverter.DeviceID = "1"
	inverter.IP = ip
	inverter.Port = uint16(p)
	return inverter
}

func TestInverterStatistics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, ValidResponse)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	actual, err := inverter.GetInverterStatistics()

	if err != nil {
		t.Fatalf("Should not produce Error: %s", err)
	}

	var expected DailyStatistics
	expected.DailyProduction = 1.0
	expected.YearlyProduction = 2.0
	expected.TotalProduction = 3.0
	expected.ErrorCode = ErrorCode(0)

	if actual != expected {
		t.Errorf("Error actual = %v, and expected = %v.", actual, expected)
	}

}

func TestInverterStatisticsStatusError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, wrongStatus)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	_, err = inverter.GetInverterStatistics()

	if err == nil {
		t.Fatalf("Should have produced error")
	}
}

func TestInverterStatisticsNonValidPort(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, wrongStatus)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	inverter.Port = 0
	_, err = inverter.GetInverterStatistics()

	if err == nil {
		t.Fatalf("Should have produced error")
	}
}
