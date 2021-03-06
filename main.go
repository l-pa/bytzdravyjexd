package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const WINNERS_URL = "https://www.mfsr.sk/components/mfsrweb/winners/data-ajax.jsp"

var lastCronUpdate time.Time

const DB_FILE_NAME = "slobodajexd.db"
var Db, Err = gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})

type StatusJSON struct {
	Status   int32
	Text     string
	Response interface{}
}

type WinnerJSON []struct {
	Code    string `json:"kod"`
	Village string `json:"obec"`
	Amount  string `json:"vyherna suma"`
	Name    string `json:"meno"`
}

type ResponseJSON struct {
	Code    string `json:"Code"`
	Village string `json:"Village"`
	Amount  int64 `json:"Amount"`
	Name    string `json:"Name"`
}

type VillagesJSON struct {
	Village string
	TotalAmount  int
	Lat string
	Lon string
}

func getLatestWinners() []ResponseJSON {
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, WINNERS_URL, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1")

	res, getErr := httpClient.Do(req)

	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		log.Fatal(readErr)
	}

	winners := WinnerJSON{}

	jsonErr := json.Unmarshal(body, &winners)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	var arr []ResponseJSON

	for _, s := range winners {
		var v, _ = strconv.ParseInt(s.Amount, 10, 32)
		arr = append(arr, ResponseJSON{Code: s.Code, Village: s.Village, Name: s.Name, Amount: v})
	}

	return arr
}

func GetWinnersJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := getLatestWinners()
	if len(winners) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
		w.Write(data)
	}
}

func GetDbWinnersJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDbWinners()

	if len(winners) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
		w.Write(data)
	}
}

func GetDbUpdatesJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDbUpdates()
	// jData, err := json.Marshal(&winners)

	/*
		if err != nil {
			log.Fatal(err)
		}
	*/

	data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
	if len(winners) > 0 {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
		w.Write(data)
		return
	}
	w.Write(data)
}

func GetDbInsertsJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDbInserts()

	if len(winners) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
		w.Write(data)
	}
}

func SumVillagesJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	count := GetDbVillageJoinWinners()


	if len(count) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: count})
		w.Write(data)
	}
}

func SumNamesJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDbNames()

	if len(winners) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
		w.Write(data)
	}
}

func GetDbLastUpdateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// winners := GetDbGo()

	data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: lastCronUpdate})
	w.Write(data)
}

func GetGeoJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, _ := json.Marshal(GetGeoJson())
	w.Write(data)
}

func GetDbLast24UpdateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: GetDb24Update()})
	w.Write(data)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Piticko ???? a livance ????")
}

func main() {
	s := gocron.NewScheduler(time.Local)
	s.Every(1).Day().At("09:30").Do(func ()  {
		UpdateDbWinners()
		lastCronUpdate = time.Now()
		fmt.Println("Cron ???")
		
	})

	mux := http.NewServeMux()


	mux.HandleFunc("/", NotFoundHandler)
	mux.HandleFunc("/msfs", GetWinnersJSON)
	mux.HandleFunc("/villages", SumVillagesJSON)
	mux.HandleFunc("/names", SumNamesJSON)
	mux.HandleFunc("/db", GetDbWinnersJSON)
	mux.HandleFunc("/updates", GetDbUpdatesJSON)
	mux.HandleFunc("/inserts", GetDbInsertsJSON)
	mux.HandleFunc("/24", GetDbLast24UpdateJSON)
	mux.HandleFunc("/geojson", GetGeoJSON)

	mux.HandleFunc("/cron", GetDbLastUpdateJSON)


	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:1234", "http://localhost"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	handler := c.Handler(mux)

	UpdateDbWinners()
	
	lastCronUpdate = time.Now()
	s.StartAsync()
	fmt.Println("Server http://localhost:5000 ???")
	http.ListenAndServe(":5000", handler)
}
