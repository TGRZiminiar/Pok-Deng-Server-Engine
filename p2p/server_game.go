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
	for i := 3; i > 0; i-- {
		if err := s.broadcastSameMessage(roomId, fmt.Sprintf("\nGame starting in %d seconds\n", i)); err != nil {
			slog.Error(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	room.Deck = deck.Shuffle(room.Deck)
	var dealer *Player
	for _, player := range room.Players {
		player.Card = [3]deck.Card{room.Deck[0], room.Deck[1]}
		room.Deck = room.Deck[2:]
		player.PlayerAction.Set(int32(PlayerActionBet))
		if player.isOwner {
			dealer = player
		}
	}
	var dealerPok bool = false
	for _, player := range room.Players {
		cards := s.CurrentCard(player.Peer, roomId)
		player.Peer.Send([]byte(fmt.Sprintf("\nYour cards: %s\n", cards)))

		// calculate the card point of the players
		isPok, point, deng := player.CulculateTwoCard()

		// if someone got pok their can't have another card
		// if he was an dealer of the room no player can have more action
		// only way player can win or deal is got a pok and more deng than dealer
		if isPok {
			player.Peer.Send([]byte(fmt.Sprintf("\nYou got a pok of %d and %d of deng\n", point, deng)))
			if player.isOwner && isPok {
				dealerPok = true
				room.GameState.SetStatus(GameStatusPok, 1)
				handleTwoCard(room, dealer)
			}
		}
	}
	if !dealerPok {
		s.broadcastSameMessage(roomId, "\ndealer doesn't got pok now it turn of player 1\n")
		time.Sleep(1 * time.Second)
	}
}

// return a string of a current value and suit of the card that you holding
// TODO: may be we use the player to check the current card instead
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

func handleTwoCard(room *Room, dealer *Player) {
	dealerPok, dealerPts, dealerDeng := dealer.CulculateTwoCard()
	for _, player := range room.Players {
		fmt.Println(player.Peer.conn.RemoteAddr().String(), player.isOwner)
		if player.isOwner {
			continue
		}
		player.Send([]byte(fmt.Sprintf("\nDealer got %s\n", dealer.CurrentCard())))

		playerPok, playerPts, playerDeng := player.CulculateTwoCard()
		resultMessage := ""

		// if dealer got pok
		if dealerPok {
			// if player not pok mean they instantly lose
			if !playerPok {
				player.Money -= player.Bet
				dealer.Money += player.Bet
				resultMessage = fmt.Sprintf("You lost to the dealer's Pok You lost %d\nYour current balance is %d\n", player.Bet, player.Money)
				player.Send([]byte(resultMessage))
				continue
			}

			if dealerPts > playerPts {
				player.Money -= player.Bet * playerDeng
				dealer.Money += player.Bet * playerDeng
				resultMessage = fmt.Sprintf("You lost to the dealer's higher Pok points You lost %d\nYour current balance is %d\n", player.Bet*playerDeng, player.Money)
			} else if dealerPts == playerPts && dealerDeng == playerDeng {
				resultMessage = "It's a draw with the dealer."
			} else if dealerPts == playerPts && dealerDeng > playerDeng {
				player.Money -= player.Bet * (dealerDeng - playerDeng)
				dealer.Money += player.Bet * (dealerDeng - playerDeng)
				resultMessage = fmt.Sprintf("You lost to the dealer's higher Pok multiplier You lost %d\nYour current balance is %d\n", player.Bet*(dealerDeng-playerDeng), player.Money)
			} else if dealerPts == playerPts && dealerDeng < playerDeng {
				player.Money += player.Bet * (playerDeng - dealerDeng)
				dealer.Money -= player.Bet * (playerDeng - dealerDeng)
				resultMessage = fmt.Sprintf("You win against the dealer's higher Pok multiplier You won %d\nYour current balance is %d\n", player.Bet*(playerDeng-dealerDeng), player.Money)
			} else if playerPts > dealerPts {
				player.Money += player.Bet * playerDeng
				dealer.Money -= player.Bet * playerDeng
				resultMessage = fmt.Sprintf("You won with higher Pok points against the dealer! You won %d\nYour current balance is %d\n", player.Bet*playerDeng, player.Money)
			}
		} else {
			if playerPok {
				player.Money += player.Bet * playerDeng
				dealer.Money -= player.Bet * playerDeng
				resultMessage = fmt.Sprintf("You won with Pok against the dealer! You won %d\nYour current balance is %d\n", player.Bet*playerDeng, player.Money)
				player.Send([]byte(resultMessage))
				continue
			}

			if dealerPts > playerPts {
				player.Money -= player.Bet * playerDeng
				dealer.Money += player.Bet * playerDeng
				resultMessage = fmt.Sprintf("You lost to the dealer You lost %d\nYour current balance is %d\n", player.Bet*playerDeng, player.Money)
			} else if dealerPts == playerPts && dealerDeng == playerDeng {
				resultMessage = "It's a draw with the dealer"
			} else if dealerPts == playerPts && dealerDeng > playerDeng {
				player.Money -= player.Bet * (dealerDeng - playerDeng)
				dealer.Money += player.Bet * (dealerDeng - playerDeng)
				resultMessage = fmt.Sprintf("You lost to the dealer's higher multiplier You lost %d\nYour current balance is %d\n", player.Bet*(dealerDeng-playerDeng), player.Money)
			} else if dealerPts == playerPts && dealerDeng < playerDeng {
				player.Money += player.Bet * (playerDeng - dealerDeng)
				dealer.Money -= player.Bet * (playerDeng - dealerDeng)
				resultMessage = fmt.Sprintf("You won with a higher multiplier against the dealer! You won %d\nYour current balance is %d\n", player.Bet*(playerDeng-dealerDeng), player.Money)
			} else if playerPts > dealerPts {
				player.Money += player.Bet * playerDeng
				dealer.Money -= player.Bet * playerDeng
				resultMessage = fmt.Sprintf("You won against the dealer! You won %d\nYour current balance is %d\n", player.Bet*playerDeng, player.Money)
			}
		}
		player.Send([]byte(resultMessage))
	}
}
