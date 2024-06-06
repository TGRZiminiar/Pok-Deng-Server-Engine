package p2p

import (
	"sync"

	"github.com/tgrziminiar/pok-deng-server-engine/deck"
)

type PlayerAction int32

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
	Money        int
	Bet          int
	PlayerAction *AtomicInt
	Card         [3]deck.Card
	lock         sync.RWMutex
	isOwner      bool
}

type Room struct {
	RoomId    string
	Players   map[string]*Player
	RoomOwner *Player
	GameState GameState
	roomLock  sync.RWMutex
	maxPlayer int
	Deck      []deck.Card
}

type GameState struct {
	currentHand *AtomicInt
	gameStatus  *AtomicInt
}

type GameStatus int32

const (
	// GameStatusRoomReady if every player is ready
	GameStatusRoomReady GameStatus = iota
	GameStatusRoomStart
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
		lock:         sync.RWMutex{},
		isOwner:      isOwner,
		Money:        1000,
		Peer:         p,
		Bet:          -1,
		PlayerAction: NewAtomicInt(int32(PlayerActionNotReady)),
	}
}

func CreateRoom(p *Peer) *Room {
	// id, _ := gonanoid.Generate("0123456789", 8)
	owner := NewPlayer(true, p)
	var players map[string]*Player = make(map[string]*Player)
	players[p.conn.RemoteAddr().String()] = owner

	return &Room{
		roomLock:  sync.RWMutex{},
		RoomId:    "123",
		Players:   players,
		RoomOwner: owner,
		GameState: GameState{
			currentHand: NewAtomicInt(-1),
			gameStatus:  NewAtomicInt(int32(GameStatusRoomNotReady)),
		},
		maxPlayer: 10,
		Deck:      deck.NewDeck(),
	}
}

func (g *GameState) SetStatus(s GameStatus, hand int32) {
	g.gameStatus.Set(int32(s))
	g.currentHand.Set(hand)
}

func (g *GameState) currentGameStatus() string {
	switch GameStatus(g.gameStatus.Get()) {
	case GameStatusRoomReady:
		return "Game is ready to start."
	case GameStatusRoomStart:
		return "Game is starting."
	case GameStatusRoomNotReady:
		return "Game is not ready. Waiting for all players to be ready."
	case GameStatusPok:
		return "The owner of the room got Pok!"
	case GameStatusNotPok:
		return "The owner of the room did not get Pok. Drawing a new card or staying."
	case GameStatusEnd:
		return "Game has ended."
	default:
		return "Unknown game status."
	}
}
func (p *Player) currentPlayerAction() string {
	switch PlayerAction(p.PlayerAction.Get()) {
	case PlayerActionReady:
		return "Player is ready"
	case PlayerActionBet:
		return "Player should type the amount they want to bet"
	case PlayerActionNotReady:
		return "Player is not ready"
	case PlayerActionPok:
		return "Player got pok"
	case PlayerActionStay:
		return "Player satisfy with cards"
	default:
		return "Unknown current hand status"
	}
}

// Calculate the points for a given set of cards
// it return a value of a card and the multiply of the bet
func CalculatePok(cards []deck.Card) (int, int) {
	total := 0
	deng := 1
	for _, card := range cards {
		val := card.Value
		if val > 10 {
			val = 10
		}
		total += val
	}
	if cards[0].Suit == cards[1].Suit || cards[0].Value == cards[1].Value {
		deng = 2
	}
	return int(total % 10), int(deng)
}

// Check if a player has a Pok (8 or 9 points with the first two cards)
// return isPok, value, deng
func (p *Player) CulculateTwoCard() (bool, int, int) {
	if len(p.Card) < 2 {
		return false, 0, 0
	}
	points, deng := CalculatePok(p.Card[:2])
	return points == 8 || points == 9, points, deng
}
