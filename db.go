package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"gopkg.in/guregu/null.v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB_FILE_NAME = "slobodajexd.db"

type winners struct {
	Code   string
	Amount int64
	City   string
	Name   string
}

type updates struct {
	Code          string
	Amount_before int64
	Date_Time     string
	Update_type   string
}

type updatesjoin struct {
	Code          string
	Amount        int64
	City          string
	Name          string
	Amount_before int64
	Date_Time     string
	Update_type   string
}

type cities struct {
	name   string
	region null.String
	lat    null.String
	long   null.String
}


type nominatimResponse []struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
	Icon        string   `json:"icon"`
}

func UpdateDbWinners() {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var dbWinners []winners
	var _ = db.Find(&dbWinners)

	latestWinners := getLatestWinners()

	var winnerTmp winners
	var cityTmp cities

	for _, v := range latestWinners {

		var citiesResult = db.First(&cityTmp, "name = ?", v.Village)

		if errors.Is(citiesResult.Error, gorm.ErrRecordNotFound) {

			var cityName = strings.Replace(v.Village, " ", "%20", -1)

			httpClient := http.Client{
				Timeout: time.Second * 10,
			}

			req, err := http.NewRequest(http.MethodGet, "https://nominatim.openstreetmap.org/search?q="+cityName+"&countrycodes=sk&format=json", nil)

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

			c := nominatimResponse{}

			jsonErr := json.Unmarshal(body, &c)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}

			if (len(c) > 0){
				var _ = db.Create(cities{name: v.Village, region: null.StringFromPtr(nil), lat: null.StringFrom(c[0].Lat), long: null.StringFrom(c[0].Lon)})
			} else {
				var _ = db.Create(cities{name: v.Village, region: null.StringFromPtr(nil), lat: null.StringFromPtr(nil),long: null.StringFromPtr(nil)})
			}

		}

		result := db.First(&winnerTmp, "code = ? AND amount = ?", v.Code, v.Amount)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			result := db.First(&winnerTmp, "code = ?", v.Code)

			var _ = db.Create(winners{Name: v.Name, Code: v.Code, Amount: v.Amount, City: v.Village})

			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				var _ = db.Create(updates{Code: v.Code, Amount_before: winnerTmp.Amount, Date_Time: time.Now().Format("2006-01-02 15:04:05"), Update_type: "UPDATE"})
			}
		}

	}

	fmt.Println("Updated ✅")
}

func GetDbWinners() []winners {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var dbWinners []winners

	var _ = db.Find(&dbWinners)

	return dbWinners
}

func GetDbVillage(name string) cities {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var dbVillages []cities

	var _ = db.First(&dbVillages).Where("name = ?", name)

	return dbVillages[0]
}

func GetDbUpdates() []updates {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var dbUpdates []updates

	var _ = db.Where("update_type = ?", "UPDATE").Find(&dbUpdates)

	return dbUpdates
}

func GetDbInserts() []updates {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var dbUpdates []updates

	var _ = db.Where("update_type = ?", "INSERT").Find(&dbUpdates)

	return dbUpdates
}

func GetDb24Update() []updatesjoin {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	var dbUpdates []updatesjoin

	var _ = db.Raw("select * from (select * from updates u left join winners w using(code) union all select * from winners w left join updates u using (code) where w.code is null) where date_time > datetime('now','-1 day') and amount != amount_before").Find(&dbUpdates)

	// var _ = db.Where("datetime > datetime('now','-1 day')").Find(&dbUpdates)

	return dbUpdates
}
