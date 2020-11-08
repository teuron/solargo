//Package yield_forecast contains a generic yield forecast interface and an solarprognose.de implementation
package yield_forecast

import (
	"solargo/inverter"
	"time"
)

//Data of a yield forecast
type Data struct {
	Date                 time.Time
	CurrentProduction    inverter.WattHour
	CummulatedProduction inverter.WattHour
}

//GenericYieldForecast provides an abstraction over a specific forecast source
type GenericYieldForecast interface {
	//RetrieveForecast
	RetrieveForecast() ([]Data, error)
}
