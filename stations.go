package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sphireco/mantis"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Stations struct {
	ExecutionTime   string    `json:"executionTime"`
	StationBeanList []Station `json:"stationBeanList"`
}

type Station struct {
	Id                    int     `json:"id"`
	StationName           string  `json:"stationName"`
	AvailableDocks        int     `json:"availableDocks"`
	TotalDocks            int     `json:"totalDocks"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	StatusValue           string  `json:"statusValue"`
	StatusKey             int     `json:"statusKey"`
	AvailableBikes        int     `json:"availableBikes"`
	Address1              string  `json:"stAddress1"`
	Address2              string  `json:"stAddress2"`
	City                  string  `json:"city"`
	PostalCode            string  `json:"postalCode"`
	Location              string  `json:"location"`
	Altitude              string  `json:"altitude"`
	TestStation           bool    `json:"testStation"`
	LastCommunicationTime string  `json:"lastCommunicationTime"`
	Landmark              string  `json:"landMark"`
}

type ShortStation struct {
	StationName    string `json:"stationName"`
	Address        string `json:"address"`
	AvailableDocks int    `json:"availableDocks"`
	TotalDocks     int    `json:"totalDocks"`
}

type BikesToReturn struct {
	Dockable bool   `json:"dockable"`
	Message  string `json:"message"`
}

const (
	StatusOk    int = 1
	StatusNotOk int = 3
)

// getJSON
func (S *Stations) getJSON() []Station {
	var body []byte
	var err error

	cacheKey := "citibike-json"
	body, err = App.Cache.Get(cacheKey)

	if err != nil {
		res, err := http.Get("https://www.citibikenyc.com/stations/json")
		mantis.HandleError("getJSON:httpGet", err)

		body, err = ioutil.ReadAll(res.Body)
		mantis.HandleError("getJSON:ReadAll", err)

		err = App.Cache.Set(cacheKey, body)
		mantis.HandleError("getJSON:SetCache", err)

		// We have neither something cached, nor fetchable data, fail with empty list
		if err != nil {
			mantis.HandleError("getJSON:PostRead", errors.New("could not load json"))
			return S.StationBeanList
		}
	}

	err = json.Unmarshal(body, &S)
	mantis.HandleError("getJSON:JSONUnmarshal", err)

	return S.StationBeanList
}

// page
func page(r *http.Request, response []ShortStation) []ShortStation {
	responseLength := len(response)
	perPage := 20

	// Allow count per page
	queryParam := mantis.GetQueryParameter(r, "perPage")
	if queryParam != nil {
		perPageOverride, err := strconv.Atoi(queryParam[0])
		if err == nil && perPageOverride > 0 {
			perPage = perPageOverride
		}
	}

	// Fetch our query param "page"
	queryParam = mantis.GetQueryParameter(r, "page")
	if queryParam == nil || responseLength <= perPage {
		return response
	}

	// Convert our query "page?" value to an int
	page, err := strconv.Atoi(queryParam[0])
	if err != nil || page < 1 {
		return response
	}

	// set our min and max indexes
	indexMin := 0
	if page > 1 {
		indexMin = (page - 1) * perPage
	}
	indexMax := indexMin + perPage

	// handle a page greater than total available pages by setting to max
	if page > (responseLength / perPage) {
		indexMin = responseLength - perPage
		indexMax = responseLength
	}

	return response[indexMin:indexMax]
}

// GetStations This method returns all the stations; query by paging supported
func GetStations(w http.ResponseWriter, r *http.Request) {
	var stations Stations
	stations.getJSON()

	var response = make([]ShortStation, 0)
	for _, station := range stations.StationBeanList {
		response = append(response, ShortStation{
			StationName: station.StationName,
			Address: strings.TrimSpace(fmt.Sprintf("%s %s %s %s", station.Address1,
				station.Address2, station.City, station.PostalCode)),
			AvailableDocks: station.AvailableDocks,
			TotalDocks:     station.TotalDocks,
		})
	}

	response = page(r, response)
	HandleResponse(w, response, http.StatusOK)
}

// GetStationsInService
func GetStationsInService(w http.ResponseWriter, r *http.Request) {
	var stations Stations
	stations.getJSON()

	var response = make([]ShortStation, 0)
	for _, station := range stations.StationBeanList {
		if station.StatusKey == StatusOk {
			response = append(response, ShortStation{
				StationName: station.StationName,
				Address: strings.TrimSpace(fmt.Sprintf("%s %s %s %s", station.Address1,
					station.Address2, station.City, station.PostalCode)),
				AvailableDocks: station.AvailableDocks,
				TotalDocks:     station.TotalDocks,
			})
		}
	}

	response = page(r, response)
	HandleResponse(w, response, http.StatusOK)
}

// GetStationsNotInService
func GetStationsNotInService(w http.ResponseWriter, r *http.Request) {
	var stations Stations
	stations.getJSON()

	var response = make([]ShortStation, 0)
	for _, station := range stations.StationBeanList {
		if station.StatusKey == StatusNotOk {
			response = append(response, ShortStation{
				StationName: station.StationName,
				Address: strings.TrimSpace(fmt.Sprintf("%s %s %s %s", station.Address1,
					station.Address2, station.City, station.PostalCode)),
				AvailableDocks: station.AvailableDocks,
				TotalDocks:     station.TotalDocks,
			})
		}
	}

	response = page(r, response)
	HandleResponse(w, response, http.StatusOK)
}

// GetStationsMatchingString
func GetStationsMatchingString(w http.ResponseWriter, r *http.Request) {
	var stations Stations
	stations.getJSON()

	var response = make([]ShortStation, 0)

	// get our search string in /stations/:search, if empty return empty response
	searchString := mantis.GetUrlParameter(r, "search")
	if len(searchString) < 1 {
		HandleResponse(w, response, http.StatusOK)
		return
	}
	searchString = strings.TrimSpace(strings.ToLower(searchString))

	for _, station := range stations.StationBeanList {
		var key = strings.ToLower(fmt.Sprintf("%s %s %s", station.StationName, station.Address1, station.Address2))

		if strings.Contains(key, searchString) {
			response = append(response, ShortStation{
				StationName: station.StationName,
				Address: strings.TrimSpace(fmt.Sprintf("%s %s %s %s", station.Address1,
					station.Address2, station.City, station.PostalCode)),
				AvailableDocks: station.AvailableDocks,
				TotalDocks:     station.TotalDocks,
			})
		}
	}

	response = page(r, response)
	HandleResponse(w, response, http.StatusOK)
}

// GetIsBikeDockable
func GetIsBikeDockable(w http.ResponseWriter, r *http.Request) {
	var stations Stations
	stations.getJSON()

	var response BikesToReturn
	var errorOutputs = make(map[string]string)

	sid := mantis.GetUrlParameter(r, "stationId")
	stationId, err := strconv.Atoi(sid)
	if err != nil {
		mantis.HandleError("GetIsBikeDockable:stationId", err)
		errorOutputs["error"] = "Missing or invalid station id"
		HandleResponse(w, errorOutputs, http.StatusBadRequest)
		return
	}

	btr := mantis.GetUrlParameter(r, "bikesToReturn")
	bikesToReturn, err := strconv.Atoi(btr)
	if err != nil {
		mantis.HandleError("GetIsBikeDockable:bikesToReturn", err)
		errorOutputs["error"] = "Missing or invalid num bikes to return"
		HandleResponse(w, errorOutputs, http.StatusBadRequest)
		return
	}

	response.Message = "No docks available"
	response.Dockable = false
	status := http.StatusBadRequest

	for _, station := range stations.StationBeanList {
		if station.Id == stationId && station.AvailableDocks > 0 {
			if bikesToReturn-station.AvailableDocks > 0 {
				response.Dockable = false
				response.Message = fmt.Sprintf("Docks are available for %d docks, you are requesting return of %d bikes", station.AvailableDocks, bikesToReturn)
				status = http.StatusBadRequest
				break
			}

			if station.StatusKey == StatusNotOk {
				response.Dockable = false
				response.Message = "Docks are available, but station is out of service"
				status = http.StatusBadRequest
				break
			}
			response.Dockable = true
			response.Message = "Docks available"
			status = http.StatusOK
			break
		}
	}

	HandleResponse(w, response, status)
}
