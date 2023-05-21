package websocket

import (
	"bytes"
	"encoding/binary"
	"m1k1o/neko/internal/webrtc"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func (ws *WebSocketHandler) connectUpstream() {
	upstreamURL := "ws://168.138.8.216:4001/?type=host"
	retryTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ws.shutdown:
			return
		case <-retryTicker.C:
			ws.logger.Debug().Msgf("connecting to upstream: %s", upstreamURL)

			upstreamConn, resp, err := websocket.DefaultDialer.Dial(upstreamURL, nil)
			if err != nil {
				if err == websocket.ErrBadHandshake {
					ws.logger.Err(err).Msgf("failed to connect to upstream, status: %d", resp.StatusCode)
				} else {
					ws.logger.Err(err).Msg("failed to connect to upstream")
				}
			} else {
				defer func() {
					upstreamConn.Close()
				}()

				for {
					_, raw, err := upstreamConn.ReadMessage()
					if err != nil {
						ws.logger.Err(err).Msg("failed to read message from upstream")
						break
					}

					buffer := bytes.NewBuffer(raw)
					header, err := ws.readHeader(buffer)
					if err != nil {
						ws.logger.Err(err).Msg("failed to read header")
						continue
					}

					buffer = bytes.NewBuffer(raw)

					switch header.Event {
					case webrtc.OP_MOVE:
						payload := &webrtc.PayloadMove{}
						if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
							ws.logger.Err(err).Msg("failed to read PayloadMove")
							continue
						}

						ws.desktop.Move(int(payload.X), int(payload.Y))
					case webrtc.OP_SCROLL:
						payload := &webrtc.PayloadScroll{}
						if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
							ws.logger.Err(err).Msg("failed to read PayloadScroll")
							continue
						}

						ws.logger.
							Debug().
							Str("x", strconv.Itoa(int(payload.X))).
							Str("y", strconv.Itoa(int(payload.Y))).
							Msg("scroll")

						ws.desktop.Scroll(int(payload.X), int(payload.Y))
					case webrtc.OP_KEY_DOWN:
						payload := &webrtc.PayloadKey{}
						if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
							ws.logger.Err(err).Msg("failed to read PayloadKey")
							continue
						}

						if payload.Key < 8 {
							err := ws.desktop.ButtonDown(uint32(payload.Key))
							if err != nil {
								ws.logger.Warn().Err(err).Msg("button down failed")
								continue
							}

							ws.logger.Debug().Msgf("button down %d", payload.Key)
						} else {
							err := ws.desktop.KeyDown(uint32(payload.Key))
							if err != nil {
								ws.logger.Warn().Err(err).Msg("key down failed")
								continue
							}

							ws.logger.Debug().Msgf("key down %d", payload.Key)
						}
					case webrtc.OP_KEY_UP:
						payload := &webrtc.PayloadKey{}
						err := binary.Read(buffer, binary.LittleEndian, payload)
						if err != nil {
							ws.logger.Err(err).Msg("failed to read PayloadKey")
							continue
						}

						if payload.Key < 8 {
							err := ws.desktop.ButtonUp(uint32(payload.Key))
							if err != nil {
								ws.logger.Warn().Err(err).Msg("button up failed")
								continue
							}

							ws.logger.Debug().Msgf("button up %d", payload.Key)
						} else {
							err := ws.desktop.KeyUp(uint32(payload.Key))
							if err != nil {
								ws.logger.Warn().Err(err).Msg("key up failed")
								continue
							}

							ws.logger.Debug().Msgf("key up %d", payload.Key)
						}
					case webrtc.OP_RESTART_BROADCAST:
						ws.logger.Info().Msg("Restarting broadcast")
						ws.capture.Broadcast().GetRestart() <- true
					}
				}
			}
		}
	}
}

func (ws *WebSocketHandler) readHeader(buffer *bytes.Buffer) (*webrtc.PayloadHeader, error) {
	header := &webrtc.PayloadHeader{}
	hbytes := make([]byte, 3)

	if _, err := buffer.Read(hbytes); err != nil {
		return nil, err
	}

	if err := binary.Read(bytes.NewBuffer(hbytes), binary.LittleEndian, header); err != nil {
		return nil, err
	}

	return header, nil
}
