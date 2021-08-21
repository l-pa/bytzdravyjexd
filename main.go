package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const WINNERS_URL = "https://www.mfsr.sk/components/mfsrweb/winners/data-ajax.jsp"

type WinnerJSON []struct {
    Code string `json:"kod"`
    Village string `json:"obec"`
	Amount string `json:"vyherna suma"`
	Name string `json:"meno"`
}

type ResponseJSON struct {
	Code string `json:"Code"`
	Village string `json:"Village"`
	Amount uint64 `json:"Amount"`
	Name string `json:"Name"`
}

func getLatestWinners(w http.ResponseWriter) []ResponseJSON  {
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}

	w.Header().Set("Content-Type", "application/json")


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
		var v, _ = strconv.ParseUint(s.Amount,10,32)
		arr = append(arr, ResponseJSON{Code: s.Code, Village: s.Village, Name: s.Name, Amount: v})
	}

	return arr
}


func getWinnersJSON(w http.ResponseWriter, req *http.Request) {

	winners := getLatestWinners(w)
	jData, err := json.Marshal(&winners)

	if err != nil {
		log.Fatal(err)
	}

	w.Write(jData)
}

func sumVillagesJSON(w http.ResponseWriter, req *http.Request) {

	winners := getLatestWinners(w)

	villages := make(map[string]int)

	for _, winner := range winners {
		if _, ok := villages[winner.Village]; ok {
			villages[winner.Village] += int(winner.Amount)
		} else {
			villages[winner.Village] = int(winner.Amount)
		}
	}

	jData, err := json.Marshal(&villages)

	if err != nil {
		log.Fatal(err)
	}

	w.Write(jData)
}

func sumNamesJSON(w http.ResponseWriter, req *http.Request) {

	winners := getLatestWinners(w)

	names := make(map[string]int)

	for _, winner := range winners {
		if _, ok := names[winner.Name]; ok {
			names[winner.Name] += 1
		} else {
			names[winner.Name] = 1
		}
	}

	jData, err := json.Marshal(&names)

	if err != nil {
		log.Fatal(err)
	}

	w.Write(jData)
}

func main() {
    http.HandleFunc("/winners", getWinnersJSON)
	http.HandleFunc("/villages", sumVillagesJSON)
	http.HandleFunc("/names", sumNamesJSON)

	http.ListenAndServe(":5000", nil)
}