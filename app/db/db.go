package db

import (
	"alert-map-service/app/models"
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dbType := os.Getenv("DB_TYPE")
	log.Print(dbType)
	switch dbType {
	case "mysql":
		sqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
		db, err = gorm.Open(mysql.Open(sqlInfo), &gorm.Config{})
	case "postgres":
		sqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
		db, err = gorm.Open(postgres.Open(sqlInfo), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	if err != nil {
		return nil, err
	}
	DB = db
	DB.AutoMigrate(&models.Client{})
	return db, nil
}
