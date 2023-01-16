package model

import (
	"math/rand"
	"regexp"
	"time"
	"unicode/utf8"

	"url-shortner/log"
)

type Link struct {
	URL      string `json:"url"`
	ShortURL string
}

const ShortURLLength = 8
const Retry = 3

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var re *regexp.Regexp

func init() {
	re, _ = regexp.Compile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
}

func (link *Link) IsURLValid() bool {
	return re.MatchString(link.URL)
}

func (link *Link) MakeShortURL(db map[string]string) bool {
	log.Trace("start to make shortURL")
	shortURL := make([]rune, ShortURLLength)
	rand.Seed(time.Now().UnixNano())
	var counter = 0
	for counter < Retry {
		for i := range shortURL {
			shortURL[i] = letters[rand.Intn(len(letters))]
		}
		_, ok := db[string(shortURL)]
		if !ok {
			link.ShortURL = string(shortURL)
			return true

		}

		counter += 1
	}

	return false
}

func FindShortURL(shortURL string, db map[string]string) (string, bool) {
	if utf8.RuneCountInString(shortURL) == ShortURLLength {
		longURL, ok := db[shortURL]

		return longURL, ok
	}
	return "", false
}
