package database

import (
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/session/database/dummy"
	"demodesk/neko/internal/session/database/file"
	"demodesk/neko/internal/session/database/object"
	"demodesk/neko/internal/types"
)

func New(config *config.Session) types.MembersDatabase {
	switch config.DatabaseAdapter {
	case "file":
		return file.New(config.FilePath)
	case "object":
		return object.New()
	case "dummy":
		return dummy.New()
	}

	return dummy.New()
}
