package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var cacheTimeSeconds int
var sheet_url string

type Sheet struct {
	Range          string     `json:"range"`
	MajorDimension string     `json: "majorDimension"`
	Values         [][]string `json: "values"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func GetSheet() (sheet *Sheet, err error) {

	var resp *http.Response
	resp, err = http.Get(sheet_url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	sheet = &Sheet{}
	//resp.Body
	err = json.NewDecoder(resp.Body).Decode(sheet)
	if err != nil {
		return
	}

	return
}

func GetSheetData() (data []map[string]interface{}, err error) {

	var sheet *Sheet
	sheet, err = GetSheet()
	if err != nil {
		return
	}

	for i := 1; i < len(sheet.Values); i++ {
		obj := make(map[string]interface{})
		for col := 0; col < len(sheet.Values[i]); col++ {
			obj[sheet.Values[0][col]] = sheet.Values[i][col]
		}
		data = append(data, obj)
	}
	return

	var data_byte []byte
	data_byte, err = json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Println(string(data_byte))
	return
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func GetJsonFromFile(cache_file string) (data []map[string]interface{}, err error) {
	var jsonFile *os.File
	jsonFile, err = os.Open(cache_file)
	if err != nil {
		return
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}

	data = []map[string]interface{}{}
	err = json.Unmarshal([]byte(byteValue), &data)
	if err != nil {
		return
	}
	return
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func GetDataCached() (data []map[string]interface{}, err error) {
	cache_file := "cache.json"

	fInfo, ferr := os.Stat(cache_file)
	if ferr == nil {

		stat_t := fInfo.Sys().(*syscall.Stat_t)

		fileTimeCreated := timespecToTime(stat_t.Ctim) // Linux: Ctim ,  OSX: Ctimespec

		ageSeconds := int(time.Now().Sub(fileTimeCreated).Seconds())
		fmt.Printf("age of cache in seconds: %d\n", ageSeconds)
		if ageSeconds < cacheTimeSeconds {
			fmt.Println("return cached file")
			data, err = GetJsonFromFile(cache_file)
			return
		}
		fmt.Println("cached file too old")
	}

	// file does not exists or is too old

	data, err = GetSheetData()
	if err != nil {
		return
	}

	// caching

	var data_byte []byte
	data_byte, err = json.Marshal(data)
	if err != nil {
		return
	}

	err = os.WriteFile(cache_file, data_byte, 0644)
	if err != nil {
		return
	}

	return
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	data, err := GetDataCached()
	if err != nil {

		er := ErrorResponse{Error: err.Error()}
		response_json, _ := json.Marshal(er)
		http.Error(w, string(response_json), http.StatusInternalServerError)

		return
	}

	response_json, err := json.Marshal(data)
	if err != nil {

		er := ErrorResponse{Error: err.Error()}
		response_json, _ := json.Marshal(er)
		http.Error(w, string(response_json), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(response_json))

}

func main() {

	addr := "127.0.0.1:8000"

	var err error
	if os.Getenv("CACHE_TIME_SECONDS") == "" {
		cacheTimeSeconds = 60
	} else {
		cacheTimeSeconds, err = strconv.Atoi(os.Getenv("CACHE_TIME_SECONDS"))
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	sheet_url = os.Getenv("GOOGLE_SHEET_URL")
	if sheet_url == "" {
		log.Fatal("GOOGLE_SHEET_URL not defined")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	//http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Listening on %s ...\n", addr)
	log.Fatal(srv.ListenAndServe())

}
