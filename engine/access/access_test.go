package access

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	accessproto "github.com/onflow/flow/protobuf/go/flow/access"
	entitiesproto "github.com/onflow/flow/protobuf/go/flow/entities"
	execproto "github.com/onflow/flow/protobuf/go/flow/execution"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/onflow/flow-go/access"
	"github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/crypto"
	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/engine/access/ingestion"
	accessmock "github.com/onflow/flow-go/engine/access/mock"
	"github.com/onflow/flow-go/engine/access/rpc"
	"github.com/onflow/flow-go/engine/access/rpc/backend"
	factorymock "github.com/onflow/flow-go/engine/access/rpc/backend/mock"
	"github.com/onflow/flow-go/engine/common/rpc/convert"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/mempool/stdmap"
	"github.com/onflow/flow-go/module/metrics"
	module "github.com/onflow/flow-go/module/mock"
	"github.com/onflow/flow-go/network/mocknetwork"
	protocol "github.com/onflow/flow-go/state/protocol/mock"
	storage "github.com/onflow/flow-go/storage/badger"
	"github.com/onflow/flow-go/storage/badger/operation"
	"github.com/onflow/flow-go/storage/util"
	"github.com/onflow/flow-go/utils/unittest"
)

type Suite struct {
	suite.Suite
	state      *protocol.State
	snapshot   *protocol.Snapshot
	epochQuery *protocol.EpochQuery
	log        zerolog.Logger
	net        *module.Network
	request    *module.Requester
	collClient *accessmock.AccessAPIClient
	execClient *accessmock.ExecutionAPIClient
	me         *module.Local
	chainID    flow.ChainID
	metrics    *metrics.NoopCollector
	backend    *backend.Backend
}

// TestAccess tests scenarios which exercise multiple API calls using both the RPC handler and the ingest engine
// and using a real badger storage
func TestAccess(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (suite *Suite) SetupTest() {
	suite.log = zerolog.New(os.Stderr)
	suite.net = new(module.Network)
	suite.state = new(protocol.State)
	suite.snapshot = new(protocol.Snapshot)

	suite.epochQuery = new(protocol.EpochQuery)
	suite.state.On("Sealed").Return(suite.snapshot, nil).Maybe()
	suite.state.On("Final").Return(suite.snapshot, nil).Maybe()
	suite.snapshot.On("Epochs").Return(suite.epochQuery).Maybe()

	suite.collClient = new(accessmock.AccessAPIClient)
	suite.execClient = new(accessmock.ExecutionAPIClient)

	suite.request = new(module.Requester)
	suite.request.On("EntityByID", mock.Anything, mock.Anything)

	suite.me = new(module.Local)

	accessIdentity := unittest.IdentityFixture(unittest.WithRole(flow.RoleAccess))
	suite.me.
		On("NodeID").
		Return(accessIdentity.NodeID)

	suite.chainID = flow.Testnet
	suite.metrics = metrics.NewNoopCollector()
}

func (suite *Suite) RunTest(
	f func(handler *access.Handler, db *badger.DB, blocks *storage.Blocks, headers *storage.Headers),
) {
	unittest.RunWithBadgerDB(suite.T(), func(db *badger.DB) {
		headers, _, _, _, _, blocks, _, _, _, _ := util.StorageLayer(suite.T(), db)
		transactions := storage.NewTransactions(suite.metrics, db)
		collections := storage.NewCollections(db, transactions)
		results := storage.NewExecutionResults(suite.metrics, db)
		receipts := storage.NewExecutionReceipts(suite.metrics, db, results)

		suite.backend = backend.New(
			suite.state,
			suite.execClient,
			suite.collClient,
			nil,
			blocks,
			headers,
			collections,
			transactions,
			receipts,
			suite.chainID,
			suite.metrics,
			nil,
			false,
			backend.DefaultMaxHeightRange,
			nil,
			nil,
			suite.log,
		)

		handler := access.NewHandler(suite.backend, suite.chainID.Chain())

		f(handler, db, blocks, headers)
	})
}

func (suite *Suite) TestSendAndGetTransaction() {
	suite.RunTest(func(handler *access.Handler, _ *badger.DB, _ *storage.Blocks, _ *storage.Headers) {
		referenceBlock := unittest.BlockHeaderFixture()
		transaction := unittest.TransactionFixture()
		transaction.SetReferenceBlockID(referenceBlock.ID())

		refSnapshot := new(protocol.Snapshot)

		suite.state.
			On("AtBlockID", referenceBlock.ID()).
			Return(refSnapshot, nil)

		refSnapshot.
			On("Head").
			Return(&referenceBlock, nil).
			Twice()

		suite.snapshot.
			On("Head").
			Return(&referenceBlock, nil).
			Once()

		expected := convert.TransactionToMessage(transaction.TransactionBody)
		sendReq := &accessproto.SendTransactionRequest{
			Transaction: expected,
		}
		sendResp := accessproto.SendTransactionResponse{}

		suite.collClient.
			On("SendTransaction", mock.Anything, mock.Anything).
			Return(&sendResp, nil).
			Once()

		// Send transaction
		resp, err := handler.SendTransaction(context.Background(), sendReq)
		suite.Require().NoError(err)
		suite.Require().NotNil(resp)

		id := transaction.ID()
		getReq := &accessproto.GetTransactionRequest{
			Id: id[:],
		}

		// Get transaction
		gResp, err := handler.GetTransaction(context.Background(), getReq)
		suite.Require().NoError(err)
		suite.Require().NotNil(gResp)

		actual := gResp.GetTransaction()
		suite.Require().Equal(expected, actual)
	})
}

func (suite *Suite) TestSendExpiredTransaction() {
	suite.RunTest(func(handler *access.Handler, _ *badger.DB, _ *storage.Blocks, _ *storage.Headers) {
		referenceBlock := unittest.BlockHeaderFixture()

		// create latest block that is past the expiry window
		latestBlock := unittest.BlockHeaderFixture()
		latestBlock.Height = referenceBlock.Height + flow.DefaultTransactionExpiry*2

		transaction := unittest.TransactionFixture()
		transaction.SetReferenceBlockID(referenceBlock.ID())

		refSnapshot := new(protocol.Snapshot)

		suite.state.
			On("AtBlockID", referenceBlock.ID()).
			Return(refSnapshot, nil)

		refSnapshot.
			On("Head").
			Return(&referenceBlock, nil).
			Twice()

		suite.snapshot.
			On("Head").
			Return(&latestBlock, nil).
			Once()

		req := &accessproto.SendTransactionRequest{
			Transaction: convert.TransactionToMessage(transaction.TransactionBody),
		}

		_, err := handler.SendTransaction(context.Background(), req)
		suite.Require().Error(err)
	})
}

type mockCloser struct{}

func (mc *mockCloser) Close() error { return nil }

// TestSendTransactionToRandomCollectionNode tests that collection nodes are chosen from the appropriate cluster when
// forwarding transactions by sending two transactions bound for two different collection clusters.
func (suite *Suite) TestSendTransactionToRandomCollectionNode() {
	unittest.RunWithBadgerDB(suite.T(), func(db *badger.DB) {

		// create a transaction
		referenceBlock := unittest.BlockHeaderFixture()
		transaction := unittest.TransactionFixture()
		transaction.SetReferenceBlockID(referenceBlock.ID())

		// setup the state and snapshot mock expectations
		suite.state.On("AtBlockID", referenceBlock.ID()).Return(suite.snapshot, nil)
		suite.snapshot.On("Head").Return(&referenceBlock, nil)

		// create storage
		metrics := metrics.NewNoopCollector()
		transactions := storage.NewTransactions(metrics, db)
		collections := storage.NewCollections(db, transactions)

		// create collection node cluster
		count := 2
		collNodes := unittest.IdentityListFixture(count, unittest.WithRole(flow.RoleCollection))
		assignments := unittest.ClusterAssignment(uint(count), collNodes)
		clusters, err := flow.NewClusterList(assignments, collNodes)
		suite.Require().Nil(err)
		collNode1 := clusters[0][0]
		collNode2 := clusters[1][0]
		epoch := new(protocol.Epoch)
		suite.epochQuery.On("Current").Return(epoch)
		epoch.On("Clustering").Return(clusters, nil)

		// create two transactions bound for each of the cluster
		cluster1 := clusters[0]
		cluster1tx := unittest.AlterTransactionForCluster(transaction.TransactionBody, clusters, cluster1, func(transaction *flow.TransactionBody) {})
		tx1 := convert.TransactionToMessage(cluster1tx)
		sendReq1 := &accessproto.SendTransactionRequest{
			Transaction: tx1,
		}
		cluster2 := clusters[1]
		cluster2tx := unittest.AlterTransactionForCluster(transaction.TransactionBody, clusters, cluster2, func(transaction *flow.TransactionBody) {})
		tx2 := convert.TransactionToMessage(cluster2tx)
		sendReq2 := &accessproto.SendTransactionRequest{
			Transaction: tx2,
		}
		sendResp := accessproto.SendTransactionResponse{}

		// create mock access api clients for each of the collection node expecting the correct transaction once
		col1ApiClient := new(accessmock.AccessAPIClient)
		col1ApiClient.On("SendTransaction", mock.Anything, sendReq1).Return(&sendResp, nil).Once()
		col2ApiClient := new(accessmock.AccessAPIClient)
		col2ApiClient.On("SendTransaction", mock.Anything, sendReq2).Return(&sendResp, nil).Once()

		// create a mock connection factory
		connFactory := new(factorymock.ConnectionFactory)
		connFactory.On("GetAccessAPIClient", collNode1.Address).Return(col1ApiClient, &mockCloser{}, nil)
		connFactory.On("GetAccessAPIClient", collNode2.Address).Return(col2ApiClient, &mockCloser{}, nil)

		backend := backend.New(
			suite.state,
			nil,
			nil, // setting collectionRPC to nil to choose a random collection node for each send tx request
			nil,
			nil,
			nil,
			collections,
			transactions,
			nil,
			suite.chainID,
			metrics,
			connFactory, // passing in the connection factory
			false,
			backend.DefaultMaxHeightRange,
			nil,
			nil,
			suite.log,
		)

		handler := access.NewHandler(backend, suite.chainID.Chain())

		// Send transaction 1
		resp, err := handler.SendTransaction(context.Background(), sendReq1)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), resp)

		// Send transaction 2
		resp, err = handler.SendTransaction(context.Background(), sendReq2)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), resp)

		// verify that a collection node in the correct cluster was contacted exactly once
		col1ApiClient.AssertExpectations(suite.T())
		col2ApiClient.AssertExpectations(suite.T())
		epoch.AssertNumberOfCalls(suite.T(), "Clustering", 2)

		// additionally do a GetTransaction request for the two transactions
		getTx := func(tx flow.TransactionBody) {
			id := tx.ID()
			getReq := &accessproto.GetTransactionRequest{
				Id: id[:],
			}
			gResp, err := handler.GetTransaction(context.Background(), getReq)
			require.NoError(suite.T(), err)
			require.NotNil(suite.T(), gResp)
			actual := gResp.GetTransaction()
			expected := convert.TransactionToMessage(tx)
			require.Equal(suite.T(), expected, actual)
		}

		getTx(cluster1tx)
		getTx(cluster1tx)
	})
}

func (suite *Suite) TestGetBlockByIDAndHeight() {
	suite.RunTest(func(handler *access.Handler, db *badger.DB, blocks *storage.Blocks, _ *storage.Headers) {

		// test block1 get by ID
		block1 := unittest.BlockFixture()
		// test block2 get by height
		block2 := unittest.BlockFixture()
		block2.Header.Height = 2

		require.NoError(suite.T(), blocks.Store(&block1))
		require.NoError(suite.T(), blocks.Store(&block2))

		// the follower logic should update height index on the block storage when a block is finalized
		err := db.Update(operation.IndexBlockHeight(block2.Header.Height, block2.ID()))
		require.NoError(suite.T(), err)

		assertHeaderResp := func(resp *accessproto.BlockHeaderResponse, err error, header *flow.Header) {
			require.NoError(suite.T(), err)
			require.NotNil(suite.T(), resp)
			actual := *resp.Block
			expected, _ := convert.BlockHeaderToMessage(header)
			require.Equal(suite.T(), *expected, actual)
		}

		assertBlockResp := func(resp *accessproto.BlockResponse, err error, block *flow.Block) {
			require.NoError(suite.T(), err)
			require.NotNil(suite.T(), resp)
			actual := resp.Block
			expected, _ := convert.BlockToMessage(block)
			require.Equal(suite.T(), expected, actual)
		}

		suite.Run("get header 1 by ID", func() {
			// get header by ID
			id := block1.ID()
			req := &accessproto.GetBlockHeaderByIDRequest{
				Id: id[:],
			}

			resp, err := handler.GetBlockHeaderByID(context.Background(), req)

			// assert it is indeed block1
			assertHeaderResp(resp, err, block1.Header)
		})

		suite.Run("get block 1 by ID", func() {
			id := block1.ID()
			// get block details by ID
			req := &accessproto.GetBlockByIDRequest{
				Id: id[:],
			}

			resp, err := handler.GetBlockByID(context.Background(), req)

			assertBlockResp(resp, err, &block1)
		})

		suite.Run("get header 2 by height", func() {

			// get header by height
			req := &accessproto.GetBlockHeaderByHeightRequest{
				Height: block2.Header.Height,
			}

			resp, err := handler.GetBlockHeaderByHeight(context.Background(), req)

			assertHeaderResp(resp, err, block2.Header)
		})

		suite.Run("get block 2 by height", func() {
			// get block details by height
			req := &accessproto.GetBlockByHeightRequest{
				Height: block2.Header.Height,
			}

			resp, err := handler.GetBlockByHeight(context.Background(), req)

			assertBlockResp(resp, err, &block2)
		})
	})
}

// TestGetSealedTransaction tests that transactions status of transaction that belongs to a sealed blocked
// is reported as sealed
func (suite *Suite) TestGetSealedTransaction() {
	suite.RunTest(func(handler *access.Handler, db *badger.DB, blocks *storage.Blocks, headers *storage.Headers) {

		// create block -> collection -> transactions
		block, collection := suite.createChain()

		// setup mocks
		originID := unittest.IdentifierFixture()
		conduit := new(mocknetwork.Conduit)
		suite.net.On("Register", engine.ReceiveReceipts, mock.Anything).Return(conduit, nil).
			Once()
		suite.request.On("Request", mock.Anything, mock.Anything).Return()

		suite.state.On("Sealed").Return(suite.snapshot, nil).Maybe()

		colIdentities := unittest.IdentityListFixture(1, unittest.WithRole(flow.RoleCollection))
		suite.snapshot.On("Identities", mock.Anything).Return(colIdentities, nil).Once()

		exeEventResp := execproto.GetTransactionResultResponse{
			Events: nil,
		}
		// assume execution node returns an empty list of events
		suite.execClient.On("GetTransactionResult", mock.Anything, mock.Anything).Return(&exeEventResp, nil)

		// initialize storage
		metrics := metrics.NewNoopCollector()
		transactions := storage.NewTransactions(metrics, db)
		collections := storage.NewCollections(db, transactions)
		collectionsToMarkFinalized, err := stdmap.NewTimes(100)
		require.NoError(suite.T(), err)
		collectionsToMarkExecuted, err := stdmap.NewTimes(100)
		require.NoError(suite.T(), err)
		blocksToMarkExecuted, err := stdmap.NewTimes(100)
		require.NoError(suite.T(), err)

		rpcEng := rpc.New(suite.log, suite.state, rpc.Config{}, nil, nil, nil, blocks, headers, collections, transactions,
			nil, suite.chainID, metrics, 0, 0, false, false, nil, nil)

		// create the ingest engine
		ingestEng, err := ingestion.New(suite.log, suite.net, suite.state, suite.me, suite.request, blocks, headers, collections,
			transactions, nil, metrics, collectionsToMarkFinalized, collectionsToMarkExecuted, blocksToMarkExecuted, rpcEng)
		require.NoError(suite.T(), err)

		// 1. Assume that follower engine updated the block storage and the protocol state. The block is reported as sealed
		err = blocks.Store(&block)
		require.NoError(suite.T(), err)
		suite.snapshot.On("Head").Return(block.Header, nil).Once()

		// 2. Ingest engine was notified by the follower engine about a new block.
		// Follower engine --> Ingest engine
		mb := &model.Block{
			BlockID: block.ID(),
		}
		ingestEng.OnFinalizedBlock(mb)

		// 3. Request engine is used to request missing collection
		suite.request.On("EntityByID", collection.ID(), mock.Anything).Return()

		// 4. Ingest engine receives the requested collection and finishes processing
		ingestEng.OnCollection(originID, &collection)
		<-ingestEng.Done()

		// 5. client requests a transaction
		tx := collection.Transactions[0]
		txID := tx.ID()
		getReq := &accessproto.GetTransactionRequest{
			Id: txID[:],
		}
		gResp, err := handler.GetTransactionResult(context.Background(), getReq)
		require.NoError(suite.T(), err)
		// assert that the transaction is reported as Sealed
		require.Equal(suite.T(), entitiesproto.TransactionStatus_SEALED, gResp.GetStatus())
	})
}

// TestExecuteScript tests the three execute Script related calls to make sure that the execution api is called with
// the correct block id
func (suite *Suite) TestExecuteScript() {
	unittest.RunWithBadgerDB(suite.T(), func(db *badger.DB) {
		headers, _, _, _, _, blocks, _, _, _, _ := util.StorageLayer(suite.T(), db)
		transactions := storage.NewTransactions(suite.metrics, db)
		collections := storage.NewCollections(db, transactions)
		results := storage.NewExecutionResults(suite.metrics, db)
		receipts := storage.NewExecutionReceipts(suite.metrics, db, results)

		identities := unittest.IdentityListFixture(1)
		suite.snapshot.On("Identities", mock.Anything).Return(identities, nil)
		executionNodeIdentity := identities[0]

		// create a mock connection factory
		connFactory := new(factorymock.ConnectionFactory)
		connFactory.On("GetExecutionAPIClient", executionNodeIdentity.Address).Return(suite.execClient, &mockCloser{}, nil)

		suite.backend = backend.New(
			suite.state,
			suite.execClient,
			suite.collClient,
			nil,
			blocks,
			headers,
			collections,
			transactions,
			receipts,
			suite.chainID,
			suite.metrics,
			connFactory,
			false,
			backend.DefaultMaxHeightRange,
			nil,
			nil,
			suite.log,
		)

		handler := access.NewHandler(suite.backend, suite.chainID.Chain())

		// initialize metrics related storage
		metrics := metrics.NewNoopCollector()
		collectionsToMarkFinalized, err := stdmap.NewTimes(100)
		require.NoError(suite.T(), err)
		collectionsToMarkExecuted, err := stdmap.NewTimes(100)
		require.NoError(suite.T(), err)
		blocksToMarkExecuted, err := stdmap.NewTimes(100)
		require.NoError(suite.T(), err)

		conduit := new(mocknetwork.Conduit)
		suite.net.On("Register", engine.ReceiveReceipts, mock.Anything).Return(conduit, nil).
			Once()
		// create the ingest engine
		ingestEng, err := ingestion.New(suite.log, suite.net, suite.state, suite.me, suite.request, blocks, headers, collections,
			transactions, receipts, metrics, collectionsToMarkFinalized, collectionsToMarkExecuted, blocksToMarkExecuted, nil)
		require.NoError(suite.T(), err)

		// create a block and a seal pointing to that block
		lastBlock := unittest.BlockFixture()
		lastBlock.Header.Height = 2
		err = blocks.Store(&lastBlock)
		require.NoError(suite.T(), err)
		err = db.Update(operation.IndexBlockHeight(lastBlock.Header.Height, lastBlock.ID()))
		require.NoError(suite.T(), err)
		suite.snapshot.On("Head").Return(lastBlock.Header, nil).Once()

		// create an execution receipt for the block
		executionReceiptLast := unittest.ReceiptForBlockFixture(&lastBlock)
		executionReceiptLast.ExecutorID = identities[0].NodeID

		// notify the ingestion engine about the receipt
		err = ingestEng.ProcessLocal(executionReceiptLast)
		require.NoError(suite.T(), err)

		// create another block as a predecessor of the block created earlier
		prevBlock := unittest.BlockFixture()
		prevBlock.Header.Height = lastBlock.Header.Height - 1
		err = blocks.Store(&prevBlock)
		require.NoError(suite.T(), err)
		err = db.Update(operation.IndexBlockHeight(prevBlock.Header.Height, prevBlock.ID()))
		require.NoError(suite.T(), err)

		// create an execution receipt for the block
		executionReceiptPrev := unittest.ReceiptForBlockFixture(&prevBlock)
		executionReceiptPrev.ExecutorID = identities[0].NodeID

		// notify the ingestion engine about the receipt
		err = ingestEng.ProcessLocal(executionReceiptPrev)
		require.NoError(suite.T(), err)
		// wait for the receipt to be persisted
		unittest.AssertClosesBefore(suite.T(), ingestEng.Done(), 3*time.Second)
		ctx := context.Background()

		script := []byte("dummy script")

		// setupExecClientMock sets up the mock the execution client and returns the access response to expect
		setupExecClientMock := func(blockID flow.Identifier) *accessproto.ExecuteScriptResponse {
			id := blockID[:]
			executionReq := execproto.ExecuteScriptAtBlockIDRequest{
				BlockId: id,
				Script:  script,
			}
			executionResp := execproto.ExecuteScriptAtBlockIDResponse{
				Value: []byte{9, 10, 11},
			}

			suite.execClient.On("ExecuteScriptAtBlockID", ctx, &executionReq).Return(&executionResp, nil).Once()

			expectedResp := accessproto.ExecuteScriptResponse{
				Value: executionResp.GetValue(),
			}
			return &expectedResp
		}

		assertResult := func(err error, expected interface{}, actual interface{}) {
			suite.Require().NoError(err)
			suite.Require().Equal(expected, actual)
			suite.execClient.AssertExpectations(suite.T())
		}

		suite.Run("execute script at latest block", func() {
			suite.state.On("Sealed").Return(suite.snapshot, nil).Maybe()

			expectedResp := setupExecClientMock(lastBlock.ID())
			req := accessproto.ExecuteScriptAtLatestBlockRequest{
				Script: script,
			}
			actualResp, err := handler.ExecuteScriptAtLatestBlock(ctx, &req)
			assertResult(err, expectedResp, actualResp)
		})

		suite.Run("execute script at block id", func() {
			expectedResp := setupExecClientMock(prevBlock.ID())
			id := prevBlock.ID()
			req := accessproto.ExecuteScriptAtBlockIDRequest{
				BlockId: id[:],
				Script:  script,
			}
			actualResp, err := handler.ExecuteScriptAtBlockID(ctx, &req)
			assertResult(err, expectedResp, actualResp)
		})

		suite.Run("execute script at block height", func() {
			expectedResp := setupExecClientMock(prevBlock.ID())
			req := accessproto.ExecuteScriptAtBlockHeightRequest{
				BlockHeight: prevBlock.Header.Height,
				Script:      script,
			}
			actualResp, err := handler.ExecuteScriptAtBlockHeight(ctx, &req)
			assertResult(err, expectedResp, actualResp)
		})
	})
}

func (suite *Suite) createChain() (flow.Block, flow.Collection) {
	collection := unittest.CollectionFixture(10)
	guarantee := &flow.CollectionGuarantee{
		CollectionID: collection.ID(),
		Signature:    crypto.Signature([]byte("signature A")),
	}
	block := unittest.BlockFixture()
	block.Payload.Guarantees = []*flow.CollectionGuarantee{guarantee}
	block.Header.PayloadHash = block.Payload.Hash()

	return block, collection
}
