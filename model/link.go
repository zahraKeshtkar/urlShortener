package model

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"url-shortner/log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Link struct {
	URL      string `json:"url"`
	ShortURL string `gorm:"-"`
	ID       uint64 `json:"-"`
}

const shortURLLength = 8

var (
	alphabets       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	alphabetsLength = 51
	urlRegex        *regexp.Regexp
)

func init() {
	urlRegex = regexp.MustCompile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
}

func (link *Link) Validate() bool {
	err := validation.ValidateStruct(link,
		validation.Field(&link.URL, validation.Match(urlRegex)),
		validation.Field(&link.ShortURL, validation.Length(shortURLLength, shortURLLength)))

	return err == nil
}

func (link *Link) MakeShortURL() {
	log.Debug("start to make shortURL")
	id := int(link.ID)
	chars := []rune(alphabets)
	var shortURL string
	for id > 0 {
		shortURL += string(chars[id%alphabetsLength])
		log.Debug(chars[id%alphabetsLength], "append to the short url")
		id = id / alphabetsLength
	}

	shortURL = link.ExpandURLLength(shortURL)
	link.ShortURL = shortURL

}

func (link *Link) ExpandURLLength(url string) string {
	var shortURL = ""
	var diff = shortURLLength - utf8.RuneCountInString(url)
	for i := 0; i < diff; i++ {
		shortURL += "Z"
	}

	shortURL += url
	log.Debug("append ", shortURL, "to the url", url)

	return shortURL
}

func (link *Link) ShortUrlToId() int {
	shortURL := strings.ReplaceAll(link.ShortURL, "Z", "")
	var id = 0
	for _, r := range shortURL {
		if int('a') <= int(r) && int(r) <= int('z') {
			id = id*alphabetsLength + int(r) - int('a')
		}

		if int('A') <= int(r) && int(r) <= int('Z') {
			id = id*alphabetsLength + int(r) - int('A') + 26
		}

	}

	return id
}
