package main

import (
	"log"

	"github.com/tgrziminiar/pok-deng-server-engine/p2p"
)

func main() {

	s := p2p.NewServer(p2p.ServerConfig{
		ListenAddr: ":3000",
	})

	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
