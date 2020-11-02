package persistence

import (
	"fmt"
	"reflect"
	"solargo/inverter"
	"testing"
	"time"
)

var wantedInfluxString = `Info Firmware=Firmware,Product="Product",Object="Object",Date="2009-11-10 23:00:00 +0000 UTC"
AC Voltage=1.000000,Current=2.000000,Frequency=3.000000,Power=4.000000
PV Voltage=5.000000,Current=6.000000,Power=7.000000,Voltage_String_1=8.000000,Current_String_1=9.000000,Voltage_String2=10.000000,Current_String_2=11.000000
Service Status=12,Temperature=13.000000,ErrorCode=14,PVPower=15.000000,MeterLocation="unknown",Mode=battery,Autonomy=16.000000,SelfConsumption=17.000000
Statistics Date="2009-11-10 23:00:00 +0000 UTC",Week=19,Month=11,Production=18.000000,WeekDay="Wednesday"
Cummulations ProductionToday=20.000000,ProductionTotal=21.000000,ProductionYear=22.000000,SumProdToday=23.000000,SumProdTotal=24.000000,SumProdYear=25.000000,SumPowerGrid=26.000000,SumPowerLoad=27.000000,SumPowerBattery=28.000000,SumPowerPV=29.000000
Meter Production=30.000000,ApparentPower=32.000000,BlindPower=33.000000,EnergyProduction=34.000000,EnergyUsed=35.000000,Feed=36.000000,Purchase=37.000000,Usage=38.000000
`

func TestInverterDataToInfluxData(t *testing.T) {
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

	var tests = []struct {
		testName string
		data     inverter.Data
		want     string
	}{
		{"Example Success", data, wantedInfluxString},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.testName)
		t.Run(testname, func(t *testing.T) {
			ans := inverterDataToInfluxData(tt.data)
			if !reflect.DeepEqual(ans, tt.want) {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}
