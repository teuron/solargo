package inverter

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	validRealtimeData          = `{"Body":{"Data":{"DeviceStatus":{"StatusCode":9,"ErrorCode":0},"DAY_ENERGY":{"Unit":"W","Value":1},"YEAR_ENERGY":{"Unit":"W","Value":2},"TOTAL_ENERGY":{"Unit":"W","Value":3}}},"Head":{"Status":{"Code":0}}}`
	validMeterZero             = `{"Body":{"Data":{"0":{"PowerReal_P_Sum":1,"PowerReactive_Q_Sum":2,"PowerApparent_S_Sum":3,"EnergyReal_WAC_Sum_Produced":4,"EnergyReal_WAC_Sum_Consumed":5}}},"Head":{"Status":{"Code":0}}}`
	validMeterOne              = `{"Body":{"Data":{"1":{"PowerReal_P_Sum":6,"PowerReactive_Q_Sum":7,"PowerApparent_S_Sum":8,"EnergyReal_WAC_Sum_Produced":9,"EnergyReal_WAC_Sum_Consumed":10}}},"Head":{"Status":{"Code":0}}}`
	validArchiveData           = `{"Body":{"Data":{"inverter/1":{"Data":{"Voltage_DC_String_1":{"Values":{"0":298}},"Current_DC_String_1":{"Values":{"0":297}},"Voltage_DC_String_2":{"Values":{"0":296}},"Current_DC_String_2":{"Values":{"0":295}},"Temperature_Powerstage":{"Values":{"0":45.5}}}}}},"Head":{"Status":{"Code":0}}}`
	validStatistics            = `{"Body":{"Data":{"DeviceStatus":{"StatusCode":9,"ErrorCode":0},"DAY_ENERGY":{"Unit":"W","Value":1},"YEAR_ENERGY":{"Unit":"W","Value":2},"TOTAL_ENERGY":{"Unit":"W","Value":3}}},"Head":{"Status":{"Code":0}}}`
	validPowerFlowRealtimeData = `{"Body":{"Data":{"Site":{"E_Day":10,"E_Year":11,"E_Total":12,"Meter_Location":"location","Mode":"mode","P_Grid":13,"P_Load":14,"P_Akku":15,"P_PV":16,"rel_Autonomy":17,"rel_SelfConsumption":18}}},"Head":{"Status":{"Code":0}}}`
	validAPIVersion            = "{\"APIVersion\": 1}"
	validCommonData            = `{"Body":{"Data":{"PAC":{"Value":100},"IAC":{"Value":101},"UAC":{"Value":102},"FAC":{"Value":103},"IDC":{"Value":104},"UDC":{"Value":105}}},"Head":{"Status":{"Code":0}}}`
	validInverterInfo          = `{"Body":{"Data":{"1":{"PVPower":1234.5}}},"Head":{"Status":{"Code":0}}}`
	errorStatus                = `{"Head":{"Status":{"Code":1,"Reason":"TestReason"}}}`
	errorNotJSON               = `{not a json`
	errorAPIVersion            = "{\"APIVersion\": 2}"
)

func inverterFromURL(url *url.URL) (i FroniusSymo) {
	i.IP = net.ParseIP(url.Hostname())
	p, _ := strconv.ParseInt(url.Port(), 10, 64)
	i.Port = uint16(p)
	return
}

func TestInverterStatistics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, validStatistics)
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
	expected.StatusCode = 9
	expected.ErrorCode = ErrorCode(0)

	if actual != expected {
		t.Errorf("Error actual = %v, and expected = %v.", actual, expected)
	}

}

func TestInverterRealtimeData(t *testing.T) {
	var expected Data
	expected.Sums.ProductionToday = 1.0
	expected.Sums.ProductionYear = 2.0
	expected.Sums.ProductionTotal = 3.0
	expected.Service.ErrorCode = 0
	expected.Service.DeviceStatus = 9
	expected.Statistics.Production = 1.0

	performTest(t, expected, validRealtimeData, func(inverter FroniusSymo, data *Data) error { return inverter.inverterRealtimeData(data) })
}

func TestFroniusSymoStatusError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, errorStatus)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	errorInner := "Error: %!s(<nil>), Inverter Reason: TestReason"
	errorGeneral := "Fronius Symo Error:"
	var data Data

	tests := []struct {
		name        string
		f           func(data *Data) error
		errorPrefix string
	}{
		{"PowerFlowRealtimeData", func(data *Data) error { return inverter.powerFlowRealtimeData(data) }, errorInner},
		{"MeterRealtimeData", func(data *Data) error { return inverter.meterRealtimeData(data) }, errorInner},
		{"InverterRealtimeData", func(data *Data) error { return inverter.inverterRealtimeData(data) }, errorInner},
		{"Inverter Info", func(data *Data) error { return inverter.inverterInfo(data) }, errorInner},
		{"Inverter Common Data", func(data *Data) error { return inverter.inverterCommonData(data) }, errorInner},
		{"Archive Data", func(data *Data) error { return inverter.archiveData(data) }, errorInner},
		{"GetStatistics", func(data *Data) error { _, err := inverter.GetInverterStatistics(); return err }, errorInner},
		{"RetrieveData", func(data *Data) error { _, err := inverter.RetrieveData(); return err }, errorGeneral},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f(&data); err == nil || !strings.HasPrefix(err.Error(), tt.errorPrefix) {
				t.Errorf("FroniusSymo error = %v, want Prefix %s", err, tt.errorPrefix)
			}
		})
	}
}

func TestFroniusSymoJSONParsingError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, errorNotJSON)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	var data Data

	tests := []struct {
		name string
		f    func(data *Data) error
	}{
		{"PowerFlowRealtimeData", func(data *Data) error { return inverter.powerFlowRealtimeData(data) }},
		{"MeterRealtimeData", func(data *Data) error { return inverter.meterRealtimeData(data) }},
		{"InverterRealtimeData", func(data *Data) error { return inverter.inverterRealtimeData(data) }},
		{"Inverter Info", func(data *Data) error { return inverter.inverterInfo(data) }},
		{"Inverter Common Data", func(data *Data) error { return inverter.inverterCommonData(data) }},
		{"Get API Version", func(data *Data) error { return inverter.getAPIVersion(data) }},
		{"Archive Data", func(data *Data) error { return inverter.archiveData(data) }},
		{"GetStatistics", func(data *Data) error { _, err := inverter.GetInverterStatistics(); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f(&data)
			if err == nil || !strings.HasPrefix(err.Error(), "Error: invalid character 'n' looking for beginning of object key string") {
				t.Errorf("FroniusSymo error = %v, want Prefix %s", err, "Error: invalid character 'n' looking for beginning of object key string")
			}
		})
	}
}

func TestFroniusSymoInvalidPort(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	inverter.Port = 0

	var data Data

	tests := []struct {
		name string
		f    func(data *Data) error
	}{
		{"PowerFlowRealtimeData", func(data *Data) error { return inverter.powerFlowRealtimeData(data) }},
		{"MeterRealtimeData", func(data *Data) error { return inverter.meterRealtimeData(data) }},
		{"InverterRealtimeData", func(data *Data) error { return inverter.inverterRealtimeData(data) }},
		{"Inverter Info", func(data *Data) error { return inverter.inverterInfo(data) }},
		{"Inverter Common Data", func(data *Data) error { return inverter.inverterCommonData(data) }},
		{"Archive Data", func(data *Data) error { return inverter.archiveData(data) }},
		{"Get API Version", func(data *Data) error { return inverter.getAPIVersion(data) }},
		{"GetStatistics", func(data *Data) error { _, err := inverter.GetInverterStatistics(); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f(&data)
			if err == nil || !strings.HasPrefix(err.Error(), "Get \"http://127.0.0.1:0") {
				t.Errorf("FroniusSymo error = %v, want a Prefix %s", err, "Get \"http://127.0.0.1:0")
			}
		})
	}
}

func TestInverterPowerFlowRealtimeDataPositiveGrid(t *testing.T) {
	var expected Data
	expected.Sums.SumProdToday = WattHour(10.0)
	expected.Sums.SumProdTotal = WattHour(12.0)
	expected.Sums.SumProdYear = WattHour(11.0)
	expected.Sums.SumPowerGrid = WattHour(13.0)
	expected.Sums.SumPowerLoad = WattHour(14.0)
	expected.Sums.SumPowerBattery = WattHour(15.0)
	expected.Sums.SumPowerPv = WattHour(16.0)
	expected.Service.MeterLocation = "location"
	expected.Service.Mode = "mode"
	expected.Service.Autonomy = 17.0
	expected.Service.SelfConsumption = 18.0
	expected.Meter.Feed = 0.0
	expected.Meter.Purchased = expected.Sums.SumPowerGrid
	expected.Meter.Used = WattHour(math.Abs(float64(expected.Sums.SumPowerLoad)))

	performTest(t, expected, validPowerFlowRealtimeData, func(inverter FroniusSymo, data *Data) error { return inverter.powerFlowRealtimeData(data) })
}

func TestInverterPowerFlowRealtimeDataNegativeGrid(t *testing.T) {
	var expected Data
	expected.Sums.SumProdToday = WattHour(10.0)
	expected.Sums.SumProdTotal = WattHour(12.0)
	expected.Sums.SumProdYear = WattHour(11.0)
	expected.Sums.SumPowerGrid = WattHour(-20.0)
	expected.Sums.SumPowerLoad = WattHour(14.0)
	expected.Sums.SumPowerBattery = WattHour(15.0)
	expected.Sums.SumPowerPv = WattHour(16.0)
	expected.Service.MeterLocation = "location"
	expected.Service.Mode = "mode"
	expected.Service.Autonomy = 17.0
	expected.Service.SelfConsumption = 18.0
	expected.Meter.Feed = WattHour(math.Abs(float64(expected.Sums.SumPowerGrid)))
	expected.Meter.Purchased = 0.0
	expected.Meter.Used = WattHour(math.Abs(float64(expected.Sums.SumPowerLoad)))

	validPowerFlowRealtimeDataNegative := `{"Body":{"Data":{"Site":{"E_Day":10,"E_Year":11,"E_Total":12,"Meter_Location":"location","Mode":"mode","P_Grid":-20,"P_Load":14,"P_Akku":15,"P_PV":16,"rel_Autonomy":17,"rel_SelfConsumption":18}}},"Head":{"Status":{"Code":0}}}`

	performTest(t, expected, validPowerFlowRealtimeDataNegative, func(inverter FroniusSymo, data *Data) error { return inverter.powerFlowRealtimeData(data) })
}

func TestInverterAPIVersion(t *testing.T) {
	var expected Data
	expected.Info.FirmWare = "1"

	performTest(t, expected, validAPIVersion, func(inverter FroniusSymo, data *Data) error { return inverter.getAPIVersion(data) })
}

func TestInverterCommonData(t *testing.T) {
	var expected Data
	expected.AC.Voltage = 102.0
	expected.AC.Current = 101.0
	expected.AC.Frequency = 103.0
	expected.AC.Power = WattHour(100.0)
	expected.PV.Voltage = 105.0
	expected.PV.Current = 104.0
	expected.PV.Power = WattHour(expected.PV.Voltage * expected.PV.Current)

	performTest(t, expected, validCommonData, func(inverter FroniusSymo, data *Data) error { return inverter.inverterCommonData(data) })
}

func TestInverterInfo(t *testing.T) {
	var expected Data
	expected.Service.PVPower = 1234.5

	performTest(t, expected, validInverterInfo, func(inverter FroniusSymo, data *Data) error { return inverter.inverterInfo(data) })
}

func TestInverterArchiveData(t *testing.T) {
	var expected Data
	expected.PV.String1.Voltage = 298.0
	expected.PV.String1.Current = 297.0
	expected.PV.String2.Voltage = 296.0
	expected.PV.String2.Current = 295.0
	expected.Service.Temperature = 45.5

	performTest(t, expected, validArchiveData, func(inverter FroniusSymo, data *Data) error { return inverter.archiveData(data) })
}

func TestInverterMeterRealtimeData(t *testing.T) {
	var expectedZero Data
	expectedZero.Meter.Production = 1
	expectedZero.Meter.ApparentPower = 3
	expectedZero.Meter.BlindPower = 2
	expectedZero.Meter.EnergyProduction = 4
	expectedZero.Meter.EnergyUsed = 5

	var expectedOne Data
	expectedOne.Meter.Production = 6
	expectedOne.Meter.ApparentPower = 8
	expectedOne.Meter.BlindPower = 7
	expectedOne.Meter.EnergyProduction = 9
	expectedOne.Meter.EnergyUsed = 10

	tests := []struct {
		name       string
		expected   Data
		httpResult string
	}{
		{"Zero", expectedZero, validMeterZero},
		{"One", expectedOne, validMeterOne},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			performTest(t, tt.expected, tt.httpResult, func(inverter FroniusSymo, data *Data) error { return inverter.meterRealtimeData(data) })
		})
	}
}

func performTest(t *testing.T, expected Data, httpResponse string, f func(inverter FroniusSymo, data *Data) error) {
	var actual Data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, httpResponse)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)

	if err := f(inverter, &actual); err != nil {
		t.Fatalf("Should not produce Error: %s", err)
	}

	if actual != expected {
		t.Errorf("Error actual = %v, and expected = %v.", actual, expected)
	}
}

func TestRetrieveDataSuccess(t *testing.T) {
	var expected Data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/solar_api/GetAPIVersion.cgi") {
			fmt.Fprintln(w, validAPIVersion)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetPowerFlowRealtimeData.fcgi") {
			fmt.Fprintln(w, validPowerFlowRealtimeData)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetMeterRealtimeData.cgi?Scope=System") {
			fmt.Fprintln(w, validMeterZero)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetInverterRealtimeData.cgi?Scope=Device&DeviceID=&DataCollection=CommonInverterData") {
			fmt.Fprintln(w, validCommonData)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetArchiveData.cgi?Scope=System&StartDate") {
			fmt.Fprintln(w, validArchiveData)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetInverterInfo.cgi") {
			fmt.Fprintln(w, validInverterInfo)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetInverterRealtimeData.cgi?Scope=Device&DeviceID=&DataCollection=CumulationInverterData") {
			fmt.Fprintln(w, validStatistics)
		} else {
			t.Fatalf("No other requests allowed.")
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	actual, err := inverter.RetrieveData()

	expected.Info.Date = actual.Info.Date
	expected.Statistics.Date = actual.Statistics.Date
	_, week := time.Now().ISOWeek()
	expected.Statistics.Week = week
	_, month, _ := time.Now().Date()
	expected.Statistics.Month = int(month)
	expected.Statistics.WeekDay = time.Now().Weekday().String()
	expected.Info.Product = "Fronius Symo Series"
	expected.Info.Object = "SolarGo"
	expected.Meter.Production = 1
	expected.Meter.ApparentPower = 3
	expected.Meter.BlindPower = 2
	expected.Meter.EnergyProduction = 4
	expected.Meter.EnergyUsed = 5
	expected.PV.String1.Voltage = 298.0
	expected.PV.String1.Current = 297.0
	expected.PV.String2.Voltage = 296.0
	expected.PV.String2.Current = 295.0
	expected.Service.Temperature = 45.5
	expected.Service.PVPower = 1234.5
	expected.AC.Voltage = 102.0
	expected.AC.Current = 101.0
	expected.AC.Frequency = 103.0
	expected.AC.Power = WattHour(100.0)
	expected.PV.Voltage = 105.0
	expected.PV.Current = 104.0
	expected.PV.Power = WattHour(expected.PV.Voltage * expected.PV.Current)
	expected.Info.FirmWare = "1"
	expected.Sums.SumProdToday = WattHour(10.0)
	expected.Sums.SumProdTotal = WattHour(12.0)
	expected.Sums.SumProdYear = WattHour(11.0)
	expected.Sums.SumPowerGrid = WattHour(13.0)
	expected.Sums.SumPowerLoad = WattHour(14.0)
	expected.Sums.SumPowerBattery = WattHour(15.0)
	expected.Sums.SumPowerPv = WattHour(16.0)
	expected.Service.MeterLocation = "location"
	expected.Service.Mode = "mode"
	expected.Service.Autonomy = 17.0
	expected.Service.SelfConsumption = 18.0
	expected.Meter.Feed = 0.0
	expected.Meter.Purchased = expected.Sums.SumPowerGrid
	expected.Meter.Used = WattHour(math.Abs(float64(expected.Sums.SumPowerLoad)))
	expected.Sums.ProductionToday = 1.0
	expected.Sums.ProductionYear = 2.0
	expected.Sums.ProductionTotal = 3.0
	expected.Service.ErrorCode = 0
	expected.Service.DeviceStatus = 9
	expected.Statistics.Production = 1.0

	if err != nil {
		t.Fatalf("Should not produce Error: %s", err)
	}

	if actual != expected {
		t.Errorf("Error actual = %v\n, and expected = %v\n.", actual, expected)
	}
}

func TestRetrieveDataNotSupportedAPIVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, errorAPIVersion)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	errorPrefix := "Fronius Symo Error: Wrong Api-Version 2"

	if _, err := inverter.RetrieveData(); err == nil || !strings.HasPrefix(err.Error(), errorPrefix) {
		t.Errorf("FroniusSymo error = %v, want Prefix %s", err, errorPrefix)
	}
}

func TestRetrieveDataInvalidAPIVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/solar_api/GetAPIVersion.cgi") {
			fmt.Fprintln(w, errorNotJSON)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetPowerFlowRealtimeData.fcgi") {
			fmt.Fprintln(w, validPowerFlowRealtimeData)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetMeterRealtimeData.cgi?Scope=System") {
			fmt.Fprintln(w, validMeterZero)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetInverterRealtimeData.cgi?Scope=Device&DeviceID=&DataCollection=CommonInverterData") {
			fmt.Fprintln(w, validCommonData)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetArchiveData.cgi?Scope=System&StartDate") {
			fmt.Fprintln(w, validArchiveData)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetInverterInfo.cgi") {
			fmt.Fprintln(w, validInverterInfo)
		} else if strings.HasPrefix(r.RequestURI, "/solar_api/v1/GetInverterRealtimeData.cgi?Scope=Device&DeviceID=&DataCollection=CumulationInverterData") {
			fmt.Fprintln(w, validStatistics)
		} else {
			t.Fatalf("No other requests allowed.")
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Could not parse httptest URL")
	}

	inverter := inverterFromURL(u)
	errorPrefix := "Fronius Symo Error: Error: invalid character 'n'"

	if _, err := inverter.RetrieveData(); err == nil || !strings.HasPrefix(err.Error(), errorPrefix) {
		t.Errorf("FroniusSymo error = %v, want Prefix %s", err, errorPrefix)
	}
}
