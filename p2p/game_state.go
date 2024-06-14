package p2p

import (
	"fmt"
	"sort"
	"strings"
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

	// PlayerActionExtraCard player ask for another card so palyer should have 3 cards
	PlayerActionExtraCard

	// PlayerActionStay player got a pok card
	PlayerActionPok
)

type Player struct {
	*Peer
	HandNumber     int
	Money          int
	Bet            int
	PlayerAction   *AtomicInt
	Card           [3]deck.Card
	lock           sync.RWMutex
	isOwner        bool
	PlayerActionCh chan (*PlayerActionMsg)
}
type PlayerActionMsg struct {
	PlayerID string
	Action   string
}

type Room struct {
	RoomId    string
	Players   map[string]*Player
	RoomOwner *Player
	GameState *GameState
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

	// GameStatusRoomStart game have been start
	GameStatusRoomStart

	// GameStatusRoomNotReady if not every player is ready
	GameStatusRoomNotReady

	// GameStatusPok The Owner of the room got pok
	GameStatusPok

	// GameStatusExtraCard The Owner of the room got pok
	GameStatusExtraCard

	// GameStatusNotPok The owner of the room is not getting pok
	// drawing a new card or stay
	GameStatusNotPok

	// GameStatusEnd Game end
	GameStatusEnd
)

func NewPlayer(isOwner bool, p *Peer, handNumber int) *Player {
	return &Player{
		lock:           sync.RWMutex{},
		isOwner:        isOwner,
		Money:          1000,
		Peer:           p,
		HandNumber:     handNumber,
		Bet:            10,
		PlayerAction:   NewAtomicInt(int32(PlayerActionNotReady)),
		PlayerActionCh: make(chan *PlayerActionMsg),
	}
}

func CreateRoom(p *Peer) *Room {
	// id, _ := gonanoid.Generate("0123456789", 8)
	owner := NewPlayer(true, p, 1)
	var players map[string]*Player = make(map[string]*Player)
	players[p.conn.RemoteAddr().String()] = owner

	return &Room{
		roomLock:  sync.RWMutex{},
		RoomId:    "1",
		Players:   players,
		RoomOwner: owner,
		GameState: &GameState{
			currentHand: NewAtomicInt(0),
			gameStatus:  NewAtomicInt(int32(GameStatusRoomNotReady)),
		},
		maxPlayer: 10,
		Deck:      deck.NewDeck(),
	}
}

func (r *Room) handleIncrementCurrentHand() {
	currentHand := r.GameState.currentHand.Get()
	if currentHand == int32(len(r.Players)) {
		r.GameState.currentHand.Set(1)
	} else {
		r.GameState.currentHand.Inc()
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
	case GameStatusExtraCard:
		return "Every player can ask for an extra card"
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

// return isPok, value, deng
// Check if a player has a Pok (8 or 9 points with the first two cards)
func (p *Player) CulculateTwoCard() (bool, int, int) {
	if len(p.Card) < 2 {
		return false, 0, 1
	}
	points, deng := calculatePok(p.Card[:2])
	return points == 8 || points == 9, points, deng
}

// it return a value of a card and the multiply of the bet
// Calculate the points for a given set of cards
func calculatePok(cards []deck.Card) (int, int) {
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

type SpecialWinning int32

const (
	Straight SpecialWinning = iota
	ThreeOfAKind
	ThreeOfJQK
	NotSpecialWinning
)

func (sw SpecialWinning) String() string {
	switch sw {
	case Straight:
		return "Straight"
	case ThreeOfAKind:
		return "Three Of A Kind"
	case ThreeOfJQK:
		return "Three Of JQK"
	case NotSpecialWinning:
		return "Not Special Winning"
	default:
		return "Unknown"
	}
}

const (
	fiveDeng  = 5
	threeDeng = 3
)

// return isSpecial, value, suit, deng
// suit will be ignore if isSpecial is false
func (p *Player) CulculateThreeCard() (SpecialWinning, int, deck.Suit, int) {

	// if it was three of a kind any value we return is the same and deng will be 5
	if val, suit, isThreeOfAKind := isThreeOfAKind(p.Card[:]); isThreeOfAKind {
		return ThreeOfAKind, val, suit, fiveDeng
	}

	// if it straight deng will be 3 and we will return the most value of the straight out
	// TODO: if dealer and player straight we might need to compare with suit
	if val, suit, isStraight := isStraight(p.Card[:]); isStraight {
		return Straight, val, suit, threeDeng
	}

	if val, suit, isStraight := isThreeOfJQK(p.Card[:]); isStraight {
		return ThreeOfJQK, val, suit, threeDeng
	}

	totalPts := 0
	deng := 1
	for _, card := range p.Card {
		val := card.Value
		if val > 10 {
			val = 10
		}
		totalPts += val
	}

	if (p.Card[0].Suit == p.Card[1].Suit && p.Card[1].Suit == p.Card[2].Suit) ||
		(p.Card[0].Value == p.Card[1].Value && p.Card[1].Value == p.Card[2].Value) {
		deng = threeDeng
	}

	return NotSpecialWinning, totalPts % 10, 0, deng

}

func isThreeOfAKind(cards []deck.Card) (int, deck.Suit, bool) {
	return cards[0].Value, cards[0].Suit, cards[0].Value == cards[1].Value && cards[1].Value == cards[2].Value
}

// isThreeOfJQK return the highest value and the highest suit and it's straight or not
func isThreeOfJQK(cards []deck.Card) (int, deck.Suit, bool) {

	values := []deck.Card{cards[0], cards[1], cards[2]}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Value < values[j].Value
	})

	for _, card := range cards {
		if card.Value < 11 || card.Value > 13 {
			return values[2].Value, values[2].Suit, false
		}
	}

	return values[2].Value, values[2].Suit, true
}

// isStraight return the highest value and the highest suit and it's straight or not
func isStraight(cards []deck.Card) (int, deck.Suit, bool) {

	values := []deck.Card{cards[0], cards[1], cards[2]}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Value < values[j].Value
	})

	return values[2].Value, values[2].Suit, values[0].Value+1 == values[1].Value && values[1].Value+1 == values[2].Value
}

// return a string of a current value and suit of the card that you holding
func (p *Player) CurrentCard() string {

	var cards []string
	for _, v := range p.Card {
		if v.Value != 0 {
			cardDetail := fmt.Sprintf("%s%s", deck.SuitToUnicode(v.Suit), v.SpecialCardValue(v.Value))
			cards = append(cards, cardDetail)
		}
	}
	// return fmt.Sprint("\n", strings.Join(cards, " "), "\n")
	return fmt.Sprint(strings.Join(cards, " "))
}
