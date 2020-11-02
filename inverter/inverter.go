package inverter

import (
	"fmt"
	"time"
)

//WattHour type
type WattHour float64

//KWh type
type KWh float64

//ErrorCode type
type ErrorCode int64

//DailyStatistics of an Inverter with error code
type DailyStatistics struct {
	DailyProduction  WattHour
	YearlyProduction WattHour
	TotalProduction  WattHour
	ErrorCode        ErrorCode
	StatusCode       int64
	ErrorString      string
}

//Data to be saved in the Database
type Data struct {
	Info struct {
		FirmWare string    //Firmware Version
		Product  string    //Product name
		Object   string    //SolarGo
		Date     time.Time //Current date
	}
	AC struct {
		Voltage   float64  //Voltage on Inverter AC side
		Current   float64  //Current on Inverter AC side
		Frequency float64  //Frequency on Inverter AC side
		Power     WattHour //Power on Inverter AC side
	}
	PV struct {
		Voltage float64  //Voltage on Inverter PV side
		Current float64  //Current on Inverter PV side
		Power   WattHour //Power on Inverter PV side
		String1 struct {
			Voltage float64 // Voltage of that string
			Current float64 // Current of that string
		}
		String2 struct { //Some inverter support two solar strings
			Voltage float64 // Voltage of that string
			Current float64 // Current of that string
		}
	}
	Service struct {
		DeviceStatus    int64    // Status of the inverter
		Temperature     float64  // Temperature in °C
		ErrorCode       int      // Error Code of the inverter
		PVPower         WattHour // Photovoltaic production
		MeterLocation   string   // Is the meter on "load" or "grid" or "unknown"
		Mode            string   // In what mode the inverter is operated
		Autonomy        float64  // Autonomy Degree in %
		SelfConsumption float64  // Selfconsumption of the produced electricity in %
	}
	Statistics struct {
		Date       time.Time // Current Time
		Week       int       // Todays Week
		Month      int       //	Todays Month
		Production WattHour  // Todays Production
		WeekDay    string    //	Todays Weekday
	}
	Sums struct {
		ProductionToday WattHour // Daily Production in Wh
		ProductionTotal WattHour // Total Production in Wh
		ProductionYear  WattHour // Yearly Production in Wh
		SumProdToday    WattHour // Daily Production in Wh
		SumProdTotal    WattHour // Total Production in Wh
		SumProdYear     WattHour // Yearly Production in Wh
		SumPowerGrid    WattHour // negative if we direct power to the grid, positive if we consume power from the grid
		SumPowerLoad    WattHour // negative if consuming power, positive if generating
		SumPowerBattery WattHour // negative if charging, positive if discharging
		SumPowerPv      WattHour // electricity production
	}
	Meter struct {
		Production       float64  // Current Production
		ApparentPower    float64  // Apparent Power
		BlindPower       float64  // Blind Power
		EnergyProduction float64  // Smart-Meter energy produced
		EnergyUsed       float64  // Smart-Meter energy used
		Feed             WattHour // Fed into the Grid
		Purchased        WattHour // Purchased from Grid
		Used             WattHour // Locally used power
	}
}

//GenericInverter provides an abstraction over a specific inverter
type GenericInverter interface {
	//GetInverterStatistics of the inverter
	GetInverterStatistics() (DailyStatistics, error)

	//RetrieveData of the inverter
	RetrieveData() (Data, error)
}

//ToKWh converts Wh to kWh
func (w *WattHour) ToKWh() KWh {
	return KWh(*w / WattHour(1000.0))
}

//Converts the Daily Statistics into a human readable form
func (s *DailyStatistics) String() string {
	return fmt.Sprintf("Daily Production: %.2f kWh\nYearly Production: %.2f kWh\nTotal Production: %.2f kWh", s.DailyProduction.ToKWh(), s.YearlyProduction.ToKWh(), s.TotalProduction.ToKWh())
}
