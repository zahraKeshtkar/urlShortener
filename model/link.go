package model

import (
	"crypto/md5"
	"encoding/base64"
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
