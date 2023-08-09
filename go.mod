module github.com/onflow/flow-go

go 1.13

require (
	cloud.google.com/go/storage v1.10.0
	github.com/HdrHistogram/hdrhistogram-go v0.9.0 // indirect
	github.com/bsipos/thist v1.0.0
	github.com/btcsuite/btcd v0.22.1
	github.com/codahale/hdrhistogram v0.9.0 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/ef-ds/deque v1.0.4
	github.com/ethereum/go-ethereum v1.9.13
	github.com/fxamacker/cbor/v2 v2.2.1-0.20201006223149-25f67fca9803
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2 v2.0.0-rc.2
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-20200501113911-9a95f0fdbfea
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/improbable-eng/grpc-web v0.12.0
	github.com/ipfs/go-log v1.0.4
	github.com/jrick/bitset v1.0.0
	github.com/libp2p/go-addr-util v0.0.2
	github.com/libp2p/go-libp2p v0.27.8
	github.com/libp2p/go-libp2p-core v0.19.0
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-mplex v0.4.1 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-swarm v0.4.0
	github.com/libp2p/go-libp2p-transport-upgrader v0.4.0
	github.com/libp2p/go-libp2p-yamux v0.5.1 // indirect
	github.com/libp2p/go-tcp-transport v0.2.1
	github.com/m4ksio/wal v1.0.0
	github.com/multiformats/go-multiaddr v0.9.0
	github.com/onflow/cadence v0.14.4
	github.com/onflow/flow-core-contracts/lib/go/contracts v0.7.1
	github.com/onflow/flow-go-sdk v0.17.0
	github.com/onflow/flow-go/crypto v0.12.0
	github.com/onflow/flow/protobuf/go/flow v0.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/rs/zerolog v1.19.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.8.2
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.3.0+incompatible // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/vmihailenco/msgpack/v4 v4.3.11
	go.uber.org/atomic v1.10.0
	golang.org/x/crypto v0.7.0
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/api v0.31.0
	google.golang.org/grpc v1.33.2
	gotest.tools v2.2.0+incompatible
)

replace mellium.im/sasl => github.com/mellium/sasl v0.2.1

// temp fix for MacOS build. See comment https://github.com/ory/dockertest/issues/208#issuecomment-686820414
replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

replace github.com/onflow/flow-go/crypto => ./crypto
