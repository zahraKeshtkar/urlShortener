package model

import (
	"crypto/md5"
	"encoding/base64"
	"net/http"
	"net/url"
)

type Link struct {
	url  string
	hash string
}

func NewLink(url string) {
	l := new(Link)
	l.url = url
	md5 := md5.Sum([]byte(url))
	hash := base64.StdEncoding.EncodeToString(md5[:])
	l.hash = hash[:6]
}
func IsUrlValid(longUrl string) bool {
	_, err := url.ParseRequestURI(longUrl)
	return err == nil
}
func IsLinkExits(longUrl string) bool {
	_, err := http.Head(longUrl)
	return err == nil
}
