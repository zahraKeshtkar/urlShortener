package store

import (
	"gorm.io/gorm"

	"url-shortner/log"
	"url-shortner/model"
)

type LinkStore struct {
	db *gorm.DB
}

func NewLinkStore(db *gorm.DB) *LinkStore {
	return &LinkStore{
		db: db,
	}
}

func (linkStore *LinkStore) CreateLinkTable() error {
	if !linkStore.db.Migrator().HasTable(model.Link{}) {
		err := linkStore.db.Migrator().CreateTable(&model.Link{})
		if err != nil {
			log.Errorf("creating LinkStore fail %s", err)

			return err
		}
	}

	return nil
}

func (linkStore *LinkStore) GetLink(id int) model.Link {
	var link model.Link
	linkStore.db.Model(&link).First(&link, id)

	return link
}

func (linkStore *LinkStore) InsertLink(link *model.Link) {
	linkStore.db.Table("links").Create(&link)
}
