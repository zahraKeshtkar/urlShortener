package Link

import (
	"url-shortner/model"
)

type Store interface {
	CreateLinkTable() error
	GetLink(id int) model.Link
	InsertLink(link *model.Link)
}
