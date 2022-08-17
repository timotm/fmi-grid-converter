package parse

import (
	"fmt"
	"log"
	"time"

	"github.com/amsokol/go-grib2"
)

type ForecastItem struct {
	Temperature   float32 `json:"temperature"`
	WindSpeed     float32 `json:"windSpeed"`
	WindDirection int     `json:"windDirection"`
	Precipitation float32 `json:"precipitation"`
}

type ForecastKey struct {
	lat float64
	lon float64
}

func (f ForecastKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%f/%f", f.lat, f.lon)), nil
}

type Forecast map[ForecastKey]map[time.Time]ForecastItem

func Parse(input []byte) (Forecast, error) {
	gribs, err := grib2.Read(input)
	if err != nil {
		return nil, err
	}

	forecast := make(Forecast)

	for _, g := range gribs {
		for _, v := range g.Values {
			lon := v.Longitude
			if lon > 180.0 {
				lon -= 360.0
			}

			key := ForecastKey{lat: v.Latitude, lon: lon}
			_, exists := forecast[key]
			if !exists {
				forecast[key] = make(map[time.Time]ForecastItem)
			}
			f, exists := forecast[key][g.VerfTime]
			if !exists {
				f = ForecastItem{}
			}

			switch g.Name {
			case "TMP":
				f.Temperature = v.Value - 273.15
			case "var192_140_242":
				f.WindDirection = int(v.Value)
			case "WIND":
				f.WindSpeed = v.Value
			case "var192_201_113":
				f.Precipitation = v.Value
			default:
				log.Printf("Unknown parameter: %s", g.Name)
			}
			forecast[key][g.VerfTime] = f
		}
	}

	return forecast, nil
}
