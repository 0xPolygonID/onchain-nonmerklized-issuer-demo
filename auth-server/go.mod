module github.com/0xPolygonID/onchain-issuer-integration-demo/server

go 1.20

require (
	github.com/ethereum/go-ethereum v1.13.10
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/render v1.0.3
	github.com/google/uuid v1.5.0
	github.com/iden3/go-iden3-auth v1.2.3
	github.com/iden3/go-iden3-core v1.0.2
	github.com/iden3/iden3comm v1.1.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rs/cors v1.10.1
	github.com/ugorji/go/codec v1.2.12
	golang.ngrok.com/ngrok v1.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.12.1 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20231225121904-e25f5bc08668 // indirect
	github.com/crate-crypto/go-kzg-4844 v0.7.0 // indirect
	github.com/dchest/blake512 v1.0.0 // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dustinxie/ecc v0.0.0-20210511000915-959544187564 // indirect
	github.com/ethereum/c-kzg-4844 v0.4.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/iden3/contracts-abi/state/go/abi v1.0.1 // indirect
	github.com/iden3/go-circuits v1.0.3 // indirect
	github.com/iden3/go-iden3-crypto v0.0.15 // indirect
	github.com/iden3/go-jwz v1.0.0 // indirect
	github.com/iden3/go-merkletree-sql/v2 v2.0.6 // indirect
	github.com/iden3/go-rapidsnark/prover v0.0.10 // indirect
	github.com/iden3/go-rapidsnark/types v0.0.3 // indirect
	github.com/iden3/go-rapidsnark/verifier v0.0.5 // indirect
	github.com/iden3/go-rapidsnark/witness v0.0.6 // indirect
	github.com/iden3/go-schema-processor v1.3.1 // indirect
	github.com/iden3/wasmer-go v0.0.1 // indirect
	github.com/inconshreveable/log15 v3.0.0-testing.5+incompatible // indirect
	github.com/inconshreveable/log15/v3 v3.0.0-testing.5 // indirect
	github.com/ipfs/boxo v0.17.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/ipfs/go-ipfs-api v0.7.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.4 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.0.19 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-flow-metrics v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.32.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.12.1 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/piprate/json-gold v0.5.1-0.20230111113000-6ddbe6e6f19f // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/supranational/blst v0.3.11 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.ngrok.com/muxado/v2 v2.0.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/exp v0.0.0-20240112132812-db7319d0e0e3 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/term v0.16.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.2 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
