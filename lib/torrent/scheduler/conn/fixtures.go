package conn

import (
	"net"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/uber-go/tally"

	"code.uber.internal/infra/kraken/core"
	"code.uber.internal/infra/kraken/lib/torrent/networkevent"
	"code.uber.internal/infra/kraken/lib/torrent/storage"
	"code.uber.internal/infra/kraken/utils/testutil"
)

type noopEvents struct{}

func (e noopEvents) ConnClosed(*Conn) {}

// noopDeadline wraps a Conn which does not support deadlines (e.g. net.Pipe)
// and makes it accept deadlines.
type noopDeadline struct {
	net.Conn
}

func (n noopDeadline) SetDeadline(t time.Time) error      { return nil }
func (n noopDeadline) SetReadDeadline(t time.Time) error  { return nil }
func (n noopDeadline) SetWriteDeadline(t time.Time) error { return nil }

// PipeFixture returns Conns for both sides of a live connection for testing.
func PipeFixture(
	config Config, info *storage.TorrentInfo) (local *Conn, remote *Conn, cleanupFunc func()) {

	var cleanup testutil.Cleanup
	defer cleanup.Recover()

	nc1, nc2 := net.Pipe()
	cleanup.Add(func() { nc1.Close() })
	cleanup.Add(func() { nc2.Close() })

	var err error

	local, err = HandshakerFixture(config).newConn(
		noopDeadline{nc1}, core.PeerIDFixture(), info, false)
	if err != nil {
		panic(err)
	}

	remote, err = HandshakerFixture(config).newConn(
		noopDeadline{nc2}, core.PeerIDFixture(), info, true)
	if err != nil {
		panic(err)
	}

	return local, remote, cleanup.Run
}

// Fixture returns a single local Conn for testing.
func Fixture() (*Conn, func()) {
	var cleanup testutil.Cleanup
	defer cleanup.Recover()

	info, c := storage.TorrentInfoFixture(1, 1)
	cleanup.Add(c)

	local, _, c := PipeFixture(Config{}, info)
	cleanup.Add(c)

	return local, cleanup.Run
}

// HandshakerFixture returns a Handshaker for testing.
func HandshakerFixture(config Config) *Handshaker {
	return NewHandshaker(
		config,
		tally.NewTestScope("", nil),
		clock.New(),
		networkevent.NewTestProducer(),
		core.PeerIDFixture(),
		noopEvents{})
}

// ConfigFixture returns a Config for testing.
func ConfigFixture() Config {
	return Config{}.applyDefaults()
}
