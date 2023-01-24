package repository

import (
	"gorm.io/gorm"

	"url-shortner/log"
	"url-shortner/model"
)

type LinkStore struct {
	DB *gorm.DB
}

func (linkStore *LinkStore) CreateTable() error {
	if !linkStore.DB.Migrator().HasTable(model.Link{}) {
		err := linkStore.DB.Migrator().CreateTable(&model.Link{})
		if err != nil {
			log.Errorf("creating LinkStore fail %s", err)

			return err
		}
	}

	return nil
}

func (linkStore *LinkStore) Get(id int) model.Link {
	var link model.Link
	linkStore.DB.Model(&link).First(&link, id)

	return link
}

func (linkStore *LinkStore) Insert(link *model.Link) error {
	r := linkStore.DB.Table("links").Create(&link)
	return r.Error
}
