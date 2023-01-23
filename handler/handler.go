package handler

import (
	"url-shortner/link"
)

type Handler struct {
	linkStore Link.Store
}

func NewHandler(linkStore Link.Store) *Handler {
	return &Handler{
		linkStore: linkStore,
	}
}
