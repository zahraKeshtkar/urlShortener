package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"url-shortner/log"
)

func NewConnection(host string,
	retry int,
	retryTimeout time.Duration,
	user string,
	password string,
	database string,
	port int) (*gorm.DB, error) {
	var err error
	var db *gorm.DB
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable database=%s", host, port, user, password, database)
	db, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Errorf("Cannot open database %s: %s", host, err)

		return db, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Errorf("Getting database error: %s", err)

		return db, err
	}

	counter := 0
	tickerChannel := time.NewTicker(retryTimeout)
	for ; true; <-tickerChannel.C {
		counter++
		err = sqlDB.Ping()
		if err == nil {
			tickerChannel.Stop()

			break
		}

		log.Errorf("Cannot connect to database %s: %s", host, err)
		if counter >= retry {
			log.Errorf("Cannot connect to database %s after %d retries: %s", host, counter, err)

			return db, err
		}

	}

	log.Infof("Connected to postgres database: %s", host)

	return db, nil
}

func Close(db *gorm.DB) error {
	dbInstance, err := db.DB()
	closeError := dbInstance.Close()
	if err != nil {
		log.Errorf("Cannot close database : %s", err)

		return err
	}

	if closeError != nil {
		log.Errorf("Cannot close database : %s", closeError)

		return closeError
	}

	return nil
}
