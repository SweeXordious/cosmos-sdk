package smt

import (
	"crypto/sha256"
	"io"
	"sync"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/lazyledger/smt"
)

var (
	_ types.KVStore                 = (*Store)(nil)
	_ types.CommitStore             = (*Store)(nil)
	_ types.CommitKVStore           = (*Store)(nil)
	_ types.Queryable               = (*Store)(nil)
	_ types.StoreWithInitialVersion = (*Store)(nil)
)

// Store Implements types.KVStore and CommitKVStore.
type Store struct {
	tree *smt.SparseMerkleTree
	db   dbm.DB

	mtx sync.RWMutex
}

func NewStore(underlyingDB dbm.DB) *Store {
	return &Store{
		tree: smt.NewSparseMerkleTree(underlyingDB, sha256.New()),
		db:   underlyingDB,
	}
}

// KVStore interface below:

func (s *Store) GetStoreType() types.StoreType {
	return types.StoreTypeSMT
}

// CacheWrap branches a store.
func (s *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(s)
}

// CacheWrapWithTrace branches a store with tracing enabled.
func (s *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

// Get returns nil iff key doesn't exist. Panics on nil key.
func (s *Store) Get(key []byte) []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	val, err := s.tree.Get(key)
	// TODO(tzdybal): how to handle this err?
	if err != nil {
		return nil
	}
	return val
}

// Has checks if a key exists. Panics on nil key.
func (s *Store) Has(key []byte) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	has, err := s.db.Has(indexKey(key))
	return err == nil && has
}

// Set sets the key. Panics on nil key or value.
func (s *Store) Set(key []byte, value []byte) {
	s.mtx.Lock()
	// TODO(tzdybal): how to handle this err?
	// TODO(tzdybal): it should be done as batch, but probably using mutex is easier
	_, err := s.tree.Update(key, value)
	if err != nil {
		panic(err.Error())
	}
	err = s.db.Set(indexKey(key), []byte{})
	if err != nil {
		panic(err.Error())
	}
	s.mtx.Unlock()
}

// Delete deletes the key. Panics on nil key.
func (s *Store) Delete(key []byte) {
	s.mtx.Lock()
	_, _ = s.tree.Delete(key)
	_ = s.db.Delete(indexKey(key))
	s.mtx.Unlock()
}

// Iterator over a domain of keys in ascending order. End is exclusive.
// Start must be less than end, or the Iterator is invalid.
// Iterator must be closed by caller.
// To iterate over entire domain, use store.Iterator(nil, nil)
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// Exceptionally allowed for cachekv.Store, safe to write in the modules.
func (s *Store) Iterator(start []byte, end []byte) types.Iterator {
	iter, err := newIterator(s, start, end, false)
	if err != nil {
		panic(err.Error())
	}
	return iter
}

// Iterator over a domain of keys in descending order. End is exclusive.
// Start must be less than end, or the Iterator is invalid.
// Iterator must be closed by caller.
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// Exceptionally allowed for cachekv.Store, safe to write in the modules.
func (s *Store) ReverseIterator(start []byte, end []byte) types.Iterator {
	iter, err := newIterator(s, start, end, true)
	if err != nil {
		panic(err.Error())
	}
	return iter
}

// CommitStore interface below:

func (s *Store) Commit() types.CommitID {
	panic("not implemented") // TODO(tzdybal): Implement
}

func (s *Store) LastCommitID() types.CommitID {
	panic("not implemented") // TODO(tzdybal): Implement
}

func (s *Store) SetPruning(_ types.PruningOptions) {
	panic("not implemented") // TODO(tzdybal): Implement
}

func (s *Store) GetPruning() types.PruningOptions {
	panic("not implemented") // TODO(tzdybal): Implement
}

// Queryable interface below:

func (s *Store) Query(_ abci.RequestQuery) abci.ResponseQuery {
	panic("not implemented") // TODO(tzdybal): Implement
}

// StoreWithInitialVersion interface below:

// SetInitialVersion sets the initial version of the IAVL tree. It is used when
// starting a new chain at an arbitrary height.
func (s *Store) SetInitialVersion(version int64) {
	panic("not implemented") // TODO(tzdybal): Implement
}
