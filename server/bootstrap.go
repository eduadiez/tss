package server

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/tss/common"
	"github.com/bnb-chain/tss/p2p"

	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p/core/crypto"
)

var logger = log.Logger("srv")

type TssBootstrapServer struct{}

func NewTssBootstrapServer(home string, config common.P2PConfig) *TssBootstrapServer {
	bs := TssBootstrapServer{}

	var privKey crypto.PrivKey
	pathToNodeKey := path.Join(home, "node_key")
	if _, err := os.Stat(pathToNodeKey); err == nil {
		bytes, err := ioutil.ReadFile(pathToNodeKey)
		if err != nil {
			common.Panic(err)
		}
		privKey, err = crypto.UnmarshalPrivateKey(bytes)
		if err != nil {
			common.Panic(err)
		}
	} else {
		common.Panic(err)
	}

	addr, err := multiaddr.NewMultiaddr(config.ListenAddr)
	if err != nil {
		common.Panic(err)
	}

	ctx := context.Background()
	host, err := libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.Identity(privKey),
		libp2p.EnableRelay(),
		libp2p.NATPortMap(),
	)
	if err != nil {
		common.Panic(err)
	}

	ds, err := leveldb.NewDatastore(path.Join(home, "rt/"), nil)
	if err != nil {
		common.Panic(err)
	}

	kademliaDHT, err := libp2pdht.New(
		ctx,
		host,
		libp2pdht.Datastore(ds),
		libp2pdht.Mode(libp2pdht.ModeServer),
	)
	if err != nil {
		common.Panic(err)
	}

	go p2p.DumpDHTRoutine(kademliaDHT)
	go p2p.DumpPeersRoutine(host)

	logger.Info("Bootstrap server has started, id: ", host.ID().String())

	return &bs
}
