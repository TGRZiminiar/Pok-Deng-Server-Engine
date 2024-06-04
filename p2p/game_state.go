package p2p

import (
	"sync"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type PlayerAction int

const (
	// PlayerActionReady player read to play
	PlayerActionReady PlayerAction = iota

	// PlayerActionNotReady player not ready to play
	PlayerActionNotReady

	// PlayerActionBet player type the amount that they want to bet
	PlayerActionBet

	// PlayerActionStay player satisfy with 2 cards
	PlayerActionStay

	// PlayerActionStay player got a pok card
	PlayerActionPok
)

type Player struct {
	*Peer
	Money   int
	Bet     int
	Action  *AtomicInt
	lock    sync.RWMutex
	isOwner bool
}

type Room struct {
	RoomId    string
	Players   []*Player
	RoomOwner *Player
	GameState GameState
	roomLock  sync.RWMutex
	maxPlayer int
}

type GameState struct {
	currentHand *AtomicInt
	gameStatus  *AtomicInt
}

type GameStatus int32

const (
	// GameStatusRoomReady if every player is ready
	GameStatusRoomReady GameStatus = iota

	// GameStatusRoomNotReady if not every player is ready
	GameStatusRoomNotReady

	// GameStatusPok The Owner of the room got pok
	GameStatusPok

	// GameStatusNotPok The owner of the room is not getting pok
	// drawing a new card or stay
	GameStatusNotPok

	// GameStatusEnd Game end
	GameStatusEnd
)

// func NewRoom() *Room {
// 	r := &Room{
// 		roomLock:  sync.RWMutex{},
// 		Players:   []*Player{},
// 		RoomOwner: &Player{},
// 		GameState: GameState{},
// 	}

// 	return r
// }

func NewPlayer(isOwner bool, p *Peer) *Player {
	return &Player{
		lock:    sync.RWMutex{},
		isOwner: isOwner,
		Money:   1000,
		Peer:    p,
		Bet:     -1,
		Action:  NewAtomicInt(int32(PlayerActionNotReady)),
	}
}

func CreateRoom(p *Peer) *Room {
	id, _ := gonanoid.Generate("0123456789", 8)
	owner := NewPlayer(true, p)
	return &Room{
		roomLock:  sync.RWMutex{},
		RoomId:    id,
		Players:   []*Player{owner},
		RoomOwner: owner,
		GameState: GameState{
			currentHand: NewAtomicInt(-1),
			gameStatus:  NewAtomicInt(int32(GameStatusRoomNotReady)),
		},
		maxPlayer: 10,
	}
}
