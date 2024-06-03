package p2p

import (
	"fmt"
	"log/slog"
	"sync"
)

type ServerConfig struct {
	ListenAddr string
}

type Server struct {
	ServerConfig
	tcpTransport *TCPTransport
	peerLock     sync.RWMutex
	peers        map[string]*Peer
	addPeer      chan *Peer
	delPeer      chan *Peer
	msgCh        chan *Message
}

func NewServer(cfg ServerConfig) *Server {

	serverCfg := ServerConfig{
		ListenAddr: cfg.ListenAddr,
	}

	transport := NewTCPTransport(serverCfg.ListenAddr)
	s := &Server{
		ServerConfig: serverCfg,
		peerLock:     sync.RWMutex{},
		peers:        make(map[string]*Peer),
		addPeer:      make(chan *Peer, 10),
		delPeer:      make(chan *Peer),
		tcpTransport: transport,
		msgCh:        make(chan *Message, 100),
	}
	// when accpet the connection so we can trigger the read loop usingb channel
	transport.AddPeer = s.addPeer
	return s
}

// Start will listen to the new connection using tcp transprot
// and also start an infinite read loop to accept the channel of
// adding new peer/ delete peer/  message
func (s *Server) Start() error {

	go s.loop()

	// blocking with infinite loop here
	if err := s.tcpTransport.ListenAndAccept(); err != nil {
		slog.Error("TCPTransport failed to listen and accpet", "error", err)
		return err
	}

	return nil
}

// function to handle all the incoming message
func (s *Server) loop() {
	defer func() {
		if err := s.tcpTransport.Close(); err != nil {
			slog.Error("TCPTransport failed to close the connection", "error", err)
		}
	}()

	for {
		select {
		case peer := <-s.addPeer:
			if err := s.handleIncomingNewPeer(peer); err != nil {
				s.removeAndClosePeerConnection(peer)
				slog.Error("failed to handle incoming new peer", "error", err)
			}
			_ = peer
		case msg := <-s.msgCh:
			fmt.Println(msg)
			_ = msg
		}
	}
}

func (s *Server) handleIncomingNewPeer(p *Peer) error {
	if err := s.handshake(p); err != nil {
		s.removeAndClosePeerConnection(p)
		return err
	}

	go p.PeerReadLoop(s.msgCh)

	peerInLists := s.checkPeerInPeers(p)
	if !peerInLists {
		s.AddPeerToPeers(p)
	} else {
		p.Send([]byte("failed to add you in peers list"))
		s.removeAndClosePeerConnection(p)
	}

	return nil
}

func (s *Server) AddPeerToPeers(p *Peer) {
	defer s.peerLock.Unlock()
	s.peerLock.Lock()

	s.peers[p.conn.RemoteAddr().String()] = p
	slog.Info("connected with remote addr %s\n", "remote", p.conn.RemoteAddr().String())

}

func (s *Server) checkPeerInPeers(p *Peer) bool {
	_, ok := s.peers[p.conn.RemoteAddr().String()]
	return ok
}

func (s *Server) handshake(p *Peer) error {
	return nil
}

// close peer connection and remove them from the list
func (s *Server) removeAndClosePeerConnection(p *Peer) {
	p.conn.Close()
	delete(s.peers, p.conn.RemoteAddr().String())
}
