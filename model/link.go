package model

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"url-shortner/log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	alphabets       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	alphabetsLength = 51
	shortURLLength  = 8
)

var urlRegex *regexp.Regexp

func init() {
	urlRegex = regexp.MustCompile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
}

type Link struct {
	ID       uint64 `json:"-" gorm:"primaryKey"`
	URL      string `json:"url" gorm:"not null"`
	ShortURL string `gorm:"-:all"`
}

func (link *Link) Validate() bool {
	err := validation.ValidateStruct(link,
		validation.Field(&link.URL, validation.Match(urlRegex)),
		validation.Field(&link.ShortURL, validation.Length(shortURLLength, shortURLLength)))

	return err == nil
}

func (link *Link) MakeShortURL() error {
	log.Debugf("start to make shortURL for long url %s", link.URL)
	id := int(link.ID)
	if id <= 0 {
		return errors.New("id is not valid")
	}

	chars := []rune(alphabets)
	var shortURL string
	for id > 0 {
		shortURL += string(chars[id%alphabetsLength])
		log.Debug(chars[id%alphabetsLength], "append to the short url")
		id = id / alphabetsLength
	}

	shortURL = link.expandURLLength(shortURL)
	link.ShortURL = shortURL

	return nil
}

func (link *Link) expandURLLength(url string) string {
	var shortURL = ""
	var diff = shortURLLength - utf8.RuneCountInString(url)
	for i := 0; i < diff; i++ {
		shortURL += "Z"
	}

	shortURL += url
	log.Debug("append ", shortURL, "to the url", url)

	return shortURL
}

func (link *Link) ShortURLToID() (int, error) {
	shortURL := strings.ReplaceAll(link.ShortURL, "Z", "")
	var id = 0
	for _, r := range shortURL {
		if int('a') <= int(r) && int(r) <= int('z') {
			id = id*alphabetsLength + int(r) - int('a')
		}

		if int('A') <= int(r) && int(r) < int('Z') {
			id = id*alphabetsLength + int(r) - int('A') + 26
		}

	}

	if id > 0 {
		return id, nil
	}

	return id, errors.New("id is not valid")
}
