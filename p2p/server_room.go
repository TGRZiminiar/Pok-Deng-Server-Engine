package p2p

import (
	"bytes"
	"fmt"
)

// create the room and the owner of the room is the person who create it by default
func (s *Server) handleCreateRoom(p *Peer) *Room {
	room := CreateRoom(p)
	s.addRoom(room)
	p.Send([]byte(fmt.Sprintf("\nCreate room success!\nYour roomId is : %s\n\n", room.RoomId)))
	return room
}

func (s *Server) addRoom(room *Room) {
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
	if !ok {
		text = "\nRoom " + roomId + " is not exist\n"
	} else {
		player := NewPlayer(false, p)
		room.Players = append(room.Players, player)
		text = "\nJoin room " + roomId + " success!\n"
	}
	p.Send([]byte(text))
}
