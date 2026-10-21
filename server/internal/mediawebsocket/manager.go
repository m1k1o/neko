package mediawebsocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/internal/media"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type Manager struct {
	logger zerolog.Logger
	media  *media.Manager
}

func New(manager *media.Manager) *Manager {
	return &Manager{
		logger: log.With().Str("module", "media-websocket").Logger(),
		media:  manager,
	}
}

func (manager *Manager) Upgrade(checkOrigin types.CheckOrigin) types.RouterHandler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 64 * 1024,
		CheckOrigin:     checkOrigin,
	}

	return func(w http.ResponseWriter, request *http.Request) error {
		ticket, session, err := manager.media.ConsumeTicket(request.URL.Query().Get("ticket"), media.TransportWebSocket)
		if err != nil {
			return utils.HttpUnauthorized().WithInternalErr(err)
		}

		connection, err := upgrader.Upgrade(w, request, nil)
		if err != nil {
			return err
		}
		manager.media.Serve(ticket, newPeer(manager, connection, session.ID()))
		return nil
	}
}
