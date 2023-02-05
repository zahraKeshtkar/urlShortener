package model

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"url-shortner/log"
)

const (
	alphabets       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	alphabetsLength = 51
	shortURLLength  = 8
)

var urlRegex *regexp.Regexp
var shortURLRegex *regexp.Regexp

func init() {
	urlRegex = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[.\!\/\\w]*))?)`)
	shortURLRegex = regexp.MustCompile(`^[a-zA-Z]{7}[a-yA-Y]$`)
}

type Link struct {
	ID       int    `json:"-" gorm:"column:id;primaryKey"`
	URL      string `json:"url" gorm:"column:url"`
	ShortURL string `gorm:"-:all"`
}

func (link *Link) Validate() bool {
	return urlRegex.MatchString(link.URL) || shortURLRegex.MatchString(link.ShortURL)
}

func (link *Link) MakeShortURL() error {
	log.Debugf("Start to make shortURL for long url %s", link.URL)
	id := link.ID
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
	log.Debug("Append ", shortURL, "to the url", url)

	return shortURL
}

func (link *Link) ShortURLToID() (int, error) {
	if !link.Validate() {
		return 0, errors.New("shortURL is not valid")
	}

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
