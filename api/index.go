package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/timotm/fmi-grid-converter/pkg/parse"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	timeFormat := "2006-01-02T15:04:05Z"

	startTime, err := time.Parse(timeFormat, r.FormValue("startTime"))
	if err != nil {
		error := fmt.Sprintf("Invalid startTime: %s", err.Error())
		log.Printf(error)
		http.Error(w, error, http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(timeFormat, r.FormValue("endTime"))
	if err != nil {
		error := fmt.Sprintf("Invalid endTime: %s", err.Error())
		log.Printf(error)
		http.Error(w, error, http.StatusBadRequest)
		return
	}

	bbox := r.FormValue("bbox")
	match := regexp.MustCompile(`^\d+\.\d+,\d+\.\d+,\d+\.\d+,\d+\.\d+$`).Match([]byte(bbox))
	if !match {
		log.Printf("Invalid bbox: %s", bbox)
		http.Error(w, "invalid bbox", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf(`https://opendata.fmi.fi/download?producer=harmonie_scandinavia_surface`+
		`&param=Temperature,WindDirection,WindSpeedMS,PrecipitationAmount`+
		`&bbox=%s`+
		`&starttime=%s`+
		`&endtime=%s`+
		`&format=grib2`+
		`&projection=EPSG:4326`+
		`&levels=0`+
		`&timestep=60`, bbox, startTime.Format(timeFormat), endTime.Format(timeFormat))

	res, err := http.Get(url)

	log.Printf("URL %s", url)
	if err != nil {
		log.Printf("Error fetching data: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		downloadError := strings.Join(res.Header["X-Download-Error"], ", ")
		log.Printf("Error downloading data: %s", downloadError)
		http.Error(w, downloadError, http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading data: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	forecast, err := parse.Parse(data)
	if err != nil {
		log.Printf("Error parsing data: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := forecast.ToJson()
	if err != nil {
		log.Printf("Error marshalling data: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
	log.Printf("Responded with %d bytes", len(output))
}
