package control

import (
	"errors"
	"testing"
	"time"
)

func TestLeaseRequestReleaseAndEpoch(t *testing.T) {
	lease := New(time.Minute)

	state, granted, err := lease.Request("alice")
	if err != nil || !granted || state.Holder != "alice" || state.Epoch != 1 {
		t.Fatalf("unexpected first grant: %+v granted=%v err=%v", state, granted, err)
	}
	initialEpoch := state.Epoch

	state, granted, err = lease.Request("bob")
	if !errors.Is(err, ErrConflict) || granted || state.Queue[0] != "bob" {
		t.Fatalf("unexpected queued request: %+v granted=%v err=%v", state, granted, err)
	}

	state, err = lease.Release("alice", initialEpoch)
	if err != nil || state.Holder != "bob" || len(state.Queue) != 0 || state.Epoch <= initialEpoch {
		t.Fatalf("unexpected release: %+v err=%v", state, err)
	}

	if err := lease.Validate("alice", initialEpoch); !errors.Is(err, ErrNotHolder) {
		t.Fatalf("stale holder should be rejected after release: %v", err)
	}
}

func TestLeaseExpiresAndDisconnectsQueuedSession(t *testing.T) {
	now := time.Unix(100, 0)
	lease := New(time.Second)
	lease.now = func() time.Time { return now }

	state, _, err := lease.Request("alice")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = lease.Request("bob")
	if !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Second)

	state = lease.Snapshot()
	if state.Holder != "" || state.Epoch != 2 {
		t.Fatalf("expired lease was not invalidated: %+v", state)
	}
	state = lease.Disconnect("bob")
	if len(state.Queue) != 0 {
		t.Fatalf("disconnected requester remained queued: %+v", state)
	}
}

func TestLeaseRejectsStaleEpoch(t *testing.T) {
	lease := New(time.Minute)
	state, _, err := lease.Request("alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := lease.Validate("alice", state.Epoch-1); !errors.Is(err, ErrStaleEpoch) {
		t.Fatalf("expected stale epoch, got %v", err)
	}
}

func TestLeaseValidationRenewsActiveHolder(t *testing.T) {
	now := time.Unix(100, 0)
	lease := New(time.Second)
	lease.now = func() time.Time { return now }

	state, _, err := lease.Request("alice")
	if err != nil {
		t.Fatal(err)
	}
	firstExpiry := state.ExpiresAt

	now = now.Add(500 * time.Millisecond)
	if err := lease.Validate("alice", state.Epoch); err != nil {
		t.Fatal(err)
	}

	state = lease.Snapshot()
	if !state.ExpiresAt.After(firstExpiry) || state.Holder != "alice" {
		t.Fatalf("active validation did not renew lease: %+v", state)
	}
}
