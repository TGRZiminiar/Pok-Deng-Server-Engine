package p2p

import (
	"bytes"
	"fmt"
)

// create the room and the owner of the room is the person who create it by default
func (s *Server) handleCreateRoom(p *Peer) *Room {

	roomId, exist := s.playerInAnyRoom(p)
	if !exist {
		room := CreateRoom(p)
		s.addRoom(room)
		p.Send([]byte(fmt.Sprintf("\nCreate room success!\nYour roomId is : %s\nYou are player %d\n", room.RoomId, 1)))
		return room

	} else {
		p.Send([]byte(fmt.Sprintf("\nYou already have a room that you have joined %s\n\n", roomId)))
		return nil
	}

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
	if !ok {
		text = "\nRoom " + roomId + " is not exist\n"
	} else {
		_, exist := s.playerInAnyRoom(p)

		if !exist {
			defer room.roomLock.Unlock()
			room.roomLock.Lock()

			id := p.conn.RemoteAddr().String()
			player := NewPlayer(false, p, len(room.Players)+1)
			room.Players[id] = player

			if err := s.broadcastSameMessage(roomId, fmt.Sprintf("\nbc -> new player [%d] have join the room (%d/%d)\n", player.HandNumber, len(room.Players), room.maxPlayer)); err != nil {
				s.removeAndClosePeerConnection(p)
			}

			text = "\nJoin room " + roomId + " success!\nYou are player" + fmt.Sprint(player.HandNumber) + "!\n"

		} else {
			text = fmt.Sprintf("\nYou already have a room that you have joined your room is: %s\n\n", roomId)
		}
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

func (s *Server) addRoom(room *Room) {
	defer room.roomLock.Unlock()
	room.roomLock.Lock()
	s.rooms[room.RoomId] = room
}
