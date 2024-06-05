package p2p

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/tgrziminiar/pok-deng-server-engine/deck"
)

// When game start broadcast to everyone and made a countdown
// and after that shuffle the deck and assign every peers of 2 card
// check for the pock of the dealer if deal got pock no one can take an extra card
// assign the hand for the players that is index of 1 first because the dealer must be the last who will take an action

func (s *Server) handleGameStart(p *Peer) {
	roomId, ok := s.playerInAnyRoom(p)
	if !ok {
		return
	}

	room := s.rooms[roomId]

	// set the game status to be start
	// set current hand to the player 1 and deal the cards to every palyer
	room.GameState.SetStatus(GameStatusRoomStart, 1)

	if err := s.broadcastSameMessage(roomId, fmt.Sprintln("Game will starting in 3 seconds . . .")); err != nil {
		slog.Error(err.Error())
		return
	}
	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		if err := s.broadcastSameMessage(roomId, fmt.Sprintf("\nGame starting in %d seconds\n", i)); err != nil {
			slog.Error(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	room.Deck = deck.Shuffle(room.Deck)

	for _, player := range room.Players {
		player.Card = [3]deck.Card{room.Deck[0], room.Deck[1]}
		// player.Card[0] = room.Deck[0]
		// player.Card[1] = room.Deck[1]
		room.Deck = room.Deck[2:]
		// fmt.Println("player remote: ", player.Peer.conn.RemoteAddr().String(), " ", player.Card[0], player.Card[1])
	}

	for _, player := range room.Players {
		player.Peer.Send([]byte(s.CurrentCard(player.Peer, roomId)))
	}

}

func (s *Server) CurrentCard(p *Peer, roomId string) string {

	room := s.rooms[roomId]
	peer := room.Players[p.conn.RemoteAddr().String()]
	var cards []string
	for _, v := range peer.Card {
		if v.Value != 0 {
			cardDetail := fmt.Sprintf("%s%s", deck.SuitToUnicode(v.Suit), v.SpecialCardValue(v.Value))
			cards = append(cards, cardDetail)
		}
	}
	return fmt.Sprint("\n", strings.Join(cards, " "), "\n")
}

func (s *Server) handleCurrentGame(p *Peer) {
	roomId, ok := s.playerInAnyRoom(p)
	if !ok {
		return
	}

	room := s.rooms[roomId]

	gameStatus := room.GameState.currentGameStatus()
	_ = gameStatus

}
