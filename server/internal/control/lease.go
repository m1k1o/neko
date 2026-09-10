package control

import (
	"errors"
	"slices"
	"sync"
	"time"
)

var (
	ErrConflict     = errors.New("control lease is held by another session")
	ErrNotHolder    = errors.New("session does not hold the control lease")
	ErrStaleEpoch   = errors.New("control lease epoch is stale")
	ErrLeaseExpired = errors.New("control lease has expired")
)

// State is the serializable control lease snapshot shared with API and
// signaling adapters. Epoch is incremented whenever ownership changes.
type State struct {
	Holder    string
	Epoch     uint64
	ExpiresAt time.Time
	Queue     []string
}

// Lease owns the concurrency and expiry rules for room control. It does not
// know about sessions or desktop input, which keeps it unit-testable and
// usable by both REST and WebRTC adapters.
type Lease struct {
	mu  sync.Mutex
	ttl time.Duration
	now func() time.Time

	holder    string
	epoch     uint64
	expiresAt time.Time
	queue     []string
}

func New(ttl time.Duration) *Lease {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &Lease{ttl: ttl, now: time.Now}
}

func (l *Lease) Snapshot() State {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	return l.stateLocked()
}

// Request grants an unheld lease, renews an existing holder, or queues a
// distinct requester. The bool reports whether the caller became the holder.
func (l *Lease) Request(sessionID string) (State, bool, error) {
	if sessionID == "" {
		return State{}, false, errors.New("session id is required")
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()

	if l.holder == "" {
		l.grantLocked(sessionID)
		return l.stateLocked(), true, nil
	}
	if l.holder == sessionID {
		l.renewLocked()
		return l.stateLocked(), true, nil
	}
	if !slices.Contains(l.queue, sessionID) && len(l.queue) < 32 {
		l.queue = append(l.queue, sessionID)
	}
	return l.stateLocked(), false, ErrConflict
}

// Grant forcibly assigns the lease to sessionID. It is used by implicit
// hosting and administrator takeover.
func (l *Lease) Grant(sessionID string) State {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	l.grantLocked(sessionID)
	l.queue = nil
	return l.stateLocked()
}

func (l *Lease) Renew(sessionID string, epoch uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	if l.holder == "" {
		return ErrLeaseExpired
	}
	if l.holder != sessionID {
		return ErrNotHolder
	}
	if l.epoch != epoch {
		return ErrStaleEpoch
	}
	l.renewLocked()
	return nil
}

func (l *Lease) Validate(sessionID string, epoch uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	if l.holder == "" {
		return ErrLeaseExpired
	}
	if l.holder != sessionID {
		return ErrNotHolder
	}
	if l.epoch != epoch {
		return ErrStaleEpoch
	}
	l.renewLocked()
	return nil
}

func (l *Lease) Release(sessionID string, epoch uint64) (State, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	if l.holder == "" {
		return l.stateLocked(), ErrLeaseExpired
	}
	if l.holder != sessionID {
		return l.stateLocked(), ErrNotHolder
	}
	if l.epoch != epoch {
		return l.stateLocked(), ErrStaleEpoch
	}
	l.releaseLocked(true)
	return l.stateLocked(), nil
}

// ForceRelease clears the current holder and invalidates all outstanding
// input packets. The next request starts a new epoch.
func (l *Lease) ForceRelease() State {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	l.releaseLocked(false)
	l.queue = nil
	return l.stateLocked()
}

func (l *Lease) Disconnect(sessionID string) State {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.expireLocked()
	if l.holder == sessionID {
		l.releaseLocked(true)
	} else {
		l.queue = slices.DeleteFunc(l.queue, func(id string) bool { return id == sessionID })
	}
	return l.stateLocked()
}

func (l *Lease) expireLocked() {
	if l.holder != "" && !l.expiresAt.IsZero() && !l.now().Before(l.expiresAt) {
		l.releaseLocked(false)
	}
}

func (l *Lease) grantLocked(sessionID string) {
	l.holder = sessionID
	l.epoch++
	l.renewLocked()
}

func (l *Lease) renewLocked() {
	l.expiresAt = l.now().Add(l.ttl)
}

func (l *Lease) releaseLocked(promote bool) {
	l.holder = ""
	l.expiresAt = time.Time{}
	l.epoch++
	if promote && len(l.queue) > 0 {
		next := l.queue[0]
		l.queue = l.queue[1:]
		l.grantLocked(next)
	}
}

func (l *Lease) stateLocked() State {
	queue := append([]string(nil), l.queue...)
	return State{Holder: l.holder, Epoch: l.epoch, ExpiresAt: l.expiresAt, Queue: queue}
}
