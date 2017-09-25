package repository

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"github.com/garyburd/redigo/redis"
)

var db *gorm.DB
var redisClient redis.Conn

func init()  {
	log.Printf("==repository init==")
	var err error
	db, err = gorm.Open("mysql", "cmsdemo:cmsdemo123@tcp(127.0.0.1:3306)/cms?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("[ERROR] mysql connection error: %v\n", err)
		db = nil
	}
	redisClient, err = redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Printf("[ERROR] redis connection error: %v\n", err)
		redisClient = nil
	}
}

func Close() {
	if db != nil {
		db.Close()
	}
	if redisClient != nil {
		redisClient.Close()
	}
}