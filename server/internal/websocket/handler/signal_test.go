package handler

import (
	"testing"

	"github.com/pion/webrtc/v4"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

// The fakes below embed the real interfaces so we only implement the methods
// signalRequest actually calls.
type fakeSession struct {
	types.Session
	profile types.MemberProfile
}

func (s *fakeSession) ID() string { return "test-session" }

func (s *fakeSession) Profile() types.MemberProfile { return s.profile }

func (s *fakeSession) PrivateModeEnabled() bool { return false }

func (s *fakeSession) Send(string, any) {}

type fakeWebRTC struct {
	types.WebRTCManager
	peer *fakePeer
}

func (w *fakeWebRTC) ICEServers() []types.ICEServer { return nil }

func (w *fakeWebRTC) CreatePeer(types.Session) (*webrtc.SessionDescription, types.WebRTCPeer, error) {
	w.peer = &fakePeer{}
	return &webrtc.SessionDescription{SDP: "fake-offer"}, w.peer, nil
}

type fakePeer struct {
	types.WebRTCPeer
	video types.PeerVideoRequest
}

func (p *fakePeer) SetPaused(bool) error { return nil }

func (p *fakePeer) SetVideo(v types.PeerVideoRequest) error {
	p.video = v
	return nil
}

func (p *fakePeer) Video() types.PeerVideo { return types.PeerVideo{} }

func (p *fakePeer) SetAudio(types.PeerAudioRequest) error { return nil }

func (p *fakePeer) Audio() types.PeerAudio { return types.PeerAudio{} }

type fakeCapture struct {
	types.CaptureManager
	ids []string
}

func (c *fakeCapture) Video() types.EncodedStreamSelector {
	return &fakeSelector{ids: c.ids}
}

type fakeSelector struct {
	types.EncodedStreamSelector
	ids []string
}

func (s *fakeSelector) IDs() []string { return s.ids }

func newSignalTestHandler(capture types.CaptureManager) (*MessageHandlerCtx, *fakeWebRTC) {
	webrtcManager := &fakeWebRTC{}
	h := &MessageHandlerCtx{
		capture: capture,
		webrtc:  webrtcManager,
	}
	return h, webrtcManager
}

func watchSession() *fakeSession {
	return &fakeSession{profile: types.MemberProfile{CanWatch: true}}
}

// On master, an empty video id list made signalRequest index videos[0] and
// panic the whole server. It must return a descriptive error instead.
func TestSignalRequestNoVideoStreams(t *testing.T) {
	h, _ := newSignalTestHandler(&fakeCapture{ids: nil})

	payload := &message.SignalRequest{}
	if err := h.signalRequest(watchSession(), payload); err == nil {
		t.Fatal("expected an error when no video streams are configured, got nil")
	}
}

// When streams exist and the client sends no selector, the first video id is
// used as the default stream.
func TestSignalRequestDefaultsToFirstVideo(t *testing.T) {
	h, webrtcManager := newSignalTestHandler(&fakeCapture{ids: []string{"high", "low", "medium"}})

	payload := &message.SignalRequest{}
	if err := h.signalRequest(watchSession(), payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	selector := webrtcManager.peer.video.Selector
	if selector == nil {
		t.Fatal("expected a default selector to be set")
	}
	if selector.ID != "high" {
		t.Fatalf("expected default video id %q, got %q", "high", selector.ID)
	}
	if selector.Type != types.StreamSelectorTypeExact {
		t.Fatalf("expected exact selector type, got %v", selector.Type)
	}
}
