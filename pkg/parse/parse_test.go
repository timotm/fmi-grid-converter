package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	infile, err := os.Open("./sample.grib2")
	if err != nil {
		t.Error(err)
	}
	defer infile.Close()

	data, err := ioutil.ReadAll(infile)
	if err != nil {
		t.Error(err)
	}

	forecast, err := Parse(data)
	if err != nil {
		t.Error(err)
	}

	if len(forecast) != 1674 {
		t.Errorf("Expected 1674 forecasts, got %d", len(forecast))
	}

	randomPlaceKey := ForecastKey{lat: 59.920322999999996, lon: 24.168623999999998}
	randomPlace, exists := forecast[randomPlaceKey]
	if !exists {
		t.Errorf("Expected key %+v to exist", randomPlaceKey)
	}

	if len(randomPlace) != 5 {
		t.Errorf("Expected 5 forecasts for %+v, got %d", randomPlaceKey, len(randomPlace))
	}

	timestamp := time.Date(2022, time.August, 17, 12, 0, 0, 0, time.UTC)
	randomForecast, exists := randomPlace[timestamp]
	if !exists {
		t.Errorf("Expected forecast for %+v at %s to exist", randomPlaceKey, timestamp)
	}

	expected := ForecastItem{Temperature: 20.799988, WindSpeed: 7.4772997, WindDirection: 207, Precipitation: 0.1}
	if randomForecast != expected {
		t.Errorf("Expected forecast for %+v at %s to be %+v, got %+v", randomPlaceKey, timestamp, expected, randomForecast)
	}
}

type ForecastItemWithTs struct {
	Timestamp     time.Time `json:"ts"`
	Temperature   float32   `json:"t"`
	WindSpeed     float32   `json:"ws"`
	WindDirection int       `json:"wd"`
	Precipitation float32   `json:"p"`
}

type Location struct {
	Lat       float64              `json:"lat"`
	Lon       float64              `json:"lon"`
	Forecasts []ForecastItemWithTs `json:"forecasts"`
}

type FinalForecast struct {
	Locations []Location `json:"locations"`
}

func TestJsonOutput(t *testing.T) {
	f := make(Forecast)
	k1 := ForecastKey{lat: 59.920322999999996, lon: 24.168623999999998}
	k2 := ForecastKey{lat: 60.1, lon: 24.2}

	f[k1] = make(map[time.Time]ForecastItem)
	f[k2] = make(map[time.Time]ForecastItem)

	f[k1][time.Date(2022, time.August, 17, 12, 0, 0, 0, time.UTC)] = ForecastItem{Temperature: 1, WindSpeed: 2, WindDirection: 3, Precipitation: 4}
	f[k1][time.Date(2022, time.August, 17, 13, 0, 0, 0, time.UTC)] = ForecastItem{Temperature: 5, WindSpeed: 6, WindDirection: 7, Precipitation: 8}

	f[k2][time.Date(2022, time.August, 18, 12, 0, 0, 0, time.UTC)] = ForecastItem{Temperature: 9, WindSpeed: 10, WindDirection: 11, Precipitation: 12}

	jsonForecast, err := f.ToJson()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s\n", jsonForecast)

	var final FinalForecast
	err = json.Unmarshal(jsonForecast, &final)
	if err != nil {
		t.Error(err)
	}

	expectedPretty := `
	{
		"locations": [
		  {
			"lat": 59.920323,
			"lon": 24.168624,
			"forecasts": [
			  { "ts": "2022-08-17T12:00:00Z", "t": 1, "ws": 2, "wd": 3, "p": 4 },
			  { "ts": "2022-08-17T13:00:00Z", "t": 5, "ws": 6, "wd": 7, "p": 8 }
			]
		  },
		  {
			"lat": 60.1,
			"lon": 24.2,
			"forecasts": [
			  { "ts": "2022-08-18T12:00:00Z", "t": 9, "ws": 10, "wd": 11, "p": 12 }
			]
		  }
		]
	  }
	  `
	expected := &bytes.Buffer{}
	err = json.Compact(expected, []byte(expectedPretty))
	if err != nil {
		t.Error(err)
	}
	output, err := json.Marshal(final)
	if err != nil {
		t.Error(err)
	}
	if expected.String() != string(output) {
		t.Errorf("Expected %s, got %s", expected, output)
	}

	fmt.Println(string(output))
}
