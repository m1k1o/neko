package database

import (
	"demodesk/neko/internal/session/database/dummy"
	"demodesk/neko/internal/session/database/object"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
)

func New(config *config.Session) types.MembersDatabase {
	// TODO: Load from config.
	adapter := "object"

	switch adapter {
	case "object":
		return object.New()
	case "dummy":
		return dummy.New()
	}

	return dummy.New()
}
