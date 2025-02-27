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

package trcnhash

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/speker/go-tarcoin/common"
	"github.com/speker/go-tarcoin/core/types"
	"github.com/speker/go-tarcoin/internal/testlog"
	"github.com/speker/go-tarcoin/log"
)

// Tests whether remote HTTP servers are correctly notified of new work.
func TestRemoteNotify(t *testing.T) {
	// Start a simple web server to capture notifications.
	sink := make(chan [3]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		blob, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Errorf("failed to read miner notification: %v", err)
		}
		var work [3]string
		if err := json.Unmarshal(blob, &work); err != nil {
			t.Errorf("failed to unmarshal miner notification: %v", err)
		}
		sink <- work
	}))
	defer server.Close()

	// Create the custom trcnhash engine.
	trcnhash := NewTester([]string{server.URL}, false)
	defer trcnhash.Close()

	// Stream a work task and ensure the notification bubbles out.
	header := &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(100)}
	block := types.NewBlockWithHeader(header)

	trcnhash.Seal(nil, block, nil, nil)
	select {
	case work := <-sink:
		if want := trcnhash.SealHash(header).Hex(); work[0] != want {
			t.Errorf("work packet hash mismatch: have %s, want %s", work[0], want)
		}
		if want := common.BytesToHash(SeedHash(header.Number.Uint64())).Hex(); work[1] != want {
			t.Errorf("work packet seed mismatch: have %s, want %s", work[1], want)
		}
		target := new(big.Int).Div(new(big.Int).Lsh(big.NewInt(1), 256), header.Difficulty)
		if want := common.BytesToHash(target.Bytes()).Hex(); work[2] != want {
			t.Errorf("work packet target mismatch: have %s, want %s", work[2], want)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("notification timed out")
	}
}

// Tests that pushing work packages fast to the miner doesn't cause any data race
// issues in the notifications.
func TestRemoteMultiNotify(t *testing.T) {
	// Start a simple web server to capture notifications.
	sink := make(chan [3]string, 64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		blob, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Errorf("failed to read miner notification: %v", err)
		}
		var work [3]string
		if err := json.Unmarshal(blob, &work); err != nil {
			t.Errorf("failed to unmarshal miner notification: %v", err)
		}
		sink <- work
	}))
	defer server.Close()

	// Create the custom trcnhash engine.
	trcnhash := NewTester([]string{server.URL}, false)
	trcnhash.config.Log = testlog.Logger(t, log.LvlWarn)
	defer trcnhash.Close()

	// Stream a lot of work task and ensure all the notifications bubble out.
	for i := 0; i < cap(sink); i++ {
		header := &types.Header{Number: big.NewInt(int64(i)), Difficulty: big.NewInt(100)}
		block := types.NewBlockWithHeader(header)
		trcnhash.Seal(nil, block, nil, nil)
	}

	for i := 0; i < cap(sink); i++ {
		select {
		case <-sink:
		case <-time.After(10 * time.Second):
			t.Fatalf("notification %d timed out", i)
		}
	}
}

// Tests whether stale solutions are correctly processed.
func TestStaleSubmission(t *testing.T) {
	trcnhash := NewTester(nil, true)
	defer trcnhash.Close()
	api := &API{trcnhash}

	fakeNonce, fakeDigest := types.BlockNonce{0x01, 0x02, 0x03}, common.HexToHash("deadbeef")

	testcases := []struct {
		headers     []*types.Header
		submitIndex int
		submitRes   bool
	}{
		// Case1: submit solution for the latest mining package
		{
			[]*types.Header{
				{ParentHash: common.BytesToHash([]byte{0xa}), Number: big.NewInt(1), Difficulty: big.NewInt(100000000)},
			},
			0,
			true,
		},
		// Case2: submit solution for the previous package but have same parent.
		{
			[]*types.Header{
				{ParentHash: common.BytesToHash([]byte{0xb}), Number: big.NewInt(2), Difficulty: big.NewInt(100000000)},
				{ParentHash: common.BytesToHash([]byte{0xb}), Number: big.NewInt(2), Difficulty: big.NewInt(100000001)},
			},
			0,
			true,
		},
		// Case3: submit stale but acceptable solution
		{
			[]*types.Header{
				{ParentHash: common.BytesToHash([]byte{0xc}), Number: big.NewInt(3), Difficulty: big.NewInt(100000000)},
				{ParentHash: common.BytesToHash([]byte{0xd}), Number: big.NewInt(9), Difficulty: big.NewInt(100000000)},
			},
			0,
			true,
		},
		// Case4: submit very old solution
		{
			[]*types.Header{
				{ParentHash: common.BytesToHash([]byte{0xe}), Number: big.NewInt(10), Difficulty: big.NewInt(100000000)},
				{ParentHash: common.BytesToHash([]byte{0xf}), Number: big.NewInt(17), Difficulty: big.NewInt(100000000)},
			},
			0,
			false,
		},
	}
	results := make(chan *types.Block, 16)

	for id, c := range testcases {
		for _, h := range c.headers {
			trcnhash.Seal(nil, types.NewBlockWithHeader(h), results, nil)
		}
		if res := api.SubmitWork(fakeNonce, trcnhash.SealHash(c.headers[c.submitIndex]), fakeDigest); res != c.submitRes {
			t.Errorf("case %d submit result mismatch, want %t, get %t", id+1, c.submitRes, res)
		}
		if !c.submitRes {
			continue
		}
		select {
		case res := <-results:
			if res.Header().Nonce != fakeNonce {
				t.Errorf("case %d block nonce mismatch, want %x, get %x", id+1, fakeNonce, res.Header().Nonce)
			}
			if res.Header().MixDigest != fakeDigest {
				t.Errorf("case %d block digest mismatch, want %x, get %x", id+1, fakeDigest, res.Header().MixDigest)
			}
			if res.Header().Difficulty.Uint64() != c.headers[c.submitIndex].Difficulty.Uint64() {
				t.Errorf("case %d block difficulty mismatch, want %d, get %d", id+1, c.headers[c.submitIndex].Difficulty, res.Header().Difficulty)
			}
			if res.Header().Number.Uint64() != c.headers[c.submitIndex].Number.Uint64() {
				t.Errorf("case %d block number mismatch, want %d, get %d", id+1, c.headers[c.submitIndex].Number.Uint64(), res.Header().Number.Uint64())
			}
			if res.Header().ParentHash != c.headers[c.submitIndex].ParentHash {
				t.Errorf("case %d block parent hash mismatch, want %s, get %s", id+1, c.headers[c.submitIndex].ParentHash.Hex(), res.Header().ParentHash.Hex())
			}
		case <-time.NewTimer(time.Second).C:
			t.Errorf("case %d fetch trcnhash result timeout", id+1)
		}
	}
}
