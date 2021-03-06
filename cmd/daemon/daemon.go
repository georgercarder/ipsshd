package main

import (
	"flag"
	"fmt"
	"io"
	"net"

	nd "github.com/georgercarder/ip-sesh/node"
	sh "github.com/georgercarder/ip-sesh/shell"
	sg "github.com/georgercarder/ip-sesh/subnet-genie"

	"github.com/ipfs/go-ipfs/core"
)

func main() {
	role := flag.String("role", "server",
		"daemon role can be server or client.")
	flag.Parse()
	switch *role {
	case "server":
		fmt.Println("server")
		fmt.Println("initializing node ...")
		n := nd.G_Node()
		fmt.Println("Identity", n.Identity)
		// fast bootstrap
		sg.FastBootstrap((*core.IpfsNode)(n))
		ps := n.Peerstore.Peers()
		fmt.Println("peers", len(ps))
		// announce provide
		go sg.AnnounceProvide((*core.IpfsNode)(n))
		// serve domain
		nd.ServeDomain("test.domain.com")
		break
	case "client":
		fmt.Println("client")
		fmt.Println("initializing node ...")
		n := nd.G_Node()
		// fast bootstrap
		sg.FastBootstrap((*core.IpfsNode)(n))
		ps := n.Peerstore.Peers()
		fmt.Println("peers", len(ps))
		go sg.JoinProviders((*core.IpfsNode)(n))
		ln, err := net.Listen("tcp", ":8081")
		// make sure 8081 is firewalled
		if err != nil {
			panic(err) // TODO LOG
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				// TODO LOG
				break // TODO UPDATE FOR FOR LOOP
			}
			//defer conn.Close()
			fmt.Println("debug connection", conn)
			b := make([]byte, 1024)
			n, err := conn.Read(b)
			if err != nil {
				// LOG ERR
				continue
			}
			go func() {
				defer conn.Close()
				domain := string(b[:n])
				ok := nd.ClientDomainIsValid(domain)
				if !ok {
					// LOG
					return
				}
				fmt.Println("debug domain valid", domain)
				// TODO start thread for that session
				// TODO pipe to calling client
				connBundleCH := nd.StartHandshake(domain)
				// TODO PUT TIMEOUT
				cb := <-connBundleCH
				conn.Write([]byte(sh.StartShellSession))
				go func() {
					_, _ = io.Copy(cb.Conn, conn)
					cb.StopCH <- true
				}()
				_, err = io.Copy(conn, cb.Conn)
				cb.StopCH <- true
			}()
		}

	}
	select {}
}
