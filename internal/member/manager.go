package member

import (
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/member/dummy"
	"demodesk/neko/internal/member/file"
	"demodesk/neko/internal/member/object"
	"demodesk/neko/internal/types"
)

func New(config *config.Member) types.MemberManager {
	switch config.Provider {
	case "file":
		return file.New(config.FilePath)
	case "object":
		return object.New()
	case "dummy":
		return dummy.New()
	}

	return dummy.New()
}
