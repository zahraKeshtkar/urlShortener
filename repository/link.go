package repository

import (
	"gorm.io/gorm"

	"url-shortner/log"
	"url-shortner/model"
)

type Link struct {
	DB *gorm.DB
}

func (linkStore *Link) CreateTable() error {
	if !linkStore.DB.Migrator().HasTable(model.Link{}) {
		err := linkStore.DB.Migrator().CreateTable(&model.Link{})
		if err != nil {
			log.Errorf("creating Link fail %s", err)

			return err
		}
	}

	return nil
}

func (linkStore *Link) Get(id int) model.Link {
	var link model.Link
	linkStore.DB.Model(&link).First(&link, id)

	return link
}

func (linkStore *Link) Insert(link *model.Link) error {
	r := linkStore.DB.Table("links").Create(&link)
	return r.Error
}
