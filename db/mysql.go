package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
	db, err := gorm.Open(mysql.Open(dsn))
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
	if err != nil {
		log.Errorf("Cannot close database : %s", err)

		return err
	}

	err = dbInstance.Close()
	if err != nil {
		log.Errorf("Cannot close database : %s", err)

		return err
	}

	return nil
}
