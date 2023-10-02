package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"jameesjohn.com/uploader-api/config"
	"jameesjohn.com/uploader-api/models"
	"log"
)

var Db *gorm.DB

func ConnectDatabase() {
	if !config.IsLoaded() {
		log.Fatal("Config not loaded")
	}

	//	Attempt to connect to mysql database
	dsn := fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		config.Environment.DbUser,
		config.Environment.DbPassword,
		config.Environment.DbHost,
		config.Environment.DbPort,
		config.Environment.DbName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("unable to connect to database::", err)
	}

	Db = db
}

func Migrate() {
	if Db == nil {
		log.Fatal("Database not setup")
	}

	err := Db.AutoMigrate(&models.Video{}, &models.UploadStatistic{})
	if err != nil {
		log.Fatal("unable to migrate:", err)
	}
}
