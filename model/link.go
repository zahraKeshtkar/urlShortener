package model

import (
	"regexp"
	"unicode/utf8"

	"url-shortner/log"
)

type Link struct {
	URL      string
	ShortURL string
}

const ShortURLLength = 8

func NewLink(id int, URL string) *Link {
	l := new(Link)
	l.URL = URL
	l.ShortURL = MakeShortURL(id)

	return l
}

func IsURLValid(longURL string) bool {
	r, _ := regexp.Compile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

	return r.MatchString(longURL)
}

func MakeShortURL(id int) string {
	log.Trace("start to make shortURL")
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	chars := []rune(str)
	var shortURL string
	for id > 0 {
		shortURL += string(chars[id%51])
		log.Trace(chars[id%51], "append to the short url")
		id = id / 51
	}

	shortURL = ExpandURLLength(shortURL)

	return shortURL
}

func ExpandURLLength(url string) string {
	var shortURL = ""
	var diff = ShortURLLength - utf8.RuneCountInString(url)
	for i := 0; i < diff; i++ {
		shortURL += "Z"
	}

	shortURL += url
	log.Trace("append ", shortURL, "to the url", url)

	return shortURL
}

func FindShortURL(shortURL string, db map[string]string) (string, bool) {
	if utf8.RuneCountInString(shortURL) == ShortURLLength {
		longURL, ok := db[shortURL]

		return longURL, ok
	}
	return "", false
}
