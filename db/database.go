package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"url-shortner/log"
	"url-shortner/model"
)

type DataBase struct {
	gormDB *gorm.DB
}

func (db *DataBase) NewConnection(host string, retry int, retryTimeout time.Duration, user string, password string, database string, port string) {
	var err error
	counter := 0
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disabled database=%s", host, port, user, password, database)
	db.gormDB, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Fatalf("Cannot open database %s: %s", host, err)
	}

	sqlDB, err := db.gormDB.DB()
	if err != nil {
		log.Fatalf("Getting database error: %s", err)
	}

	counter = 0
	for time.Now(); true; <-time.NewTicker(retryTimeout).C {
		counter++
		err := sqlDB.Ping()
		if err == nil {
			break
		}

		log.Errorf("Cannot connect to database %s: %s", host, err)
		if counter >= retry {
			log.Fatalf("Cannot connect to database %s after %d retries: %s", host, counter, err)
		}

	}

	log.Debugf("Connected to postgres database: %s", host)
}

func (db *DataBase) CloseDataBase() {
	dbInstance, _ := db.gormDB.DB()
	err := dbInstance.Close()
	if err != nil {
		log.Fatalf("Cannot close database : %s", err)
	}

}

func (db *DataBase) CreateTable() {
	if !db.gormDB.Migrator().HasTable(model.Link{}) {
		err := db.gormDB.Migrator().CreateTable(&model.Link{})
		if err != nil {
			fmt.Println(err)
		}

	}

}

func (db *DataBase) GetLink(id int) model.Link {
	var link model.Link
	db.gormDB.Model(&link).First(&link, id)
	return link

}

func (db *DataBase) InsertLink(link *model.Link) {
	db.gormDB.Table("links").Create(&link)
}
