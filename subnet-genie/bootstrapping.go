package subnet_genie

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	. "github.com/georgercarder/ip-sesh/common"
	nd "github.com/georgercarder/ip-sesh/node"

	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-core/peer"
)

func FastBootstrap(n *core.IpfsNode) {
	numPeers := 0
	fmt.Println("fast bootstrapping ...")
	for numPeers < 1000 {
		ps := n.Peerstore.Peers()
		numPeers = len(ps)
		time.Sleep(100 * time.Millisecond)
		go func() {
			dht := n.DHT
			rval := make([]byte, 32)
			rand.Read(rval)
			ctx, cancel := context.WithTimeout(
				context.Background(), 2*time.Second)
			defer cancel()
			_, err := dht.GetValue(ctx, string(rval))
			if err != nil {
				//fmt.Println("debug err", err)
			}
		}()
	}
	fmt.Println("fast bootstrapping done.")
}

func AnnounceProvide(n *core.IpfsNode) {
	for { // an interval
		key, err := String2CID("/ip-sesh/0.0.1")
		if err != nil {
			fmt.Println("debug conv err", err)
		}
		ctx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second)
		defer cancel()
		err = n.Routing.Provide(ctx, key, true)
		if err != nil {
			fmt.Println("debug provide err", err)
		}
		time.Sleep(5 * time.Minute)
	}
}

// the point is to connect with providers of "/ip-sesh/0.0.1"
// to propagate the pubsub pubs to these "providers"
func JoinProviders(n *core.IpfsNode) {
	key, err := String2CID("/ip-sesh/0.0.1")
	if err != nil {
		fmt.Println("debug conv err", err)
	}
	numProvs := 1
	foundOne := false
	for !foundOne && numProvs < 1024 { // TODO make max const
		ctx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second)
		defer cancel()
		pchan := n.DHT.FindProvidersAsync(ctx, key, numProvs)
		ct := 0
		for ct < numProvs { // TODO PUT IN TIMEOUT
			ct++
			p := <-pchan
			//fmt.Println("debug Provider found", p)
			if len(p.Addrs) < 1 {
				continue
			}
			go func(pp peer.AddrInfo) {
				ctx2, cancel2 := context.WithTimeout(
					context.Background(), 10*time.Second)
				defer cancel2()
				_, err := nd.FindPeer(ctx2, pp.ID)
				if err != nil {
					//fmt.Println("debug FindPeer", err)
				}
				//fmt.Println("debug FindPeer done", pr.ID)
				foundOne = true // this is janky FIXME
			}(p)
		}
		numProvs *= 2
	}
}
