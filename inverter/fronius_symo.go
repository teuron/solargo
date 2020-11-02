//Package inverter contains implementations of varius solar inverters
//Here we implement the Fronius Symo Series inverter
package inverter

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

//FroniusSymo inverter
type FroniusSymo struct {
	IP       net.IP
	Port     uint16
	DeviceID string
}

//GetInverterStatistics of the Fronius inverter
func (f *FroniusSymo) GetInverterStatistics() (DailyStatistics, error) {
	var statistics DailyStatistics

	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetInverterRealtimeData.cgi?Scope=Device&DeviceID=%s&DataCollection=CumulationInverterData", f.IP.String(), f.Port, f.DeviceID)
	httpResult, err := http.Get(uri)

	if err != nil {
		return statistics, err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Body struct {
			Data struct {
				Day struct {
					Energy float64 `json:"Value"`
				} `json:"DAY_ENERGY"`
				Year struct {
					Energy float64 `json:"Value"`
				} `json:"YEAR_ENERGY"`
				Total struct {
					Energy float64 `json:"Value"`
				} `json:"TOTAL_ENERGY"`
			}
			DeviceStatus struct {
				StatusCode int64
				ErrorCode  int
			}
		}
		Head struct {
			Status struct {
				Code   int
				Reason string
			}
		}
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || result.Head.Status.Code != 0 {
		return statistics, fmt.Errorf("Error: %s, Inverter Reason: %s", err, result.Head.Status.Reason)
	}

	statistics.DailyProduction = WattHour(result.Body.Data.Day.Energy)
	statistics.YearlyProduction = WattHour(result.Body.Data.Year.Energy)
	statistics.TotalProduction = WattHour(result.Body.Data.Total.Energy)
	statistics.StatusCode = result.Body.DeviceStatus.StatusCode
	statistics.ErrorCode = ErrorCode(result.Body.DeviceStatus.ErrorCode)
	statistics.ErrorString = "" //TODO set right string

	return statistics, nil

}

//RetrieveData of the inverter
func (f *FroniusSymo) RetrieveData() (Data, error) {
	var data Data
	data.Info.Product = "Fronius Symo Series"
	data.Info.Object = "SolarGo"
	data.Info.Date = time.Now()
	data.Statistics.Date = time.Now()
	_, week := time.Now().ISOWeek()
	data.Statistics.Week = week
	_, month, _ := time.Now().Date()
	data.Statistics.Month = int(month)
	data.Statistics.WeekDay = time.Now().Weekday().String()

	if err := f.getAPIVersion(&data); err != nil {
		log.Info("Could not retrieve GetApiVersion", err)
		return data, err
	}

	if !cmp.Equal(data.Info.FirmWare, "1") {
		log.Info("Wrong Api-Version", data.Info.FirmWare)
		return data, fmt.Errorf("Wrong Api-Version %s", data.Info.FirmWare)
	}

	if err := f.powerFlowRealtimeData(&data); err != nil {
		log.Info("Could not retrieve PowerflowRealtimeData", err)
		return data, err
	}

	//It is ok to have an error here -> not everyone has the right meter
	if err := f.meterRealtimeData(&data); err != nil {
		log.Info("Could not retrieve PowerflowRealtimeData", err)
	}

	if err := f.inverterRealtimeData(&data); err != nil {
		log.Info("Could not retrieve InverterRealtimeData", err)
		return data, err
	}

	if err := f.inverterInfo(&data); err != nil {
		log.Info("Could not retrieve InverterRealtimeData", err)
		return data, err
	}

	if err := f.inverterCommonData(&data); err != nil {
		log.Info("Could not retrieve common inverter data", err)
		return data, err
	}

	if err := f.archiveData(&data); err != nil {
		log.Info("Could not retrieve archive data", err)
		return data, err
	}
	log.Info(fmt.Sprintf("Received the following data: %#v", data))
	return data, nil
}

func (f *FroniusSymo) powerFlowRealtimeData(data *Data) error {
	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetPowerFlowRealtimeData.fcgi", f.IP.String(), f.Port)
	httpResult, err := http.Get(uri)

	if err != nil {
		return err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Body struct {
			Data struct {
				Site struct {
					Day             float64 `json:"E_Day"`
					Year            float64 `json:"E_Year"`
					Total           float64 `json:"E_Total"`
					Location        string  `json:"Meter_Location"`
					Mode            string  `json:"Mode"`
					Grid            float64 `json:"P_Grid"`
					Load            float64 `json:"P_Load"`
					Battery         float64 `json:"P_Akku"`
					PV              float64 `json:"P_PV"`
					Autonomy        float64 `json:"rel_Autonomy"`
					SelfConsumption float64 `json:"rel_SelfConsumption"`
				}
			}
		}
		Head struct {
			Status struct {
				Code   int
				Reason string
			}
		}
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || result.Head.Status.Code != 0 {
		return fmt.Errorf("Error: %s, Inverter Reason: %s", err, result.Head.Status.Reason)
	}

	data.Sums.SumProdToday = WattHour(result.Body.Data.Site.Day)
	data.Sums.SumProdTotal = WattHour(result.Body.Data.Site.Total)
	data.Sums.SumProdYear = WattHour(result.Body.Data.Site.Year)
	data.Sums.SumPowerGrid = WattHour(result.Body.Data.Site.Grid)
	data.Sums.SumPowerLoad = WattHour(result.Body.Data.Site.Load)
	data.Sums.SumPowerBattery = WattHour(result.Body.Data.Site.Battery)
	data.Sums.SumPowerPv = WattHour(result.Body.Data.Site.PV)

	data.Service.MeterLocation = result.Body.Data.Site.Location
	data.Service.Mode = result.Body.Data.Site.Mode

	data.Service.Autonomy = result.Body.Data.Site.Autonomy
	data.Service.SelfConsumption = result.Body.Data.Site.SelfConsumption

	if data.Sums.SumPowerGrid < 0.0 {
		data.Meter.Feed = WattHour(math.Abs(float64(data.Sums.SumPowerGrid)))
		data.Meter.Purchased = 0.0
		data.Meter.Used = WattHour(math.Abs(float64(data.Sums.SumPowerLoad)))
	} else {
		data.Meter.Feed = 0.0
		data.Meter.Purchased = data.Sums.SumPowerGrid
		data.Meter.Used = WattHour(math.Abs(float64(data.Sums.SumPowerLoad)))
	}

	return nil
}

func (f *FroniusSymo) meterRealtimeData(data *Data) error {
	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetMeterRealtimeData.cgi?Scope=System", f.IP.String(), f.Port)
	httpResult, err := http.Get(uri)

	if err != nil {
		return err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Body struct {
			Data struct {
				Zero struct {
					Power            float64 `json:"PowerReal_P_Sum"`
					BlindPower       float64 `json:"PowerReactive_Q_Sum"`
					ApparentPower    float64 `json:"PowerApparent_S_Sum"`
					EnergyProduction float64 `json:"EnergyReal_WAC_Sum_Produced"`
					EnergyUsed       float64 `json:"EnergyReal_WAC_Sum_Consumed"`
				} `json:"0"`
				One struct {
					Power            float64 `json:"PowerReal_P_Sum"`
					BlindPower       float64 `json:"PowerReactive_Q_Sum"`
					ApparentPower    float64 `json:"PowerApparent_S_Sum"`
					EnergyProduction float64 `json:"EnergyReal_WAC_Sum_Produced"`
					EnergyUsed       float64 `json:"EnergyReal_WAC_Sum_Consumed"`
				} `json:"1"`
			}
		}
		Head struct {
			Status struct {
				Code   int
				Reason string
			}
		}
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || result.Head.Status.Code != 0 {
		return fmt.Errorf("Error: %s, Inverter Reason: %s", err, result.Head.Status.Reason)
	}

	//Check if channel is set or not
	if result.Body.Data.Zero.Power == 0.0 {
		data.Meter.Production = result.Body.Data.One.Power
		data.Meter.ApparentPower = result.Body.Data.One.ApparentPower
		data.Meter.BlindPower = result.Body.Data.One.BlindPower
		data.Meter.EnergyProduction = result.Body.Data.One.EnergyProduction
		data.Meter.EnergyUsed = result.Body.Data.One.EnergyUsed
	} else {
		data.Meter.Production = result.Body.Data.Zero.Power
		data.Meter.ApparentPower = result.Body.Data.Zero.ApparentPower
		data.Meter.BlindPower = result.Body.Data.Zero.BlindPower
		data.Meter.EnergyProduction = result.Body.Data.Zero.EnergyProduction
		data.Meter.EnergyUsed = result.Body.Data.Zero.EnergyUsed
	}

	return nil
}

func (f *FroniusSymo) inverterRealtimeData(data *Data) error {
	s, err := f.GetInverterStatistics()
	if err != nil {
		return err
	}
	data.Sums.ProductionToday = s.DailyProduction
	data.Sums.ProductionTotal = s.TotalProduction
	data.Sums.ProductionYear = s.YearlyProduction
	data.Service.ErrorCode = int(s.ErrorCode)
	data.Service.DeviceStatus = s.StatusCode
	data.Statistics.Production = s.DailyProduction
	return nil
}

func (f *FroniusSymo) getAPIVersion(data *Data) error {
	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetAPIVersion.cgi", f.IP.String(), f.Port)
	httpResult, err := http.Get(uri)

	if err != nil {
		return err
	}

	defer httpResult.Body.Close()

	type Result struct {
		APIVersion int
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("Error: %s", err)
	}

	data.Info.FirmWare = fmt.Sprintf("%d", result.APIVersion)
	return nil
}

func (f *FroniusSymo) inverterInfo(data *Data) error {
	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetInverterInfo.cgi", f.IP.String(), f.Port)
	httpResult, err := http.Get(uri)

	if err != nil {
		return err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Body struct {
			Data struct {
				WR struct {
					PVPower float64
				} `json:"1"`
			}
		}
		Head struct {
			Status struct {
				Code   int
				Reason string
			}
		}
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || result.Head.Status.Code != 0 {
		return fmt.Errorf("Error: %s, Inverter Reason: %s", err, result.Head.Status.Reason)
	}
	data.Service.PVPower = WattHour(result.Body.Data.WR.PVPower)
	return nil
}

func (f *FroniusSymo) inverterCommonData(data *Data) error {
	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetInverterRealtimeData.cgi?Scope=Device&DeviceID=%s&DataCollection=CommonInverterData", f.IP.String(), f.Port, f.DeviceID)
	httpResult, err := http.Get(uri)

	if err != nil {
		return err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Body struct {
			Data struct {
				PAC struct {
					Value float64
				}
				IAC struct {
					Value float64
				}
				UAC struct {
					Value float64
				}
				FAC struct {
					Value float64
				}
				IDC struct {
					Value float64
				}
				UDC struct {
					Value float64
				}
			}
		}
		Head struct {
			Status struct {
				Code   int
				Reason string
			}
		}
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || result.Head.Status.Code != 0 {
		return fmt.Errorf("Error: %s, Inverter Reason: %s", err, result.Head.Status.Reason)
	}
	data.AC.Voltage = result.Body.Data.UAC.Value
	data.AC.Current = result.Body.Data.IAC.Value
	data.AC.Frequency = result.Body.Data.FAC.Value
	data.AC.Power = WattHour(result.Body.Data.PAC.Value)
	data.PV.Voltage = result.Body.Data.UDC.Value
	data.PV.Current = result.Body.Data.IDC.Value
	data.PV.Power = WattHour(data.PV.Voltage * data.PV.Current)

	return nil
}

func (f *FroniusSymo) archiveData(data *Data) error {
	now := time.Now()
	//Just taken 400
	before := now.Add(-400 * time.Second)
	uri := fmt.Sprintf("http://%s:%d/solar_api/v1/GetArchiveData.cgi?Scope=System&StartDate=%s&EndDate=%s&Channel=Voltage_DC_String_1&Channel=Voltage_DC_String_2&Channel=Current_DC_String_1&Channel=Current_DC_String_2&Channel=Temperature_Powerstage", f.IP.String(), f.Port, before.Format(time.RFC3339), now.Format(time.RFC3339))
	httpResult, err := http.Get(uri)

	if err != nil {
		return err
	}

	defer httpResult.Body.Close()

	type Result struct {
		Body struct {
			Data struct {
				Inverter struct {
					Data struct {
						String1Voltage struct {
							Values float64
						} `json:"Voltage_DC_String_1"`
						String1Current struct {
							Values float64
						} `json:"Current_DC_String_1"`
						String2Voltage struct {
							Values float64
						} `json:"Voltage_DC_String_2"`
						String2Current struct {
							Values float64
						} `json:"Current_DC_String_2"`
						Temperature struct {
							Values float64
						} `json:"Temperature_Powerstage"`
					}
				} `json:"inverter/1"`
			}
		}
		Head struct {
			Status struct {
				Code   int
				Reason string
			}
		}
	}

	var result Result
	err = json.NewDecoder(httpResult.Body).Decode(&result)
	if err != nil || result.Head.Status.Code != 0 {
		return fmt.Errorf("Error: %s, Inverter Reason: %s", err, result.Head.Status.Reason)
	}

	data.PV.String1.Voltage = result.Body.Data.Inverter.Data.String1Voltage.Values
	data.PV.String1.Current = result.Body.Data.Inverter.Data.String1Current.Values
	data.PV.String2.Voltage = result.Body.Data.Inverter.Data.String2Voltage.Values
	data.PV.String2.Current = result.Body.Data.Inverter.Data.String2Current.Values
	data.Service.Temperature = result.Body.Data.Inverter.Data.Temperature.Values

	return nil
}
