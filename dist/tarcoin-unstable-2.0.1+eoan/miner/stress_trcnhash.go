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

// +build none

// This file contains a miner stress test based on the Trcnhash consensus engine.
package main

import (
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/speker/go-tarcoin/accounts/keystore"
	"github.com/speker/go-tarcoin/common"
	"github.com/speker/go-tarcoin/common/fdlimit"
	"github.com/speker/go-tarcoin/consensus/trcnhash"
	"github.com/speker/go-tarcoin/core"
	"github.com/speker/go-tarcoin/core/types"
	"github.com/speker/go-tarcoin/crypto"
	"github.com/speker/go-tarcoin/trcn"
	"github.com/speker/go-tarcoin/trcn/downloader"
	"github.com/speker/go-tarcoin/log"
	"github.com/speker/go-tarcoin/miner"
	"github.com/speker/go-tarcoin/node"
	"github.com/speker/go-tarcoin/p2p"
	"github.com/speker/go-tarcoin/p2p/enode"
	"github.com/speker/go-tarcoin/params"
)

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	fdlimit.Raise(2048)

	// Generate a batch of accounts to seal and fund with
	faucets := make([]*ecdsa.PrivateKey, 128)
	for i := 0; i < len(faucets); i++ {
		faucets[i], _ = crypto.GenerateKey()
	}
	// Pre-generate the trcnhash mining DAG so we don't race
	trcnhash.MakeDataset(1, filepath.Join(os.Getenv("HOME"), ".trcnhash"))

	// Create an Trcnhash network based off of the Ropsten config
	genesis := makeGenesis(faucets)

	var (
		nodes  []*node.Node
		enodes []*enode.Node
	)
	for i := 0; i < 4; i++ {
		// Start the node and wait until it's up
		node, err := makeMiner(genesis)
		if err != nil {
			panic(err)
		}
		defer node.Close()

		for node.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to al the previous ones
		for _, n := range enodes {
			node.Server().AddPeer(n)
		}
		// Start tracking the node and it's enode
		nodes = append(nodes, node)
		enodes = append(enodes, node.Server().Self())

		// Inject the signer key and start sealing with it
		store := node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
		if _, err := store.NewAccount(""); err != nil {
			panic(err)
		}
	}
	// Iterate over all the nodes and start signing with them
	time.Sleep(3 * time.Second)

	for _, node := range nodes {
		var tarcoin *trcn.TarCoin
		if err := node.Service(&tarcoin); err != nil {
			panic(err)
		}
		if err := tarcoin.StartMining(1); err != nil {
			panic(err)
		}
	}
	time.Sleep(3 * time.Second)

	// Start injecting transactions from the faucets like crazy
	nonces := make([]uint64, len(faucets))
	for {
		index := rand.Intn(len(faucets))

		// Fetch the accessor for the relevant signer
		var tarcoin *trcn.TarCoin
		if err := nodes[index%len(nodes)].Service(&tarcoin); err != nil {
			panic(err)
		}
		// Create a self transaction and inject into the pool
		tx, err := types.SignTx(types.NewTransaction(nonces[index], crypto.PubkeyToAddress(faucets[index].PublicKey), new(big.Int), 21000, big.NewInt(100000000000+rand.Int63n(65536)), nil), types.HomesteadSigner{}, faucets[index])
		if err != nil {
			panic(err)
		}
		if err := tarcoin.TxPool().AddLocal(tx); err != nil {
			panic(err)
		}
		nonces[index]++

		// Wait if we're too saturated
		if pend, _ := tarcoin.TxPool().Stats(); pend > 2048 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// makeGenesis creates a custom Trcnhash genesis block based on some pre-defined
// faucet accounts.
func makeGenesis(faucets []*ecdsa.PrivateKey) *core.Genesis {
	genesis := core.DefaultRopstenGenesisBlock()
	genesis.Difficulty = params.MinimumDifficulty
	genesis.GasLimit = 25000000

	genesis.Config.ChainID = big.NewInt(18)
	genesis.Config.EIP150Hash = common.Hash{}

	genesis.Alloc = core.GenesisAlloc{}
	for _, faucet := range faucets {
		genesis.Alloc[crypto.PubkeyToAddress(faucet.PublicKey)] = core.GenesisAccount{
			Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		}
	}
	return genesis
}

func makeMiner(genesis *core.Genesis) (*node.Node, error) {
	// Define the basic configurations for the TarCoin node
	datadir, _ := ioutil.TempDir("", "")

	config := &node.Config{
		Name:    "gtrcn",
		Version: params.Version,
		DataDir: datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
		NoUSB:             true,
		UseLightweightKDF: true,
	}
	// Start the node and configure a full TarCoin node on it
	stack, err := node.New(config)
	if err != nil {
		return nil, err
	}
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return trcn.New(ctx, &trcn.Config{
			Genesis:         genesis,
			NetworkId:       genesis.Config.ChainID.Uint64(),
			SyncMode:        downloader.FullSync,
			DatabaseCache:   256,
			DatabaseHandles: 256,
			TxPool:          core.DefaultTxPoolConfig,
			GPO:             trcn.DefaultConfig.GPO,
			Trcnhash:          trcn.DefaultConfig.Trcnhash,
			Miner: miner.Config{
				GasFloor: genesis.GasLimit * 9 / 10,
				GasCeil:  genesis.GasLimit * 11 / 10,
				GasPrice: big.NewInt(1),
				Recommit: time.Second,
			},
		})
	}); err != nil {
		return nil, err
	}
	// Start the node and return if successful
	return stack, stack.Start()
}