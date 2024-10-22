// Copyright 2016 The go-tarcoin Authors
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

// Contains all the wrappers from the params package.

package gtrcn

import (
	"github.com/speker/go-tarcoin/p2p/discv5"
	"github.com/speker/go-tarcoin/params"
)

// MainnetGenesis returns the JSON spec to use for the main TarCoin network. It
// is actually empty since that defaults to the hard coded binary genesis block.
func MainnetGenesis() string {
	return ""
}

// RopstenGenesis returns the JSON spec to use for the Ropsten test network.
//func RopstenGenesis() string {
//	enc, err := json.Marshal(core.DefaultRopstenGenesisBlock())
//	if err != nil {
//		panic(err)
//	}
//	return string(enc)
//}

// RinkebyGenesis returns the JSON spec to use for the Rinkeby test network
//func RinkebyGenesis() string {
//	enc, err := json.Marshal(core.DefaultRinkebyGenesisBlock())
//	if err != nil {
//		panic(err)
//	}
//	return string(enc)
//}

// GoerliGenesis returns the JSON spec to use for the Goerli test network
//func GoerliGenesis() string {
//	enc, err := json.Marshal(core.DefaultGoerliGenesisBlock())
//	if err != nil {
//		panic(err)
//	}
//	return string(enc)
//}

// FoundationBootnodes returns the enode URLs of the P2P bootstrap nodes operated
// by the foundation running the V5 discovery protocol.
func FoundationBootnodes() *Enodes {
	nodes := &Enodes{nodes: make([]*discv5.Node, len(params.MainnetBootnodes))}
	for i, url := range params.MainnetBootnodes {
		nodes.nodes[i] = discv5.MustParseNode(url)
	}
	return nodes
}
