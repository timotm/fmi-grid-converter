package parse

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/amsokol/go-grib2"
)

type ForecastItem struct {
	Temperature   float32 `json:"t"`
	WindSpeed     float32 `json:"ws"`
	WindDirection int     `json:"wd"`
	Precipitation float32 `json:"p"`
}

type ForecastKey struct {
	lat float64
	lon float64
}

func (f ForecastKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%f/%f", f.lat, f.lon)), nil
}

type Forecast map[ForecastKey]map[time.Time]ForecastItem

func (f Forecast) ToJson() ([]byte, error) {
	buf := &bytes.Buffer{}
	if _, err := buf.WriteString(`{"locations": [`); err != nil {
		return nil, err
	}
	locations := []string{}
	for k := range f {
		forecasts := []string{}
		for t, v := range f[k] {
			forecasts = append(forecasts, fmt.Sprintf(`{"ts": "%s", "t": %.1f, "ws": %.1f, "wd": %d, "p": %.1f}`, t.Format(time.RFC3339), v.Temperature, v.WindSpeed, v.WindDirection, v.Precipitation))
		}
		locations = append(locations, fmt.Sprintf(`{"lat": %f, "lon": %f, "forecasts": [%s]}`, k.lat, k.lon, strings.Join(forecasts, ", ")))
	}

	if _, err := buf.WriteString(strings.Join(locations, ", ")); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString("]}"); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

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
