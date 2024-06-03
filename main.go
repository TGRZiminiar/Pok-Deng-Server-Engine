package main

import "github.com/tgrziminiar/pok-deng-server-engine/p2p"

func main() {
	s := p2p.NewServer(p2p.ServerConfig{
		ListenAddr: ":3000",
	})

	s.Start()

}
