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

package rawdb

import (
	"github.com/speker/go-tarcoin/common"
	"github.com/speker/go-tarcoin/trcndb"
	"github.com/speker/go-tarcoin/log"
)

// ReadSnapshotRoot retrieves the root of the block whose state is contained in
// the persisted snapshot.
func ReadSnapshotRoot(db trcndb.KeyValueReader) common.Hash {
	data, _ := db.Get(snapshotRootKey)
	if len(data) != common.HashLength {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteSnapshotRoot stores the root of the block whose state is contained in
// the persisted snapshot.
func WriteSnapshotRoot(db trcndb.KeyValueWriter, root common.Hash) {
	if err := db.Put(snapshotRootKey, root[:]); err != nil {
		log.Crit("Failed to store snapshot root", "err", err)
	}
}

// DeleteSnapshotRoot deletes the hash of the block whose state is contained in
// the persisted snapshot. Since snapshots are not immutable, this  method can
// be used during updates, so a crash or failure will mark the entire snapshot
// invalid.
func DeleteSnapshotRoot(db trcndb.KeyValueWriter) {
	if err := db.Delete(snapshotRootKey); err != nil {
		log.Crit("Failed to remove snapshot root", "err", err)
	}
}

// ReadAccountSnapshot retrieves the snapshot entry of an account trie leaf.
func ReadAccountSnapshot(db trcndb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(accountSnapshotKey(hash))
	return data
}

// WriteAccountSnapshot stores the snapshot entry of an account trie leaf.
func WriteAccountSnapshot(db trcndb.KeyValueWriter, hash common.Hash, entry []byte) {
	if err := db.Put(accountSnapshotKey(hash), entry); err != nil {
		log.Crit("Failed to store account snapshot", "err", err)
	}
}

// DeleteAccountSnapshot removes the snapshot entry of an account trie leaf.
func DeleteAccountSnapshot(db trcndb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(accountSnapshotKey(hash)); err != nil {
		log.Crit("Failed to delete account snapshot", "err", err)
	}
}

// ReadStorageSnapshot retrieves the snapshot entry of an storage trie leaf.
func ReadStorageSnapshot(db trcndb.KeyValueReader, accountHash, storageHash common.Hash) []byte {
	data, _ := db.Get(storageSnapshotKey(accountHash, storageHash))
	return data
}

// WriteStorageSnapshot stores the snapshot entry of an storage trie leaf.
func WriteStorageSnapshot(db trcndb.KeyValueWriter, accountHash, storageHash common.Hash, entry []byte) {
	if err := db.Put(storageSnapshotKey(accountHash, storageHash), entry); err != nil {
		log.Crit("Failed to store storage snapshot", "err", err)
	}
}

// DeleteStorageSnapshot removes the snapshot entry of an storage trie leaf.
func DeleteStorageSnapshot(db trcndb.KeyValueWriter, accountHash, storageHash common.Hash) {
	if err := db.Delete(storageSnapshotKey(accountHash, storageHash)); err != nil {
		log.Crit("Failed to delete storage snapshot", "err", err)
	}
}

// IterateStorageSnapshots returns an iterator for walking the entire storage
// space of a specific account.
func IterateStorageSnapshots(db trcndb.Iteratee, accountHash common.Hash) trcndb.Iterator {
	return db.NewIterator(storageSnapshotsKey(accountHash), nil)
}

// ReadSnapshotJournal retrieves the serialized in-memory diff layers saved at
// the last shutdown. The blob is expected to be max a few 10s of megabytes.
func ReadSnapshotJournal(db trcndb.KeyValueReader) []byte {
	data, _ := db.Get(snapshotJournalKey)
	return data
}

// WriteSnapshotJournal stores the serialized in-memory diff layers to save at
// shutdown. The blob is expected to be max a few 10s of megabytes.
func WriteSnapshotJournal(db trcndb.KeyValueWriter, journal []byte) {
	if err := db.Put(snapshotJournalKey, journal); err != nil {
		log.Crit("Failed to store snapshot journal", "err", err)
	}
}

// DeleteSnapshotJournal deletes the serialized in-memory diff layers saved at
// the last shutdown
func DeleteSnapshotJournal(db trcndb.KeyValueWriter) {
	if err := db.Delete(snapshotJournalKey); err != nil {
		log.Crit("Failed to remove snapshot journal", "err", err)
	}
}
