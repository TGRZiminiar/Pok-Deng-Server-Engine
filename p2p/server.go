package p2p

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
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
	msgCh        chan *Message
	rooms        map[string]*Room
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
		tcpTransport: transport,
		msgCh:        make(chan *Message, 100),
		rooms:        make(map[string]*Room),
	}
	// when accpet the connection so we can trigger the read loop using channel
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
			slog.Info("recieving :", "msg", msg)
			s.handleMessage(msg)
			_ = msg
		}
	}
}

// function to handle incoming new peer and read all the message that peer had sent
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

// handling message that coming from client
// we handle the type of the payload first (incase we might change the way we communicate with server)
// second we handle the command that user can input to server
func (s *Server) handleMessage(msg *Message) error {
	peer := s.peers[msg.From]
	switch v := msg.Payload.(type) {
	case CommandHelp:
		fmt.Println("Help command received")
	case string:
		switch {
		case v == CommandHelp{}.String():
			peer.Send([]byte(
				"1. /create-room \t\t[to create a room and you will be a dealer, you can create only one room per time]\n" +
					"2. /list-room \t\t\t[list all the rooms that you can join]\n" +
					"3. /join-room (roomId) \t\t[join the room with roomId]\n" +
					"5. /current-room \t\t[show the current toom that you are exist]\n" +
					"6. /start \t\t\t[start the game, only the dealer (player1) have a privilege to start the game]\n" +
					"7. /current-game \t\t[show the current status of the game, current hand and which hand are you]\n" +
					"8. /stay \t\t\t[if you are satisfy with two cards you can choose stay]\n" +
					"9. /more \t\t\t[if you are not satisfy with two cards you ask for an extra card]\n" +
					"10. /close-room \t\t[close the room and every player will get leave the room automatically, only the dealer (player1) have a privilege to close the room]\n",
			))

		case v == CommandCreateRoom{}.String():
			s.handleCreateRoom(peer)

		case v == CommandListRoom{}.String():
			s.handleListRoom(peer)

		case strings.HasPrefix(v, "/join-room"):
			parts := strings.SplitN(v, " ", 2)
			if len(parts) < 2 {
				fmt.Println("no room id provided")
				peer.Send([]byte("\nno roomId found!\n"))
			} else {
				roomId := strings.TrimSpace(parts[1])
				if roomId == "" {
					fmt.Println("no room id")
				} else {
					s.handleJoinRoom(peer, roomId)
				}
			}

		case v == CommandCurrentRoom{}.String():
			s.handleCurrentRoom(peer)

		case v == CommandStartGame{}.String():
			s.handleGameStart(peer)

		case v == CommandCurrentGame{}.String():
			s.handleCurrentGame(peer)
		case v == CommandStay{}.String():
			s.handlePlayerStay(peer)
		case v == CommandExtraCard{}.String():
			s.handlePlayerMoreExtraCard(peer)
		case v == CommandRemoveRoom{}.String():
			s.handleRemoveRoom(peer)
		case v == CommandQuitRoom{}.String():
			s.handleQuitRoom(peer)

		default:
			// fmt.Println("Message from normal string", v)
		}
	default:
	}
	return nil
}

// broadcast to every peers in room isong multiwriter
func (s *Server) broadcastSameMessage(roomId string, msg string) error {
	room, exists := s.rooms[roomId]
	if !exists {
		return errors.New("RoomId does not exist")
	}

	peers := make([]io.Writer, 0, len(room.Players))

	for _, player := range room.Players {
		if player.Peer != nil && player.Peer.conn != nil {
			peers = append(peers, player.conn)
		}
	}

	mw := io.MultiWriter(peers...)

	if _, err := mw.Write([]byte(msg)); err != nil {
		return err
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

// currenty we just implementing just a simple string
// so we doesn't need to handshake with th other
// basically how can we handshake with just a simple string lol
func (s *Server) handshake(p *Peer) error {
	return nil
}

// close peer connection and remove them from the list
func (s *Server) removeAndClosePeerConnection(p *Peer) {
	p.conn.Close()
	delete(s.peers, p.conn.RemoteAddr().String())
}

// func init() {
// 	gob.Register(CommandHelp{})
// }
