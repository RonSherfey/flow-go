// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package ingestion

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/flow/filter"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/mempool"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/module/trace"
	"github.com/onflow/flow-go/network"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/onflow/flow-go/storage"
)

// Engine represents the ingestion engine, used to funnel collections from a
// cluster of collection nodes to the set of consensus nodes. It represents the
// link between collection nodes and consensus nodes and has a counterpart with
// the same engine ID in the collection node.
type Engine struct {
	unit    *engine.Unit            // used to manage concurrency & shutdown
	log     zerolog.Logger          // used to log relevant actions with context
	tracer  module.Tracer           // used for tracing
	metrics module.EngineMetrics    // used to track sent & received messages
	mempool module.MempoolMetrics   // used to track mempool metrics
	spans   module.ConsensusMetrics // used to track consensus spans
	state   protocol.State          // used to access the protocol state
	headers storage.Headers         // used to retrieve headers
	me      module.Local            // used to access local node information
	pool    mempool.Guarantees      // used to keep pending guarantees in pool
	con     network.Conduit         // conduit to receive/send guarantees
}

// New creates a new collection propagation engine.
func New(
	log zerolog.Logger,
	tracer module.Tracer,
	metrics module.EngineMetrics,
	spans module.ConsensusMetrics,
	mempool module.MempoolMetrics,
	net module.Network,
	state protocol.State,
	headers storage.Headers,
	me module.Local,
	pool mempool.Guarantees,
) (*Engine, error) {

	// initialize the propagation engine with its dependencies
	e := &Engine{
		unit:    engine.NewUnit(),
		log:     log.With().Str("engine", "ingestion").Logger(),
		tracer:  tracer,
		metrics: metrics,
		mempool: mempool,
		spans:   spans,
		state:   state,
		headers: headers,
		me:      me,
		pool:    pool,
	}

	// register the engine with the network layer and store the conduit
	con, err := net.Register(engine.ReceiveGuarantees, e)
	if err != nil {
		return nil, fmt.Errorf("could not register engine: %w", err)
	}
	e.con = con

	return e, nil
}

// Ready returns a ready channel that is closed once the engine has fully
// started. For the ingestion engine, we consider the engine up and running
// upon initialization.
func (e *Engine) Ready() <-chan struct{} {
	return e.unit.Ready()
}

// Done returns a done channel that is closed once the engine has fully stopped.
// For the ingestion engine, it only waits for all submit goroutines to end.
func (e *Engine) Done() <-chan struct{} {
	return e.unit.Done()
}

// SubmitLocal submits an event originating on the local node.
func (e *Engine) SubmitLocal(event interface{}) {
	e.Submit(e.me.NodeID(), event)
}

// Submit submits the given event from the node with the given origin ID
// for processing in a non-blocking manner. It returns instantly and logs
// a potential processing error internally when done.
func (e *Engine) Submit(originID flow.Identifier, event interface{}) {
	e.unit.Launch(func() {
		err := e.process(originID, event)
		engine.LogError(e.log, err)
	})
}

// ProcessLocal processes an event originating on the local node.
func (e *Engine) ProcessLocal(event interface{}) error {
	return e.Process(e.me.NodeID(), event)
}

// Process processes the given event from the node with the given origin ID in
// a blocking manner. It returns the potential processing error when done.
func (e *Engine) Process(originID flow.Identifier, event interface{}) error {
	return e.unit.Do(func() error {
		return e.process(originID, event)
	})
}

// process processes the given ingestion engine event. Events that are given
// to this function originate within the expulsion engine on the node with the
// given origin ID.
func (e *Engine) process(originID flow.Identifier, event interface{}) error {
	switch ev := event.(type) {
	case *flow.CollectionGuarantee:
		e.metrics.MessageReceived(metrics.EngineConsensusIngestion, metrics.MessageCollectionGuarantee)
		return e.onGuarantee(originID, ev)
	default:
		return fmt.Errorf("invalid event type (%T)", event)
	}
}

// onGuarantee is used to process collection guarantees received
// from nodes that are not consensus nodes (notably collection nodes).
// Returns expected errors:
// * InvalidInputError
// * UnverifiableInput
// * OutdatedInputError
func (e *Engine) onGuarantee(originID flow.Identifier, guarantee *flow.CollectionGuarantee) error {

	span := e.tracer.StartSpan(guarantee.CollectionID, trace.CONProcessCollection)
	// TODO finish span if we error? How are they shown in Jaeger?
	span.SetTag("collection_id", guarantee.CollectionID)
	childSpan := e.tracer.StartSpanFromParent(span, trace.CONIngOnCollectionGuarantee)
	defer childSpan.Finish()

	guaranteeID := guarantee.ID()

	log := e.log.With().
		Hex("origin_id", originID[:]).
		Hex("collection_id", guaranteeID[:]).
		Int("signers", len(guarantee.SignerIDs)).
		Logger()

	log.Info().Msg("collection guarantee received")

	// skip collection guarantees that are already in our memory pool
	exists := e.pool.Has(guaranteeID)
	if exists {
		log.Debug().Msg("skipping known collection guarantee")
		return nil
	}

	// retrieve and validate the sender of the guarantee message
	origin, err := e.validateOrigin(originID, guarantee)
	if err != nil {
		return fmt.Errorf("origin validation error: %w", err)
	}

	// ensure that collection has not expired
	err = e.validateExpiry(guarantee)
	if err != nil {
		return fmt.Errorf("expiry validation error: %w", err)
	}

	// ensure the guarantors are allowed to produce this collection
	err = e.validateGuarantors(guarantee)
	if err != nil {
		return fmt.Errorf("guarantor validation error: %w", err)
	}

	// at this point, we can add the guarantee to the memory pool
	added := e.pool.Add(guarantee)
	if !added {
		log.Debug().Msg("discarding guarantee already in pool")
		return nil
	}

	e.mempool.MempoolEntries(metrics.ResourceGuarantee, e.pool.Size())

	log.Info().Msg("collection guarantee added to pool")

	// if the collection guarantee stems from a collection node, we should propagate it
	// to the other consensus nodes on the network; we no longer need to care about
	// fan-out here, as libp2p will take care of the pub-sub pattern adequately
	if origin.Role != flow.RoleCollection {
		return nil
	}

	// NOTE: there are two ways to go about this:
	// - expect the collection nodes to propagate the guarantee to all consensus nodes;
	// - ensure that we take care of propagating guarantees to other consensus nodes.
	// It's probably a better idea to make sure the consensus nodes take care of this.
	// The consensus committee is the backbone of the network and should rely as little
	// as possible on correct behaviour from other node roles. At the same time, there
	// are likely to be significantly more consensus nodes on the network, which means
	// it's a better usage of resources to distribute the load for propagation over
	// consensus node committee than over the collection clusters.

	// select all the consensus nodes on the network as our targets
	committee, err := e.state.Final().Identities(filter.HasRole(flow.RoleConsensus))
	if err != nil {
		return fmt.Errorf("could not get committee: %w", err)
	}

	// send the collection guarantee to all consensus committee
	err = e.con.Publish(guarantee, committee.NodeIDs()...)
	if err != nil {
		return fmt.Errorf("could not send guarantee: %w", err)
	}

	log.Info().Msg("collection guarantee broadcast to committee")

	return nil
}

// validateExpiry validates that the collection has not expired w.r.t. the local
// latest finalized block.
func (e *Engine) validateExpiry(guarantee *flow.CollectionGuarantee) error {

	// get the last finalized header and the reference block header
	final, err := e.state.Final().Head()
	if err != nil {
		return fmt.Errorf("could not get finalized header: %w", err)
	}
	ref, err := e.headers.ByBlockID(guarantee.ReferenceBlockID)
	if errors.Is(err, storage.ErrNotFound) {
		return engine.NewUnverifiableInputError("collection guarantee refers to an unknown block (id=%x): %w", guarantee.ReferenceBlockID, err)
	}

	// if head has advanced beyond the block referenced by the collection guarantee by more than 'expiry' number of blocks,
	// then reject the collection
	if ref.Height > final.Height {
		return nil // the reference block is newer than the latest finalized one
	}
	if final.Height-ref.Height > flow.DefaultTransactionExpiry {
		return engine.NewOutdatedInputErrorf("collection guarantee expired ref_height=%d final_height=%d", ref.Height, final.Height)
	}

	return nil
}

// validateGuarantors validates that the guarantors of a collection are valid,
// in that they are all from the same cluster and that cluster is allowed to
// produce the given collection w.r.t. the guarantee's reference block.
func (e *Engine) validateGuarantors(guarantee *flow.CollectionGuarantee) error {

	guarantors := guarantee.SignerIDs

	if len(guarantors) == 0 {
		return engine.NewInvalidInputError("invalid collection guarantee with no guarantors")
	}

	// get the clusters to assign the guarantee and check if the guarantor is part of it
	snapshot := e.state.AtBlockID(guarantee.ReferenceBlockID)
	clusters, err := snapshot.Epochs().Current().Clustering()
	if errors.Is(err, storage.ErrNotFound) {
		return engine.NewUnverifiableInputError("could not get clusters for unknown reference block (id=%x): %w", guarantee.ReferenceBlockID, err)
	}
	if err != nil {
		return fmt.Errorf("internal error retrieving collector clusters: %w", err)
	}
	cluster, _, ok := clusters.ByNodeID(guarantors[0])
	if !ok {
		return engine.NewInvalidInputErrorf("guarantor (id=%s) does not exist in any cluster", guarantors[0])
	}

	// NOTE: Eventually we should check the signatures, ensure a quorum of the
	// cluster, and ensure HotStuff finalization rules. Likely a cluster-specific
	// version of the follower will be a good fit for this. For now, collection
	// nodes independently decide when a collection is finalized and we only check
	// that the guarantors are all from the same cluster.

	// ensure the guarantors are from the same cluster
	clusterLookup := cluster.Lookup()
	for _, guarantorID := range guarantors {
		_, exists := clusterLookup[guarantorID]
		if !exists {
			return engine.NewInvalidInputError("inconsistent guarantors from different clusters")
		}
	}

	return nil
}

// validateOrigin validates that the message has a valid sender (origin). We
// allow Guarantee messages either from collection nodes or from consensus nodes.
// Collection nodes must be valid in the identity table w.r.t. the reference
// block of the collection; consensus nodes must be valid w.r.t. the latest
// finalized block.
//
// Returns the origin identity as validated if validation passes.
func (e *Engine) validateOrigin(originID flow.Identifier, guarantee *flow.CollectionGuarantee) (*flow.Identity, error) {

	refBlock := guarantee.ReferenceBlockID

	// get the origin identity w.r.t. the collection reference block
	origin, err := e.state.AtBlockID(refBlock).Identity(originID)
	if err != nil {
		// collection with an unknown reference block is unverifiable
		if errors.Is(err, storage.ErrNotFound) {
			return nil, engine.NewUnverifiableInputError("could not get origin (id=%x) for unknown reference block (id=%x): %w", originID, refBlock, err)
		}
		// in case of identity not found, we continue and check the other case - all other errors are unexpected
		if !protocol.IsIdentityNotFound(err) {
			return nil, fmt.Errorf("could not get origin (id=%x) identity w.r.t. reference block (id=%x) : %w", originID, refBlock, err)
		}
	} else {
		// if it is a collection node and a valid epoch participant, it is a valid origin
		if filter.And(
			filter.IsValidCurrentEpochParticipant,
			filter.HasRole(flow.RoleCollection),
		)(origin) {
			return origin, nil
		}
	}

	// At this point there are two cases:
	// 1) origin was not found at reference block state
	// 2) origin was found at reference block state, but isn't a collection node
	//
	// Now, check whether the origin is a valid consensus node using finalized state.
	// This can happen when we have just passed an epoch boundary and a consensus
	// node which has just joined sends us a guarantee referencing a block prior
	// to the epoch boundary.

	// get the origin identity w.r.t. the latest finalized block
	origin, err = e.state.Final().Identity(originID)
	if protocol.IsIdentityNotFound(err) {
		return nil, engine.NewInvalidInputErrorf("unknown origin (id=%x): %w", originID, err)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get origin (id=%x) w.r.t. finalized state: %w", originID, err)
	}

	if filter.And(
		filter.IsValidCurrentEpochParticipant,
		filter.HasRole(flow.RoleConsensus),
	)(origin) {
		return origin, nil
	}

	return nil, engine.NewInvalidInputErrorf("invalid origin (id=%x)", originID)
}
