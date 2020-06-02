// Copyright 2014 The go-tarcoin Authors
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

// Package trcn implements the TarCoin protocol.
package trcn

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/spker/go-tarcoin/accounts"
	"github.com/spker/go-tarcoin/accounts/abi/bind"
	"github.com/spker/go-tarcoin/common"
	"github.com/spker/go-tarcoin/common/hexutil"
	"github.com/spker/go-tarcoin/consensus"
	"github.com/spker/go-tarcoin/consensus/clique"
	"github.com/spker/go-tarcoin/consensus/trcnhash"
	"github.com/spker/go-tarcoin/core"
	"github.com/spker/go-tarcoin/core/bloombits"
	"github.com/spker/go-tarcoin/core/rawdb"
	"github.com/spker/go-tarcoin/core/types"
	"github.com/spker/go-tarcoin/core/vm"
	"github.com/spker/go-tarcoin/trcn/downloader"
	"github.com/spker/go-tarcoin/trcn/filters"
	"github.com/spker/go-tarcoin/trcn/gasprice"
	"github.com/spker/go-tarcoin/trcndb"
	"github.com/spker/go-tarcoin/event"
	"github.com/spker/go-tarcoin/internal/trcnapi"
	"github.com/spker/go-tarcoin/log"
	"github.com/spker/go-tarcoin/miner"
	"github.com/spker/go-tarcoin/node"
	"github.com/spker/go-tarcoin/p2p"
	"github.com/spker/go-tarcoin/p2p/enode"
	"github.com/spker/go-tarcoin/p2p/enr"
	"github.com/spker/go-tarcoin/params"
	"github.com/spker/go-tarcoin/rlp"
	"github.com/spker/go-tarcoin/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	APIs() []rpc.API
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
	SetContractBackend(bind.ContractBackend)
}

// TarCoin implements the TarCoin full node service.
type TarCoin struct {
	config *Config

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer
	dialCandiates   enode.Iterator

	// DB interfaces
	chainDb trcndb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests     chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer      *core.ChainIndexer             // Bloom indexer operating during block imports
	closeBloomHandler chan struct{}

	APIBackend *EthAPIBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	trcnbase common.Address

	networkID     uint64
	netRPCService *trcnapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and trcnbase)
}

func (s *TarCoin) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// SetClient sets a rpc client which connecting to our local node.
func (s *TarCoin) SetContractBackend(backend bind.ContractBackend) {
	// Pass the rpc client to les server if it is enabled.
	if s.lesServer != nil {
		s.lesServer.SetContractBackend(backend)
	}
}

// New creates a new TarCoin object (including the
// initialisation of the common TarCoin object)
func New(ctx *node.ServiceContext, config *Config) (*TarCoin, error) {
	// Ensure configuration values are compatible and sane
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run trcn.TarCoin in light sync mode, use les.LightEthereum")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	if config.Miner.GasPrice == nil || config.Miner.GasPrice.Cmp(common.Big0) <= 0 {
		log.Warn("Sanitizing invalid miner gas price", "provided", config.Miner.GasPrice, "updated", DefaultConfig.Miner.GasPrice)
		config.Miner.GasPrice = new(big.Int).Set(DefaultConfig.Miner.GasPrice)
	}
	if config.NoPruning && config.TrieDirtyCache > 0 {
		if config.SnapshotCache > 0 {
			config.TrieCleanCache += config.TrieDirtyCache * 3 / 5
			config.SnapshotCache += config.TrieDirtyCache * 2 / 5
		} else {
			config.TrieCleanCache += config.TrieDirtyCache
		}
		config.TrieDirtyCache = 0
	}
	log.Info("Allocated trie memory caches", "clean", common.StorageSize(config.TrieCleanCache)*1024*1024, "dirty", common.StorageSize(config.TrieDirtyCache)*1024*1024)

	// Assemble the TarCoin object
	chainDb, err := ctx.OpenDatabaseWithFreezer("chaindata", config.DatabaseCache, config.DatabaseHandles, config.DatabaseFreezer, "trcn/db/chaindata/")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	trcn := &TarCoin{
		config:            config,
		chainDb:           chainDb,
		eventMux:          ctx.EventMux,
		accountManager:    ctx.AccountManager,
		engine:            CreateConsensusEngine(ctx, chainConfig, &config.Trcnhash, config.Miner.Notify, config.Miner.Noverify, chainDb),
		closeBloomHandler: make(chan struct{}),
		networkID:         config.NetworkId,
		gasPrice:          config.Miner.GasPrice,
		trcnbase:         config.Miner.Trcnbase,
		bloomRequests:     make(chan chan *bloombits.Retrieval),
		bloomIndexer:      NewBloomIndexer(chainDb, params.BloomBitsBlocks, params.BloomConfirms),
	}

	bcVersion := rawdb.ReadDatabaseVersion(chainDb)
	var dbVer = "<nil>"
	if bcVersion != nil {
		dbVer = fmt.Sprintf("%d", *bcVersion)
	}
	log.Info("Initialising TarCoin protocol", "versions", ProtocolVersions, "network", config.NetworkId, "dbversion", dbVer)

	if !config.SkipBcVersionCheck {
		if bcVersion != nil && *bcVersion > core.BlockChainVersion {
			return nil, fmt.Errorf("database version is v%d, Gtrcn %s only supports v%d", *bcVersion, params.VersionWithMeta, core.BlockChainVersion)
		} else if bcVersion == nil || *bcVersion < core.BlockChainVersion {
			log.Warn("Upgrade blockchain database version", "from", dbVer, "to", core.BlockChainVersion)
			rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
		}
	}
	var (
		vmConfig = vm.Config{
			EnablePreimageRecording: config.EnablePreimageRecording,
			EWASMInterpreter:        config.EWASMInterpreter,
			EVMInterpreter:          config.EVMInterpreter,
		}
		cacheConfig = &core.CacheConfig{
			TrieCleanLimit:      config.TrieCleanCache,
			TrieCleanNoPrefetch: config.NoPrefetch,
			TrieDirtyLimit:      config.TrieDirtyCache,
			TrieDirtyDisabled:   config.NoPruning,
			TrieTimeLimit:       config.TrieTimeout,
			SnapshotLimit:       config.SnapshotCache,
		}
	)
	trcn.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, chainConfig, trcn.engine, vmConfig, trcn.shouldPreserve, &config.TxLookupLimit)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		trcn.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	trcn.bloomIndexer.Start(trcn.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	trcn.txPool = core.NewTxPool(config.TxPool, chainConfig, trcn.blockchain)

	// Permit the downloader to use the trie cache allowance during fast sync
	cacheLimit := cacheConfig.TrieCleanLimit + cacheConfig.TrieDirtyLimit + cacheConfig.SnapshotLimit
	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	if trcn.protocolManager, err = NewProtocolManager(chainConfig, checkpoint, config.SyncMode, config.NetworkId, trcn.eventMux, trcn.txPool, trcn.engine, trcn.blockchain, chainDb, cacheLimit, config.Whitelist); err != nil {
		return nil, err
	}
	trcn.miner = miner.New(trcn, &config.Miner, chainConfig, trcn.EventMux(), trcn.engine, trcn.isLocalBlock)
	trcn.miner.SetExtra(makeExtraData(config.Miner.ExtraData))

	trcn.APIBackend = &EthAPIBackend{ctx.ExtRPCEnabled(), trcn, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.GasPrice
	}
	trcn.APIBackend.gpo = gasprice.NewOracle(trcn.APIBackend, gpoParams)

	trcn.dialCandiates, err = trcn.setupDiscovery(&ctx.Config.P2P)
	if err != nil {
		return nil, err
	}

	return trcn, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gtrcn",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateConsensusEngine creates the required type of consensus engine instance for an TarCoin service
func CreateConsensusEngine(ctx *node.ServiceContext, chainConfig *params.ChainConfig, config *trcnhash.Config, notify []string, noverify bool, db trcndb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch config.PowMode {
	case trcnhash.ModeFake:
		log.Warn("Trcnhash used in fake mode")
		return trcnhash.NewFaker()
	case trcnhash.ModeTest:
		log.Warn("Trcnhash used in test mode")
		return trcnhash.NewTester(nil, noverify)
	case trcnhash.ModeShared:
		log.Warn("Trcnhash used in shared mode")
		return trcnhash.NewShared()
	default:
		engine := trcnhash.New(trcnhash.Config{
			CacheDir:         ctx.ResolvePath(config.CacheDir),
			CachesInMem:      config.CachesInMem,
			CachesOnDisk:     config.CachesOnDisk,
			CachesLockMmap:   config.CachesLockMmap,
			DatasetDir:       config.DatasetDir,
			DatasetsInMem:    config.DatasetsInMem,
			DatasetsOnDisk:   config.DatasetsOnDisk,
			DatasetsLockMmap: config.DatasetsLockMmap,
		}, notify, noverify)
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs return the collection of RPC services the tarcoin package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *TarCoin) APIs() []rpc.API {
	apis := trcnapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the les server
	if s.lesServer != nil {
		apis = append(apis, s.lesServer.APIs()...)
	}
	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append any APIs exposed explicitly by the les server
	if s.lesServer != nil {
		apis = append(apis, s.lesServer.APIs()...)
	}

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "trcn",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "trcn",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "trcn",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "trcn",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *TarCoin) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *TarCoin) Trcnbase() (eb common.Address, err error) {
	s.lock.RLock()
	trcnbase := s.trcnbase
	s.lock.RUnlock()

	if trcnbase != (common.Address{}) {
		return trcnbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			trcnbase := accounts[0].Address

			s.lock.Lock()
			s.trcnbase = trcnbase
			s.lock.Unlock()

			log.Info("Trcnbase automatically configured", "address", trcnbase)
			return trcnbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("trcnbase must be explicitly specified")
}

// isLocalBlock checks whether the specified block is mined
// by local miner accounts.
//
// We regard two types of accounts as local miner account: trcnbase
// and accounts specified via `txpool.locals` flag.
func (s *TarCoin) isLocalBlock(block *types.Block) bool {
	author, err := s.engine.Author(block.Header())
	if err != nil {
		log.Warn("Failed to retrieve block author", "number", block.NumberU64(), "hash", block.Hash(), "err", err)
		return false
	}
	// Check whether the given address is trcnbase.
	s.lock.RLock()
	trcnbase := s.trcnbase
	s.lock.RUnlock()
	if author == trcnbase {
		return true
	}
	// Check whether the given address is specified by `txpool.local`
	// CLI flag.
	for _, account := range s.config.TxPool.Locals {
		if account == author {
			return true
		}
	}
	return false
}

// shouldPreserve checks whether we should preserve the given block
// during the chain reorg depending on whether the author of block
// is a local account.
func (s *TarCoin) shouldPreserve(block *types.Block) bool {
	// The reason we need to disable the self-reorg preserving for clique
	// is it can be probable to introduce a deadlock.
	//
	// e.g. If there are 7 available signers
	//
	// r1   A
	// r2     B
	// r3       C
	// r4         D
	// r5   A      [X] F G
	// r6    [X]
	//
	// In the round5, the inturn signer E is offline, so the worst case
	// is A, F and G sign the block of round5 and reject the block of opponents
	// and in the round6, the last available signer B is offline, the whole
	// network is stuck.
	if _, ok := s.engine.(*clique.Clique); ok {
		return false
	}
	return s.isLocalBlock(block)
}

// SetTrcnbase sets the mining reward address.
func (s *TarCoin) SetTrcnbase(trcnbase common.Address) {
	s.lock.Lock()
	s.trcnbase = trcnbase
	s.lock.Unlock()

	s.miner.SetTrcnbase(trcnbase)
}

// StartMining starts the miner with the given number of CPU threads. If mining
// is already running, this method adjust the number of threads allowed to use
// and updates the minimum price required by the transaction pool.
func (s *TarCoin) StartMining(threads int) error {
	// Update the thread count within the consensus engine
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := s.engine.(threaded); ok {
		log.Info("Updated mining threads", "threads", threads)
		if threads == 0 {
			threads = -1 // Disable the miner from within
		}
		th.SetThreads(threads)
	}
	// If the miner was not running, initialize it
	if !s.IsMining() {
		// Propagate the initial price point to the transaction pool
		s.lock.RLock()
		price := s.gasPrice
		s.lock.RUnlock()
		s.txPool.SetGasPrice(price)

		// Configure the local mining address
		eb, err := s.Trcnbase()
		if err != nil {
			log.Error("Cannot start mining without trcnbase", "err", err)
			return fmt.Errorf("trcnbase missing: %v", err)
		}
		if clique, ok := s.engine.(*clique.Clique); ok {
			wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
			if wallet == nil || err != nil {
				log.Error("Trcnbase account unavailable locally", "err", err)
				return fmt.Errorf("signer missing: %v", err)
			}
			clique.Authorize(eb, wallet.SignData)
		}
		// If mining is started, we can disable the transaction rejection mechanism
		// introduced to speed sync times.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)

		go s.miner.Start(eb)
	}
	return nil
}

// StopMining terminates the miner, both at the consensus engine level as well as
// at the block creation level.
func (s *TarCoin) StopMining() {
	// Update the thread count within the consensus engine
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := s.engine.(threaded); ok {
		th.SetThreads(-1)
	}
	// Stop the block creating itself
	s.miner.Stop()
}

func (s *TarCoin) IsMining() bool      { return s.miner.Mining() }
func (s *TarCoin) Miner() *miner.Miner { return s.miner }

func (s *TarCoin) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *TarCoin) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *TarCoin) TxPool() *core.TxPool               { return s.txPool }
func (s *TarCoin) EventMux() *event.TypeMux           { return s.eventMux }
func (s *TarCoin) Engine() consensus.Engine { return s.engine }
func (s *TarCoin) ChainDb() trcndb.Database { return s.chainDb }
func (s *TarCoin) IsListening() bool        { return true } // Always listening
func (s *TarCoin) EthVersion() int                    { return int(ProtocolVersions[0]) }
func (s *TarCoin) NetVersion() uint64                 { return s.networkID }
func (s *TarCoin) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *TarCoin) Synced() bool                       { return atomic.LoadUint32(&s.protocolManager.acceptTxs) == 1 }
func (s *TarCoin) ArchiveMode() bool                  { return s.config.NoPruning }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *TarCoin) Protocols() []p2p.Protocol {
	protos := make([]p2p.Protocol, len(ProtocolVersions))
	for i, vsn := range ProtocolVersions {
		protos[i] = s.protocolManager.makeProtocol(vsn)
		protos[i].Attributes = []enr.Entry{s.currentEthEntry()}
		protos[i].DialCandidates = s.dialCandiates
	}
	if s.lesServer != nil {
		protos = append(protos, s.lesServer.Protocols()...)
	}
	return protos
}

// Start implements node.Service, starting all internal goroutines needed by the
// TarCoin protocol implementation.
func (s *TarCoin) Start(srvr *p2p.Server) error {
	s.startEthEntryUpdate(srvr.LocalNode())

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers(params.BloomBitsBlocks)

	// Start the RPC service
	s.netRPCService = trcnapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// TarCoin protocol.
func (s *TarCoin) Stop() error {
	// Stop all the peer-related stuff first.
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}

	// Then stop everything else.
	s.bloomIndexer.Close()
	close(s.closeBloomHandler)
	s.txPool.Stop()
	s.miner.Stop()
	s.blockchain.Stop()
	s.engine.Close()
	s.chainDb.Close()
	s.eventMux.Stop()
	return nil
}
