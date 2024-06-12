package p2p

import (
	"bytes"
	"fmt"
	"log/slog"
	"sort"
	"time"
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

// Function to handle the remove room command
func (s *Server) handleRemoveRoom(p *Peer) {
	roomId, exist := s.playerInAnyRoom(p)
	if exist {
		room, ok := s.rooms[roomId]
		if !ok {
			p.Send([]byte("Look like you haven't create any room yet, the owner (dealer) is the only one who can close the room\n"))
			return
		}
		s.removeRoom(room)
	} else {
		p.Send([]byte("Look like you haven't create any room yet, the owner (dealer) is the only one who can close the room\n"))
	}
}

// List all the room that exist on the server
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
func findMissingHandNumbers(players map[string]*Player) []int {
	// Find the maximum hand number
	maxHandNumber := 0
	for _, player := range players {
		if player.HandNumber > maxHandNumber {
			maxHandNumber = player.HandNumber
		}
	}

	// Create a boolean array to check for missing numbers
	present := make([]bool, maxHandNumber+1)
	for _, player := range players {
		present[player.HandNumber] = true
	}

	// Collect missing numbers
	missing := []int{}
	for i := 1; i <= maxHandNumber; i++ {
		if !present[i] {
			missing = append(missing, i)
		}
	}

	// Sort missing numbers in descending order
	sort.Slice(missing, func(i, j int) bool {
		return missing[i] > missing[j]
	})

	return missing
}

// Handle when the player try to join the room
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

			if room.maxPlayer < len(room.Players)+1 {
				p.Send([]byte("Since room is already full so you can't join the room"))
				return
			}
			id := p.conn.RemoteAddr().String()

			missingHandNumbers := findMissingHandNumbers(room.Players)
			var handNumber int
			if len(missingHandNumbers) > 0 {
				handNumber = missingHandNumbers[0]
			} else {
				handNumber = len(room.Players) + 1
			}

			player := NewPlayer(false, p, handNumber)
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

// Function to handle the remove room command
func (s *Server) handleQuitRoom(p *Peer) {
	roomId, exist := s.playerInAnyRoom(p)
	if exist {
		room, ok := s.rooms[roomId]
		if !ok {
			p.Send([]byte("Look like you haven't join any room yet, so you can't quit the room"))
			return
		}
		id := p.conn.RemoteAddr().String()
		player := room.Players[id]
		if player.isOwner {
			// remove room already have an broardcast of leaving
			s.removeRoom(room)
		} else {
			s.broadcastSameMessage(roomId, fmt.Sprintf("Player%d have left from the room (%d/%d)\n", player.HandNumber, len(room.Players)-1, room.maxPlayer))
			delete(room.Players, p.conn.RemoteAddr().String())
		}

	} else {
		p.Send([]byte("Look like you haven't join any room yet, so you can't quit the room\n"))
	}
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

// add room to server using mutex to prevent data race condition
func (s *Server) addRoom(room *Room) {
	defer room.roomLock.Unlock()
	room.roomLock.Lock()
	s.rooms[room.RoomId] = room
}

// remove all the players from the room so player connection still alive
// and remove the room from the server
func (s *Server) removeRoom(room *Room) {
	defer room.roomLock.Unlock()
	room.roomLock.Lock()

	if GameStatus(room.GameState.gameStatus.Get()) == GameStatusEnd || GameStatus(room.GameState.gameStatus.Get()) == GameStatusRoomNotReady {
		s.broadcastSameMessage(room.RoomId, "Room is closing in 3 seconds, since dealer (player1) desire to close to room\n")
		time.Sleep(500 * time.Millisecond)
		for i := 3; i > 0; i-- {
			if err := s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nRoom closing in %d seconds\n", i)); err != nil {
				slog.Error(err.Error())
			}
			time.Sleep(1 * time.Second)
		}

		s.broadcastSameMessage(room.RoomId, "You have been removed from the room "+room.RoomId+"\n")

		for _, player := range room.Players {
			delete(room.Players, player.Peer.conn.RemoteAddr().String())
		}

		delete(s.rooms, room.RoomId)
	}

}
