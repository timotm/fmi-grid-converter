package parse

import (
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

	for k := range forecast {
		fmt.Printf("%+v %+v\n", k, forecast[k])
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
