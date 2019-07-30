package main

import (
	"encoding/json"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func executeRequestViaRecorder(req *http.Request) *httptest.ResponseRecorder {
	httpTestRecorder := httptest.NewRecorder()

	App = Application{
		Name:    os.Getenv("APP_NAME"),
		ID:      os.Getenv("APP_ID"),
		Version: os.Getenv("APP_VERSION"),
		Log:     os.Getenv("LOG_LOCATION"),
		Server: Server{
			Address:      os.Getenv("SRV_ADDRESS"),
			Port:         os.Getenv("SRV_PORT"),
			WriteTimeout: time.Duration(10),
			ReadTimeout:  time.Duration(10),
			MemCacheTime: time.Duration(10),
		},
		Runtime: time.Now().UTC().Format(time.RFC3339),
	}
	App.Router.Load()
	App.Cache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	Logger.NewLog(App.Log)

	App.Router.router.ServeHTTP(httpTestRecorder, req)

	mux.NewRouter().ServeHTTP(httpTestRecorder, req)
	return httpTestRecorder
}

// remove404 for some reason it appends "404 page not found" to end of JSON response during testing
func remove404(response string) string {
	return strings.Replace(response, "404 page not found", "", -1)
}

func checkResponseCodeAndUnmarshalJSON(t *testing.T, expectedCode int, actualCode int, response string, doUnmarshal bool) []ShortStation {
	if expectedCode != actualCode {
		t.Errorf("Expected %d Got %d", expectedCode, actualCode)
	}
	response = remove404(response)

	station := make([]ShortStation, 0)
	if doUnmarshal {
		err := json.Unmarshal([]byte(response), &station)
		if err != nil {
			t.Errorf("JSON Unmarshal failed: %s", err.Error())
		}
	}

	return station
}

func TestStations(t *testing.T) {
	req, _ := http.NewRequest("GET", "/stations", nil)
	response := executeRequestViaRecorder(req)

	checkResponseCodeAndUnmarshalJSON(t, http.StatusOK, response.Code, response.Body.String(), true)
	body := response.Body.String()
	if len(body) == 0 || body == "[]" || body == "{}" {
		t.Errorf("Empty response")
	}
}

func TestStationsPaged(t *testing.T) {
	req, _ := http.NewRequest("GET", "/stations?page=3", nil)
	response := executeRequestViaRecorder(req)

	resp := checkResponseCodeAndUnmarshalJSON(t, http.StatusOK, response.Code, response.Body.String(), true)
	if len(resp) > 20 {
		t.Errorf("Received more values than expected")
	}
}

func TestStationsPageAndCount(t *testing.T) {
	for i := 1; i < 10; i++ {
		uri := fmt.Sprintf("/stations?page=1&perPage=%d", i)
		req, _ := http.NewRequest("GET", uri, nil)
		response := executeRequestViaRecorder(req)

		resp := checkResponseCodeAndUnmarshalJSON(t, http.StatusOK, response.Code, response.Body.String(), true)
		if len(resp) != i {
			t.Errorf("Received more (%d) values than expected (%d)", i, len(resp))
		}
	}
}

func TestStationsInService(t *testing.T) {
	req, _ := http.NewRequest("GET", "/stations/in-service", nil)
	response := executeRequestViaRecorder(req)

	resp := checkResponseCodeAndUnmarshalJSON(t, http.StatusOK, response.Code, response.Body.String(), true)
	if len(resp) == 0 {
		t.Errorf("Received no stations in service")
	}
}

func TestStationsNotInService(t *testing.T) {
	req, _ := http.NewRequest("GET", "/stations/not-in-service", nil)
	response := executeRequestViaRecorder(req)

	resp := checkResponseCodeAndUnmarshalJSON(t, http.StatusOK, response.Code, response.Body.String(), true)
	if len(resp) == 0 {
		t.Errorf("Received no stations out of service")
	}
}

func TestInvalidPage(t *testing.T) {
	req, _ := http.NewRequest("GET", "/stations/not-in-service?page=invalidpage", nil)
	response := executeRequestViaRecorder(req)

	resp := checkResponseCodeAndUnmarshalJSON(t, http.StatusOK, response.Code, response.Body.String(), true)
	if len(resp) == 0{
		t.Errorf("Received no stations out of service for invalid page")
	}
}

func TestInvalidStationIdForDockable(t *testing.T) {
	req, _ := http.NewRequest("GET", "/dockable/100000000000/100000000", nil)
	response := executeRequestViaRecorder(req)

	checkResponseCodeAndUnmarshalJSON(t, http.StatusBadRequest, response.Code, response.Body.String(), false)

	var station BikesToReturn
	err := json.Unmarshal([]byte(remove404(response.Body.String())), &station)
	if err != nil {
		t.Errorf("JSON Unmarshal failed: %s", err.Error())
	}
	if station.Message != "No docks available" {
		t.Errorf("Received more than zero for invalid dockable return")
	}
}
