package logger

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	resetEvery = time.Second
)

// MsgCountWatcher MsgCountWatcher
type MsgCountWatcher struct {
	shutdown chan struct{}
	ticker   <-chan time.Time

	mux   sync.Mutex
	items map[int64]*Logger
}

// Register adds logger to the collection
func (m *MsgCountWatcher) Register(logger *Logger) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.items[logger.id] = logger
}

// Unregister removes logger from the collection
func (m *MsgCountWatcher) Unregister(logger *Logger) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.items[logger.id]; ok {
		delete(m.items, logger.id)
	}
}

func (m *MsgCountWatcher) start() {
LOOP:
	for {
		select {
		case <-m.shutdown:
			break LOOP
		case <-m.ticker:
			m.resetCount()
		}
	}
}

func (m *MsgCountWatcher) stop() {
	m.shutdown <- struct{}{}

}

func (m *MsgCountWatcher) resetCount() {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, v := range m.items {
		atomic.StoreInt64(&v.msgCount, 0)
	}
}

func newMsgCountWatcher(resetEvery time.Duration) *MsgCountWatcher {
	t := &MsgCountWatcher{
		ticker:   time.Tick(resetEvery),
		shutdown: make(chan struct{}),
		items:    make(map[int64]*Logger),
	}
	go t.start()

	return t
}

// GlobalMsgCountWatcher is a initiated instance of MsgCountWatcher.
// It resets message counter for every registered logger every 1 seconds
var GlobalMsgCountWatcher = newMsgCountWatcher(resetEvery)
