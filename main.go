package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

const WINNERS_URL = "https://www.mfsr.sk/components/mfsrweb/winners/data-ajax.jsp"

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
	Amount  uint64 `json:"Amount"`
	Name    string `json:"Name"`
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
		var v, _ = strconv.ParseUint(s.Amount, 10, 32)
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

	winners := GetDbWinners()

	villages := make(map[string]int)

	for _, winner := range winners {
		if _, ok := villages[winner.City]; ok {
			villages[winner.City] += int(winner.Amount)
		} else {
			villages[winner.City] = int(winner.Amount)
		}
	}

	if len(villages) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: villages})
		w.Write(data)
	}
}

func SumNamesJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDbWinners()

	names := make(map[string]int)

	for _, winner := range winners {
		if _, ok := names[winner.Name]; ok {
			names[winner.Name] += 1
		} else {
			names[winner.Name] = 1
		}
	}

	if len(names) == 0 {
		data, _ := json.Marshal(StatusJSON{Status: -1, Text: "No data"})
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: names})
		w.Write(data)
	}
}

func GetDbLastUpdateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDbGo()

	data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
	w.Write(data)
}

func GetDbLast24UpdateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	winners := GetDb24Update()

	data, _ := json.Marshal(StatusJSON{Status: 0, Text: "Ok", Response: winners})
	w.Write(data)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Piticko 🍷 a livance 🥞")
}

func main() {

	c := cron.New()
	c.AddFunc("0 0 * * * *", UpdateDbWinners )

	http.HandleFunc("/", NotFoundHandler)
	http.HandleFunc("/msfs", GetWinnersJSON)
	http.HandleFunc("/villages", SumVillagesJSON)
	http.HandleFunc("/names", SumNamesJSON)
	http.HandleFunc("/db", GetDbWinnersJSON)
	http.HandleFunc("/updates", GetDbUpdatesJSON)
	http.HandleFunc("/inserts", GetDbInsertsJSON)
	http.HandleFunc("/24", GetDbLast24UpdateJSON)

	// http.HandleFunc("/lastupdate", GetDbLastUpdateJSON)

	// UpdateDbWinners()

	c.Start()

	fmt.Println("Server ✅ :5000")
	http.ListenAndServe(":5000", nil)
}
