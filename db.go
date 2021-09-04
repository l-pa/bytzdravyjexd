package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

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
	Name   string
	Region sql.NullString
	Lat    sql.NullString
	Long   sql.NullString
}

type winnersCitiesJoin struct {
	Name string
	Lat sql.NullString
	Long sql.NullString
	Total int64
}

type names struct {
	Name string
	Amount int64
	Count int64
}

type geoJson struct {
	Type string `json:"type"`
	Features []features `json:"features"`
}

type features struct {
	Type string `json:"type"`
	Geometry geometry `json:"geometry"`
	Properties geoProperties `json:"properties"`
}

type geometry struct {
	Type string `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

type geoProperties struct {
	Amount int64 `json:"amount"`
	Village string `json:"village"`
}


type nominatimResponse []struct {
	PlaceID     int64    `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int64      `json:"osm_id"`
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
	var dbWinners []winners
	var _ = Db.Find(&dbWinners)

	latestWinners := getLatestWinners()

	var winnerTmp winners

	// var cityTmp cities

	for _, v := range latestWinners {

		// var citiesResult = Db.First(&cityTmp, "name = ?", v.Village)

		// if errors.Is(citiesResult.Error, gorm.ErrRecordNotFound) {

		// 	var cityName = strings.Replace(v.Village, " ", "%20", -1)

		// 	httpClient := http.Client{
		// 		Timeout: time.Second * 10,
		// 	}

		// 	req, err := http.NewRequest(http.MethodGet, "https://nominatim.openstreetmap.org/search?q="+cityName+"&countrycodes=sk&format=json", nil)

		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}

		// 	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1")

		// 	res, getErr := httpClient.Do(req)

		// 	if getErr != nil {
		// 		log.Fatal(getErr)
		// 	}

		// 	if res.Body != nil {
		// 		defer res.Body.Close()
		// 	}

		// 	body, readErr := ioutil.ReadAll(res.Body)

		// 	if readErr != nil {
		// 		log.Fatal(readErr)
		// 	}

		// 	c := nominatimResponse{}

		// 	jsonErr := json.Unmarshal(body, &c)
		// 	if jsonErr != nil {
		// 		log.Fatal(jsonErr)
		// 	}

		// 	if len(c) > 0 {
		// 		var _ = Db.Create(cities{Name: v.Village, Region: sql.NullString{Valid: false}, Lat: sql.NullString{Valid: true, String: c[0].Lat}, Long: sql.NullString{Valid: true, String: c[0].Lon}})
		// 	} else {
		// 		var _ = Db.Create(cities{Name: v.Village, Region: sql.NullString{Valid: false}, Lat: sql.NullString{Valid: false}, Long: sql.NullString{Valid: false}})
		// 	}

		// }

		result := Db.First(&winnerTmp, "code = ? AND amount = ?", v.Code, v.Amount)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			result := Db.First(&winnerTmp, "code = ?", v.Code)

			var _ = Db.Create(winners{Name: v.Name, Code: v.Code, Amount: v.Amount, City: v.Village})

			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				var _ = Db.Create(updates{Code: v.Code, Amount_before: winnerTmp.Amount, Date_Time: time.Now().Format("2006-01-02 15:04:05"), Update_type: "UPDATE"})
			}
		}

	}

	fmt.Println("Updated ✅")
}

func GetDbWinners() []winners {
	var dbWinners []winners
	var _ = Db.Find(&dbWinners)
	return dbWinners
}

func GetDbVillageJoinWinners() []winnersCitiesJoin {
	var dbVillages []winnersCitiesJoin	
	Db.Select("cities.name", "cities.lat", "cities.long", "sum(winners.amount) as total").Table("winners").Joins("LEFT JOIN cities ON cities.name = winners.city").Group("cities.name").Order("total desc").Find(&dbVillages)
	return dbVillages
}

func GetDbNames() []names {
	var dbNames []names	
	Db.Select("name", "sum(amount) as amount", "count(name) as count").Table("winners").Group("name").Order("count desc").Find(&dbNames)
	return dbNames
}


func GetDbUpdates() []updates {
	var dbUpdates []updates
	var _ = Db.Where("update_type = ?", "UPDATE").Find(&dbUpdates)
	return dbUpdates
}

func GetDbInserts() []updates {
	var dbUpdates []updates

	var _ = Db.Where("update_type = ?", "INSERT").Find(&dbUpdates)

	return dbUpdates
}

func GetDb24Update() []updatesjoin {
	var dbUpdates []updatesjoin

	var _ = Db.Raw("select * from (select * from updates u left join winners w using(code) union all select * from winners w left join updates u using (code) where w.code is null) where date_time > datetime('now','-1 day') and amount != amount_before").Find(&dbUpdates)

	// var _ = db.Where("datetime > datetime('now','-1 day')").Find(&dbUpdates)

	return dbUpdates
}

func GetGeoJson() geoJson{
	dbVillages := GetDbVillageJoinWinners()
	var resGeoJson geoJson
	resGeoJson.Type = "FeatureCollection"

	var geoJsonFeatures []features

	for _, v := range dbVillages {
		lat, errLat := strconv.ParseFloat(v.Lat.String, 64)
		long, errLong :=strconv.ParseFloat(v.Long.String, 64)

		if errLat == nil && errLong == nil {
			geoJsonFeatures = append(geoJsonFeatures, features{Type: "Feature", Geometry: geometry{Type: "Point", Coordinates: [2]float64{long, lat}}, Properties: geoProperties{Amount: v.Total, Village: v.Name}})
		}

	}
	resGeoJson.Features = geoJsonFeatures

	
	return resGeoJson
}
