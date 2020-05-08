package util

import (
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/require"

	protocol "github.com/dapperlabs/flow-go/state/protocol/badger"
	"github.com/dapperlabs/flow-go/storage/util"
	"github.com/dapperlabs/flow-go/utils/unittest"
)

func ProtocolState(t testing.TB, db *badger.DB, options ...func(*protocol.State)) *protocol.State {
	headers, identities, _, seals, payloads, blocks := util.StorageLayer(t, db)
	proto, err := protocol.NewState(db, headers, identities, seals, payloads, blocks, options...)
	require.NoError(t, err)
	return proto
}

func RunWithProtocolState(t testing.TB, f func(*badger.DB, *protocol.State), options ...func(*protocol.State)) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		proto := ProtocolState(t, db, options...)
		f(db, proto)
	})
}
