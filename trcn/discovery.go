// Copyright 2019 The go-tarcoin Authors
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

package trcn

import (
	"github.com/spker/go-tarcoin/core"
	"github.com/spker/go-tarcoin/core/forkid"
	"github.com/spker/go-tarcoin/p2p"
	"github.com/spker/go-tarcoin/p2p/dnsdisc"
	"github.com/spker/go-tarcoin/p2p/enode"
	"github.com/spker/go-tarcoin/rlp"
)

// ethEntry is the "trcn" ENR entry which advertises trcn protocol
// on the discovery network.
type ethEntry struct {
	ForkID forkid.ID // Fork identifier per EIP-2124

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (e ethEntry) ENRKey() string {
	return "trcn"
}

// startEthEntryUpdate starts the ENR updater loop.
func (trcn *TarCoin) startEthEntryUpdate(ln *enode.LocalNode) {
	var newHead = make(chan core.ChainHeadEvent, 10)
	sub := trcn.blockchain.SubscribeChainHeadEvent(newHead)

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-newHead:
				ln.Set(trcn.currentEthEntry())
			case <-sub.Err():
				// Would be nice to sync with trcn.Stop, but there is no
				// good way to do that.
				return
			}
		}
	}()
}

func (trcn *TarCoin) currentEthEntry() *ethEntry {
	return &ethEntry{ForkID: forkid.NewID(trcn.blockchain)}
}

// setupDiscovery creates the node discovery source for the trcn protocol.
func (trcn *TarCoin) setupDiscovery(cfg *p2p.Config) (enode.Iterator, error) {
	if cfg.NoDiscovery || len(trcn.config.DiscoveryURLs) == 0 {
		return nil, nil
	}
	client := dnsdisc.NewClient(dnsdisc.Config{})
	return client.NewIterator(trcn.config.DiscoveryURLs...)
}
