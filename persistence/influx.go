package persistence

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"solargo/inverter"
	"solargo/weather"
	"solargo/yield_forecast"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//Influx Database
type Influx struct {
	URL          string
	DatabaseName string
	User         string
	Password     string
}

//Converts the inverter data into an Influx query
func inverterDataToInfluxData(data inverter.Data) string {
	i := data.Info
	info := fmt.Sprintf("Info Firmware=%s,Product=\"%s\",Object=\"%s\",Date=\"%s\"\n", i.FirmWare, i.Product, i.Object, i.Date)

	a := data.AC
	ac := fmt.Sprintf("AC Voltage=%f,Current=%f,Frequency=%f,Power=%f\n", a.Voltage, a.Current, a.Frequency, a.Power)

	p := data.PV
	pv := fmt.Sprintf("PV Voltage=%f,Current=%f,Power=%f", p.Voltage, p.Current, p.Power)
	if p.String1.Voltage != 0 && p.String1.Current != 0 {
		pv += fmt.Sprintf(",Voltage_String_1=%f,Current_String_1=%f", p.String1.Voltage, p.String1.Current)
	}
	if p.String2.Voltage != 0 && p.String2.Current != 0 {
		pv += fmt.Sprintf(",Voltage_String_2=%f,Current_String_2=%f", p.String2.Voltage, p.String2.Current)
	}
	pv += "\n"

	s := data.Service
	service := fmt.Sprintf("Service Status=%d,Temperature=%f,ErrorCode=%d,PVPower=%f,MeterLocation=\"%s\",Mode=%s,Autonomy=%f,SelfConsumption=%f\n",
		s.DeviceStatus, s.Temperature, s.ErrorCode, s.PVPower, s.MeterLocation, s.Mode, s.Autonomy, s.SelfConsumption)

	s2 := data.Statistics
	statistics := fmt.Sprintf("Statistics Date=\"%s\",Week=%d,Month=%d,Production=%f,WeekDay=\"%s\"\n",
		s2.Date, s2.Week, s2.Month, s2.Production, s2.WeekDay)

	c := data.Sums
	cums := fmt.Sprintf("Cummulations ProductionToday=%f,ProductionTotal=%f,ProductionYear=%f,SumProdToday=%f,SumProdTotal=%f,SumProdYear=%f,SumPowerGrid=%f,SumPowerLoad=%f,SumPowerBattery=%f,SumPowerPV=%f\n",
		c.ProductionToday, c.ProductionTotal, c.ProductionYear, c.SumProdToday, c.SumProdTotal, c.SumProdYear, c.SumPowerGrid, c.SumPowerLoad, c.SumPowerBattery, c.SumPowerPv)

	m := data.Meter
	meter := fmt.Sprintf("Meter Production=%f,ApparentPower=%f,BlindPower=%f,EnergyProduction=%f,EnergyUsed=%f,Feed=%f,Purchase=%f,Usage=%f\n",
		m.Production, m.ApparentPower, m.BlindPower, m.EnergyProduction, m.EnergyUsed, m.Feed, m.Purchased, m.Used)

	res := fmt.Sprintf("%s%s%s%s%s%s%s", info, ac, pv, service, statistics, cums, meter)

	log.Info("Inverter Data: ", res)
	return res
}

func weatherToInfluxData(data weather.Data) string {
	res := fmt.Sprintf("weather time=\"%s\",location=\"%s\",sunrise=\"%s\",sunset=\"%s\",humidity=%f,temperature=%f,sky_description=\"%s\",wind_speed=%f,cloud_density=%f,wind_direction=%f,rain_amount=%f,snow_amount=%f,pressure=%f\n", data.Date, data.LocationName, data.Sunrise, data.Sunset, data.Humidity, data.Temperature, data.SkyDescription, data.WindSpeed, data.CloudDensity, data.WindDirection, data.RainAmount, data.SnowAmount, data.Pressure)

	log.Info("Weather Data: ", res)
	return res
}

func yieldToInfluxData(data []yield_forecast.Data) string {
	res := ""
	for _, d := range data {
		year, month, day := d.Date.Date()
		res += fmt.Sprintf("yieldforecast date=\"%d.%d.%d\",current_production=%f,cummulated_production=%f %d\n", day, month, year, d.CurrentProduction, d.CummulatedProduction, d.Date.Unix())
	}
	log.Info("Yield Data: ", res)
	return res
}

//SendData to the Influx Database
func (db *Influx) SendData(data inverter.Data) {
	_ = db.persist(inverterDataToInfluxData(data))
}

//SendWeather to the Influx Database
func (db *Influx) SendWeather(data weather.Data) {
	_ = db.persist(weatherToInfluxData(data))
}

//SendYieldForecast updates to the Influx Database
func (db *Influx) SendYieldForecast(data []yield_forecast.Data) {
	_ = db.persist(yieldToInfluxData(data))
}

func (db *Influx) persist(data string) error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/write?db=%s&precision=s", db.URL, db.DatabaseName), strings.NewReader(data))

	if err != nil {
		log.Error("Could not create request: ", err)
		return err
	}

	//Only add authentication if username and password is provided
	if db.User != "" && db.Password != "" {
		req.SetBasicAuth(db.User, db.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Could not save data: ", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Error("Could not save data, because Username or Password is wrong")
		return fmt.Errorf("Could not save data, because Username or Password is wrong")
	}
	return nil
}

func (db *Influx) GetTodaysProduction() ([]ProductionStamps, error) {
	var ps []ProductionStamps

	year, month, day := time.Now().Date()
	query := url.QueryEscape(fmt.Sprintf(`SELECT cumulative_sum(integral("Power"))  / 3600 FROM "AC" WHERE time < now() and time >= '%d-%02d-%02dT00:00:00Z' GROUP BY time(1m)`, year, month, day))

	uri := fmt.Sprintf("%s/query?db=%s&q=%s", db.URL, db.DatabaseName, query)
	httpResult, err := http.Get(uri)
	print(uri)
	if err != nil {
		return ps, err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Results []struct {
			Series []struct {
				Values [][]interface{} `json:"values"`
			} `json:"series"`
		} `json:"results"`
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil {
		return ps, fmt.Errorf("Error: %s", err)
	}

	//Create array of appropriate size
	log.Warn("result", result)

	ps = make([]ProductionStamps, len(result.Results[0].Series[0].Values))

	//Parse the values
	for idx, v := range result.Results[0].Series[0].Values {
		t, _ := time.Parse(time.RFC3339, v[0].(string))
		ps[idx].Date = t
		ps[idx].Value = inverter.WattHour(v[1].(float64))
	}
	return ps, nil
}
