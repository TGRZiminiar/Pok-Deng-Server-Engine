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
// otherwise assign the hand for the players that is index of 1 first because the dealer must be the last who will take an action
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

	// sleep for a bit cause the message will be too quick
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
		player.PlayerAction.Set(int32(PlayerActionBet))

	}

	for _, player := range room.Players {
		player.Peer.Send([]byte(s.CurrentCard(player.Peer, roomId)))
		// calculate the card point of the players
		isPok, point, deng := player.CulculateTwoCard()

		// if someone got pok their can't have another card
		// if he was an dealer of the room no player can have more action
		// only way player can win or deal is got a pok and more deng than dealer
		if isPok {
			player.Peer.Send([]byte(fmt.Sprintf("\nYou got a pok of %d and %d of deng\n", point, deng)))
			if player.isOwner && isPok {
				room.GameState.SetStatus(GameStatusPok, 1)
				if err := s.broadcastSameMessage(roomId, fmt.Sprintln("Dealer got a pok and you can't ask for another card result of the game will be calculate in 1 seconds")); err != nil {
					slog.Error(err.Error())
				}
			}
		}
	}
}

// return a string of a current value and suit of the card that you holding
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
	playerAction := room.Players[p.conn.RemoteAddr().String()].currentPlayerAction()
	text := fmt.Sprintf("\nGame Status: %s\nYour Action: %s\n", gameStatus, playerAction)
	p.Send([]byte(text))
}

func helperWhoseWinDealer(room Room, dealer *Player) {
	dealerPok, dealerPts, dealerDeng := dealer.CulculateTwoCard()
	for _, player := range room.Players {
		plyerIsPok, playerPts, playerDeng := player.CulculateTwoCard()

		// if dealer got pok
		if dealerPok {
			// if player not pok mean they instanly lose
			if !plyerIsPok {
				playerPts = 0
				playerDeng = 0
				// remove some money and send a message to them that they are lose
			}

			if dealerPts > playerPts {
				// remove some money and send a message to them that they are lose

				// if both point and deng is the same that mean they are equal
			} else if dealerPts == playerPts && dealerDeng == playerDeng {
				return
				// dealer is winning if dealerdeng is more than palyerdeng
			} else if dealerPts == playerPts && dealerDeng > playerDeng {

				// player is winning if dealerdeng is less than palyerdeng
			} else if dealerPts == playerPts && dealerDeng < playerDeng {

			}

		}

	}
}
