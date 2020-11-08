package yield_forecast

import (
	"encoding/json"
	"fmt"
	"net/http"
	"solargo/inverter"
	"strconv"
	"time"
)

//SolarPrognoseURL Api endpoint
const SolarPrognoseURL = "https://www.solarprognose.de"

//SolarPrognose implementation of the GenericYieldForecast interface
type SolarPrognose struct {
	Token     string
	Type      string
	ID        string
	Algorithm string
	URL       string
}

//RetrieveForecast using the solarprognose.de API
func (o *SolarPrognose) RetrieveForecast() ([]Data, error) {
	var data []Data
	uri := fmt.Sprintf("%s/web/solarprediction/api/v1?access-token=%s&item=%s&id=%s&type=hourly&_format=json&algorithm=%s", o.URL, o.Token, o.Type, o.ID, o.Algorithm)
	httpResult, err := http.Get(uri)

	if err != nil {
		return data[:], err
	}
	defer httpResult.Body.Close()

	type Result struct {
		Data map[string][]float64 `json:"data"`
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || httpResult.StatusCode != http.StatusOK {
		return data[:], fmt.Errorf("Error while receiving yield forecast data: %s", err)
	}

	data = make([]Data, len(result.Data))
	cnt := 0
	for k, v := range result.Data {
		i, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return data, fmt.Errorf("Error trying to convert yield forecast data: %s", err)
		}
		data[cnt].Date = time.Unix(i, 0)
		data[cnt].CurrentProduction = inverter.WattHour(v[0] * 1000.0)
		data[cnt].CummulatedProduction = inverter.WattHour(v[1] * 1000.0)
		cnt++
	}
	return data[:], nil
}
