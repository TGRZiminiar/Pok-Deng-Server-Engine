package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// create the room and the owner of the room is the person who create it by default
func (s *Server) handleCreateRoom(p *Peer) *Room {

	roomId, exist := s.playerInAnyRoom(p)
	if !exist {
		room := CreateRoom(p)

		s.addRoom(room)
		p.Send([]byte(fmt.Sprintf("\nCreate room success!\nYour roomId is : %s\n\n", room.RoomId)))
		return room

	} else {
		p.Send([]byte(fmt.Sprintf("\nYou already have a room that you have joined %s\n\n", roomId)))
		return nil
	}

}

func (s *Server) addRoom(room *Room) {
	defer room.roomLock.Unlock()
	room.roomLock.Lock()
	s.rooms[room.RoomId] = room
}

/*
79 Rooms
1. 491042349 (1/10)
2. 124214214 (2/10)
*/

func (s *Server) handleListRoom(p *Peer) {
	var roomList bytes.Buffer

	roomList.WriteString(fmt.Sprintf("\n%d Rooms exists\n", len(s.rooms)))

	roomCount := 1
	for _, room := range s.rooms {
		fmt.Fprintf(&roomList, "%d. %s (%d/%d)\n\n", roomCount, room.RoomId, len(room.Players), room.maxPlayer)
		roomCount++
	}

	p.Send([]byte(roomList.Bytes()))
}

func (s *Server) handleJoinRoom(p *Peer, roomId string) {
	var text string

	room, ok := s.rooms[roomId]
	if ok {
		_, exist := s.playerInAnyRoom(p)

		if !exist {
			defer room.roomLock.Unlock()
			room.roomLock.Lock()

			id := p.conn.RemoteAddr().String()
			player := NewPlayer(false, p)
			room.Players[id] = player

			if err := s.broadcastInRoom(roomId, fmt.Sprintf("bc -> new player [%s] have join the room (%d/%d)", id, len(room.Players), room.maxPlayer)); err != nil {
				s.removeAndClosePeerConnection(p)
			}

			text = "\nJoin room " + roomId + " success!\n"

		} else {
			text = fmt.Sprintf("\nYou already have a room that you have joined %s\n\n", roomId)
		}
	} else {
		text = "\nRoom " + roomId + " is not exist\n"
	}
	p.Send([]byte(text))
}

// handleCurrentRoom just to show the player where is he like now
func (s *Server) handleCurrentRoom(p *Peer) {
	var room *Room = nil
	id := p.conn.RemoteAddr().String()
	for _, v := range s.rooms {
		_, ok := v.Players[id]
		if ok {
			room = v
			break
		}
	}

	if room != nil {
		p.Send([]byte(fmt.Sprintf("\ncurrent room is : %s\n", room.RoomId)))
	} else {
		p.Send([]byte("\nYou are not currently in any room.\n"))
	}
}

// broadcast to every peers in room isong multiwriter
func (s *Server) broadcastInRoom(roomId string, msg string) error {
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

// returning an roomid and if a peer is in any room or not
func (s *Server) playerInAnyRoom(p *Peer) (string, bool) {
	var exist bool = false
	id := p.conn.RemoteAddr().String()
	for _, v := range s.rooms {
		_, ok := v.Players[id]
		if ok {
			return v.RoomId, ok
		}

	}

	return "", exist

}
