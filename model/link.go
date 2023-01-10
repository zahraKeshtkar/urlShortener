package model

import (
	"crypto/md5"
	"encoding/base64"
	"net/http"
	"net/url"
)

type Link struct {
	Url  string
	Hash string
}

func NewLink(url string) *Link {
	l := new(Link)
	l.Url = url
	l.Hash = MakeShortUrl(url)
	return l
}

func IsUrlValid(longUrl string) bool {
	_, err := url.ParseRequestURI(longUrl)
	return err == nil
}

func IsLinkExits(longUrl string) bool {
	_, err := http.Head(longUrl)
	return err == nil
}

func MakeShortUrl(longUrl string) string {
	md := md5.Sum([]byte(longUrl))
	hash := base64.StdEncoding.EncodeToString(md[:])
	return hash[:6]
}
