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
	room.GameState.SetStatus(GameStatusRoomStart, 2)
	if err := s.broadcastSameMessage(roomId, fmt.Sprintf("\nGame will starting in 3 seconds . . .")); err != nil {
		slog.Error(err.Error())
		return
	}

	// sleep for a bit cause the message will be too quick
	time.Sleep(500 * time.Millisecond)
	for i := 1; i > 0; i-- {
		if err := s.broadcastSameMessage(roomId, fmt.Sprintf("\nGame starting in %d seconds\n", i)); err != nil {
			slog.Error(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	// Give random cards to the user and shuffle the card
	// set the dealer to use to check if the game status going to be pok or not
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
				room.GameState.SetStatus(GameStatusPok, 2)
				handleTwoCard(room, dealer)
				s.broadcastSameMessage(roomId, "\nSince Dealer got a pok that mean this round is end waiting for dealer to choose an option to continue the game or close the room\n")
				player.Send([]byte("\n1. /start (to continue the game)\n2. /quit to close the room\n"))
			}
		}
	}
	if !dealerPok {
		room.GameState.SetStatus(GameStatusExtraCard, 2)
		s.broadcastSameMessage(roomId, "\ndealer doesn't got pok now it turn of player 1\n")
		time.Sleep(500 * time.Millisecond)
		go func() {
			var index = 0
			for _, v := range room.Players {
				if v.isOwner {
					continue
				}
				currentHand := room.GameState.currentHand.Get()
				// if index == 1 {
				fmt.Println("this is the current hand -> ", currentHand)
				v.Send([]byte("\nNow it your turn to made an action if you not give an action in 15 seconds you will be stay automatically\n" +
					"/stay (to don't ask for an extra card\n/more (to have 1 more extra card)\n"))

				select {
				case playerAction := <-v.PlayerActionCh:
					if playerAction.Action == "stay" {
						s.broadcastSameMessage(roomId, fmt.Sprintf("player%d  choose to stay\n", currentHand))
						v.PlayerAction.Set(int32(PlayerActionStay))
						room.GameState.currentHand.Inc()
					} else if playerAction.Action == "more" {
						s.broadcastSameMessage(roomId, fmt.Sprintf("player%d  ask for extra card, now player %d got 3 cards\n", currentHand, currentHand))
						v.PlayerAction.Set(int32(PlayerActionStay))
						room.GameState.currentHand.Inc()
					}
				case <-time.After(15 * time.Second):
					s.broadcastSameMessage(roomId, fmt.Sprintf("player%d  doesn't take any action so player%d will be stay\n", currentHand, currentHand))
					v.PlayerAction.Set(int32(PlayerActionStay))
					room.GameState.currentHand.Inc()
				}
				index++
				// }

			}
		}()
	}
}

func (s *Server) handlePlayerStay(p *Peer) {
	roomId, ok := s.playerInAnyRoom(p)
	if !ok {
		return
	}
	room := s.rooms[roomId]
	id := p.conn.RemoteAddr().String()
	room.Players[id].PlayerActionCh <- &PlayerActionMsg{PlayerID: id, Action: "stay"}
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
	currentHand := room.GameState.currentHand.Get()
	text := fmt.Sprintf("\nGame Status: %s%d\nCurrent Hand: player%d\n", gameStatus, room.GameState.gameStatus.Get(), currentHand)
	p.Send([]byte(text))
}

func handleTwoCard(room *Room, dealer *Player) {
	dealerPok, dealerPts, dealerDeng := dealer.CulculateTwoCard()
	for _, player := range room.Players {
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
