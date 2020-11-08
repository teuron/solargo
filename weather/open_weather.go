package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

//OpenWeatherURL Api endpoint
const OpenWeatherURL = "https://api.openweathermap.org"

//OpenWeather implementation of the GenericWeather interface
type OpenWeather struct {
	Token        string
	City         string
	LanguageCode string
	Latitude     float64
	Longitude    float64
	URL          string
}

//RetrieveForecast using the open weather map API
func (o *OpenWeather) RetrieveForecast() (Data, error) {
	var data Data
	uri := fmt.Sprintf("%s/data/2.5/weather?id=%s&APPID=%s&lang=%s&units=metric", o.URL, o.City, o.Token, o.LanguageCode)
	httpResult, err := http.Get(uri)

	if err != nil {
		return data, err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Location string `json:"name"`
		Clouds   struct {
			Density float64 `json:"all"`
		} `json:"clouds"`
		Main struct {
			Temperature float64 `json:"temp"`
			Pressure    float64 `json:"pressure"`
			Humidity    float64 `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed     float64 `json:"speed"`
			Direction float64 `json:"deg"`
		} `json:"wind"`
		Rain struct {
			ThreeHours float64 `json:"3h"`
		} `json:"rain"`
		Snow struct {
			ThreeHours float64 `json:"3h"`
		} `json:"snow"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || httpResult.StatusCode != http.StatusOK {
		return data, fmt.Errorf("Error while receiving weather data: %s", err)
	}

	rise, set := sunrise.SunriseSunset(
		o.Latitude, o.Longitude,
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
	)

	data.Date = time.Now()
	data.LocationName = result.Location
	data.Sunrise = rise
	data.Sunset = set
	data.Humidity = result.Main.Humidity
	data.Temperature = result.Main.Temperature
	if len(result.Weather) > 0 {
		data.SkyDescription = result.Weather[0].Description
	}
	data.WindSpeed = result.Wind.Speed
	data.CloudDensity = result.Clouds.Density
	data.WindDirection = result.Wind.Direction
	data.RainAmount = result.Rain.ThreeHours
	data.SnowAmount = result.Snow.ThreeHours
	data.Pressure = result.Main.Pressure

	return data, nil
}
