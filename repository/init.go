package repository

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

var db *gorm.DB

func init()  {
	log.Printf("==repository init==")
	var err error
	db, err = gorm.Open("mysql", "cmsdemo:cmsdemo123@tcp(127.0.0.1:3306)/cms?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("[ERROR] mysql connection error: %v\n", err)
		db = nil
	}
}

func Close() {
	if db != nil {
		db.Close()
	}
}