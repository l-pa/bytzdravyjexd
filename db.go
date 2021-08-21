package main

import (
	"errors"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB_FILE_NAME = "slobodajexd.db"

type winners struct {
	Code   string
	Amount int32
	City   string
	Name   string
}

type updates struct {
	Code         string
	Amount_before int32
	DateTime     string
	Update_type string
}

type times struct {
	date string
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
		
		result := db.First(&winnerTmp, "code = ?", v.Code)

		if (errors.Is(result.Error, gorm.ErrRecordNotFound)) {
			var _ = db.Create(winners{Name: v.Name, Code: v.Code, Amount: int32(v.Amount), City: v.Village})
		}

		if (v.Amount != uint64(winnerTmp.Amount)) {
			db.Model(&winnerTmp).Where("code = ?", winnerTmp.Code).Update("Amount", v.Amount)
		}
		
	}
	
	fmt.Println("Updated 💥")
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

func GetDbGo() times {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
	  panic("failed to connect database")
	}

	var dbUpdates []times  

	var _ = db.Order("datetime(datetime) desc").First(&dbUpdates)

	return dbUpdates[0]
}

func GetDb24Update() []updates {
	db, err := gorm.Open(sqlite.Open(DB_FILE_NAME), &gorm.Config{})
	if err != nil {
	  panic("failed to connect database")
	}
	var dbUpdates []updates  

	var _ = db.Where("datetime > datetime('now','-1 day')").Find(&dbUpdates)

	return dbUpdates
}