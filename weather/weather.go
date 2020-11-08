//Package weather contains a generic weather interface and an OpenWeathermap implementation
package weather

import (
	"time"
)

//Data of a weather forecast
type Data struct {
	Date           time.Time
	LocationName   string
	CloudDensity   float64
	Temperature    float64
	Pressure       float64
	Humidity       float64
	SkyDescription string
	WindSpeed      float64
	WindDirection  float64
	Sunrise        time.Time
	Sunset         time.Time
	RainAmount     float64
	SnowAmount     float64
}

//GenericWeather provides an abstraction over a specific Weather source
type GenericWeather interface {
	//RetrieveForecast
	RetrieveForecast() (Data, error)
}
