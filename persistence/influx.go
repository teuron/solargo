package persistence

import (
	"fmt"
	"net/http"
	"solargo/inverter"
	"strings"

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
	pv := fmt.Sprintf("PV Voltage=%f,Current=%f,Power=%f,Voltage_String_1=%f,Current_String_1=%f,Voltage_String2=%f,Current_String_2=%f\n",
		p.Voltage, p.Current, p.Power, p.String1.Voltage, p.String1.Current, p.String2.Voltage, p.String2.Current)

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

	log.Info("Sending Data: ", res)
	return res
}

//SendData to the Influx Database
func (db *Influx) SendData(data inverter.Data) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/write?db=%s&precision=s", db.URL, db.DatabaseName), strings.NewReader(inverterDataToInfluxData(data)))
	if err != nil {
		log.Error("Could not create request: ", err)
		return
	}

	req.SetBasicAuth(db.User, db.Password)

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Could not save data: ", err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Error("Could not save data, because Username or Password is wrong: ", err)
		return
	}
}
