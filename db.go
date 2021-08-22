package main

import (
	"errors"
	"fmt"
	"time"

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
	Code         string
	Amount_before int64
	Date_Time     string
	Update_type string
}

type updatesjoin struct {
	Code         string
	Amount int64
	City   string
	Name   string
	Amount_before int64
	Date_Time     string
	Update_type string
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

	for _, v := range latestWinners {
		
		result := db.First(&winnerTmp, "code = ? AND amount = ?", v.Code, v.Amount)

		if (errors.Is(result.Error, gorm.ErrRecordNotFound)) {

			result := db.First(&winnerTmp, "code = ?", v.Code)
			
			var _ = db.Create(winners{Name: v.Name, Code: v.Code, Amount: v.Amount, City: v.Village})
			
			if (!errors.Is(result.Error, gorm.ErrRecordNotFound)) {
				var _ = db.Create(updates{Code: v.Code, Amount_before: winnerTmp.Amount, Date_Time: time.Now().Format("2006-01-02 15:04:05"),Update_type: "UPDATE" })
			} 
		}

		// if (v.Amount != winnerTmp.Amount) {
		// 	db.Model(&winnerTmp).Where("code = ?", winnerTmp.Code).Update("Amount", v.Amount)
		// 	fmt.Println(v.Code + " updated 💨")
		// }
		
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