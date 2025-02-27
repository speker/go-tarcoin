// Copyright 2018 The go-tarcoin Authors
// This file is part of the go-tarcoin library.
//
// The go-tarcoin library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-tarcoin library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-tarcoin library. If not, see <http://www.gnu.org/licenses/>.

package les

import (
	"crypto/rand"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/speker/go-tarcoin/crypto"
	"github.com/speker/go-tarcoin/p2p"
	"github.com/speker/go-tarcoin/p2p/enode"
)

func TestULCAnnounceThresholdLes2(t *testing.T) { testULCAnnounceThreshold(t, 2) }
func TestULCAnnounceThresholdLes3(t *testing.T) { testULCAnnounceThreshold(t, 3) }

func testULCAnnounceThreshold(t *testing.T, protocol int) {
	// todo figure out why it takes fetcher so longer to fetcher the announced header.
	t.Skip("Sometimes it can failed")
	var cases = []struct {
		height    []int
		threshold int
		expect    uint64
	}{
		{[]int{1}, 100, 1},
		{[]int{0, 0, 0}, 100, 0},
		{[]int{1, 2, 3}, 30, 3},
		{[]int{1, 2, 3}, 60, 2},
		{[]int{3, 2, 1}, 67, 1},
		{[]int{3, 2, 1}, 100, 1},
	}
	for _, testcase := range cases {
		var (
			servers   []*testServer
			teardowns []func()
			nodes     []*enode.Node
			ids       []string
		)
		for i := 0; i < len(testcase.height); i++ {
			s, n, teardown := newTestServerPeer(t, 0, protocol)

			servers = append(servers, s)
			nodes = append(nodes, n)
			teardowns = append(teardowns, teardown)
			ids = append(ids, n.String())
		}
		c, teardown := newTestLightPeer(t, protocol, ids, testcase.threshold)

		// Connect all servers.
		for i := 0; i < len(servers); i++ {
			connect(servers[i].handler, nodes[i].ID(), c.handler, protocol)
		}
		for i := 0; i < len(servers); i++ {
			for j := 0; j < testcase.height[i]; j++ {
				servers[i].backend.Commit()
			}
		}
		time.Sleep(1500 * time.Millisecond) // Ensure the fetcher has done its work.
		head := c.handler.backend.blockchain.CurrentHeader().Number.Uint64()
		if head != testcase.expect {
			t.Fatalf("chain height mismatch, want %d, got %d", testcase.expect, head)
		}

		// Release all servers and client resources.
		teardown()
		for i := 0; i < len(teardowns); i++ {
			teardowns[i]()
		}
	}
}

func connect(server *serverHandler, serverId enode.ID, client *clientHandler, protocol int) (*serverPeer, *clientPeer, error) {
	// Create a message pipe to communicate through
	app, net := p2p.MsgPipe()

	var id enode.ID
	rand.Read(id[:])

	peer1 := newServerPeer(protocol, NetworkId, true, p2p.NewPeer(serverId, "", nil), net) // Mark server as trusted
	peer2 := newClientPeer(protocol, NetworkId, p2p.NewPeer(id, "", nil), app)

	// Start the peerLight on a new thread
	errc1 := make(chan error, 1)
	errc2 := make(chan error, 1)
	go func() {
		select {
		case <-server.closeCh:
			errc1 <- p2p.DiscQuitting
		case errc1 <- server.handle(peer2):
		}
	}()
	go func() {
		select {
		case <-client.closeCh:
			errc1 <- p2p.DiscQuitting
		case errc1 <- client.handle(peer1):
		}
	}()

	select {
	case <-time.After(time.Millisecond * 100):
	case err := <-errc1:
		return nil, nil, fmt.Errorf("peerLight handshake error: %v", err)
	case err := <-errc2:
		return nil, nil, fmt.Errorf("peerFull handshake error: %v", err)
	}
	return peer1, peer2, nil
}

// newTestServerPeer creates server peer.
func newTestServerPeer(t *testing.T, blocks int, protocol int) (*testServer, *enode.Node, func()) {
	s, teardown := newServerEnv(t, blocks, protocol, nil, false, false, 0)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal("generate key err:", err)
	}
	s.handler.server.privateKey = key
	n := enode.NewV4(&key.PublicKey, net.ParseIP("127.0.0.1"), 35000, 35000)
	return s, n, teardown
}

// newTestLightPeer creates node with light sync mode
func newTestLightPeer(t *testing.T, protocol int, ulcServers []string, ulcFraction int) (*testClient, func()) {
	_, c, teardown := newClientServerEnv(t, 0, protocol, nil, ulcServers, ulcFraction, false, false)
	return c, teardown
}
