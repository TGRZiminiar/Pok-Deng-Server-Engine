package p2p

import (
	"fmt"
	"log/slog"
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
	currentPeer := room.Players[p.conn.RemoteAddr().String()]
	if !currentPeer.isOwner {
		currentPeer.Send([]byte("\nPlayer1 (Dealer) is the only one who can start the game!\n"))
		return
	}

	// set the game status to be start
	// set current hand to the player 1 and deal the cards to every palyer
	room.GameState.SetStatus(GameStatusRoomStart, 2)
	if err := s.broadcastSameMessage(roomId, fmt.Sprintln("\nGame will starting in 3 seconds . . .")); err != nil {
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
			player.PlayerAction.Set(int32(PlayerActionPok))
			// check if the dealer got pok or not
			if player.isOwner && isPok {
				dealerPok = true
				room.GameState.SetStatus(GameStatusPok, 2)
				handleTwoCard(room, dealer)
				s.broadcastSameMessage(roomId, "\nSince Dealer got a pok that mean this round is end waiting for dealer to choose an option to continue the game or close the room\n")
				player.Send([]byte("\n1. /start (to continue the game)\n2. /quit to close the room\n"))
			}
		}
	}

	// if the dealer doesn't got pock everyone can ask for an extra card
	if !dealerPok {
		room.GameState.SetStatus(GameStatusExtraCard, 2)
		go s.handleExtraCardState(room, dealer)
	}

}

// for the case that dealer doesn't get pok so every body can have an extra card
func (s *Server) handleExtraCardState(room *Room, dealer *Player) {
	s.broadcastSameMessage(room.RoomId, "\ndealer doesn't got pok now it turn of player 2\n")
	time.Sleep(500 * time.Millisecond)

	go func() {
		for _, player := range room.Players {
			if player.isOwner {
				continue
			}
			currentHand := room.GameState.currentHand.Get()

			if player.PlayerAction.Get() == int32(PlayerActionPok) {
				s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nSince Player%d Pok already no need to take action\n", currentHand))
				room.handleIncrementCurrentHand()
				continue
			}

			player.Send([]byte("\nNow it your turn to made an action if you not give an action in 15 seconds you will be stay automatically\n" +
				"/stay (to don't ask for an extra card\n/more (to have 1 more extra card)\n"))

			select {
			case playerAction := <-player.PlayerActionCh:
				if playerAction.Action == "stay" {
					s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nPlayer%d  choose to stay\nCard of Player%d are: %s\n", currentHand, currentHand, player.CurrentCard()))
					player.PlayerAction.Set(int32(PlayerActionStay))
					room.handleIncrementCurrentHand()

				} else if playerAction.Action == "more" {
					player.Card[2] = room.Deck[0]
					room.Deck = room.Deck[1:]
					s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nPlayer%d  ask for extra card\nnow Player %d got 3 cards\nCard of Player%d are: %s\n", currentHand, currentHand, currentHand, player.CurrentCard()))
					player.PlayerAction.Set(int32(PlayerActionExtraCard))
					room.handleIncrementCurrentHand()
				}
			case <-time.After(15 * time.Second):
				s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nPlayer%d doesn't take any action so ,Player%d will be stay\n", currentHand, currentHand))
				player.PlayerAction.Set(int32(PlayerActionStay))
				room.handleIncrementCurrentHand()
			}
		}

		currentHand := room.GameState.currentHand.Get()
		if currentHand == 1 {
			dealer.Send([]byte("\nNow it your turn to made an action if you not give an action in 15 seconds you will be stay automatically\n" +
				"/stay (to don't ask for an extra card\n/more (to have 1 more extra card)\n"))

			select {
			case playerAction := <-dealer.PlayerActionCh:
				if playerAction.Action == "stay" {
					s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nDealer choose to stay\nCard of Dealer are: %s\n", dealer.CurrentCard()))
					dealer.PlayerAction.Set(int32(PlayerActionStay))

				} else if playerAction.Action == "more" {
					dealer.Card[2] = room.Deck[0]
					room.Deck = room.Deck[1:]
					s.broadcastSameMessage(room.RoomId, fmt.Sprintf("\nDealer ask for extra card\nnow Dealer got 3 cards\nCard of Dealer are: %s\n", dealer.CurrentCard()))
					dealer.PlayerAction.Set(int32(PlayerActionExtraCard))
				}
			case <-time.After(15 * time.Second):
				s.broadcastSameMessage(room.RoomId, "\nDealer doesn't take any action so ,Dealer will be stay\n")
				dealer.PlayerAction.Set(int32(PlayerActionStay))
			}
		} else {
			s.broadcastSameMessage(room.RoomId, fmt.Sprintln("Something went wrong with this room closing in 3 seconds"))

		}
		// all the player can't take anymore action
		s.handleThreeCard(room, dealer)

	}()
}

func (s *Server) handlePlayerStay(p *Peer) {
	roomId, ok := s.playerInAnyRoom(p)
	if !ok {
		return
	}
	room := s.rooms[roomId]
	id := p.conn.RemoteAddr().String()
	player := room.Players[id]
	if player.HandNumber == int(room.GameState.currentHand.Get()) {
		room.Players[id].PlayerActionCh <- &PlayerActionMsg{PlayerID: id, Action: "stay"}
	} else {
		player.Send([]byte(fmt.Sprintf("Current Hand is %d you are not allowed to make an action now\nYour handnumber are %d", room.GameState.currentHand.Get(), player.HandNumber)))
	}
}
func (s *Server) handlePlayerMoreExtraCard(p *Peer) {
	roomId, ok := s.playerInAnyRoom(p)
	if !ok {
		return
	}
	room := s.rooms[roomId]
	id := p.conn.RemoteAddr().String()
	player := room.Players[id]
	if player.HandNumber == int(room.GameState.currentHand.Get()) {
		room.Players[id].PlayerActionCh <- &PlayerActionMsg{PlayerID: id, Action: "more"}
	} else {
		player.Send([]byte(fmt.Sprintf("Current Hand is %d you are not allowed to make an action now\nYour handnumber are %d", room.GameState.currentHand.Get(), player.HandNumber)))
	}
}

// return a string of a current value and suit of the card that you holding
func (s *Server) CurrentCard(p *Peer, roomId string) string {
	room := s.rooms[roomId]
	player := room.Players[p.conn.RemoteAddr().String()]
	return player.CurrentCard()
}

// Send out the current game status and current hand
func (s *Server) handleCurrentGame(p *Peer) {
	roomId, ok := s.playerInAnyRoom(p)
	if !ok {
		return
	}

	room := s.rooms[roomId]

	gameStatus := room.GameState.currentGameStatus()
	currentHand := room.GameState.currentHand.Get()
	text := fmt.Sprintf(
		"\nGame Status: %s\nYou are Player%d\nCurrent Hand: Player%d\n",
		gameStatus,
		room.Players[p.conn.RemoteAddr().String()].HandNumber,
		currentHand)
	p.Send([]byte(text))

}

// handle when player got three cards the test file is on the server_game_test
func (s *Server) handleThreeCard(room *Room, dealer *Player) {
	dealerSpecial, dealerPts, dealerSuit, dealerDeng := dealer.CulculateThreeCard()
	for _, player := range room.Players {
		if player.isOwner {
			// s.broadcastSameMessage(room.RoomId, fmt.Sprintf("Dealer (Player1) got %s\n", player.CurrentCard()))
			continue
		}

		if player.PlayerAction.Get() == int32(PlayerActionPok) {
			_, _, deng := player.CulculateTwoCard()
			s.broadcastSameMessage(room.RoomId, fmt.Sprintf(
				"Since Player%d got pok %s earn %d Player%d current balance is %d\n",
				player.HandNumber,
				player.CurrentCard(),
				player.Bet*deng,
				player.HandNumber,
				player.Money))
			continue
		}

		playerSpecial, playerPts, playerSuit, playerDeng := player.CulculateThreeCard()
		resultMessage := ""
		// dealerSpecial == true
		// some special case here
		// TODO: handle special case here
		if dealerSpecial != NotSpecialWinning {

			if playerDeng > dealerDeng {
				player.Money += player.Bet * playerDeng
				dealer.Money -= player.Bet * playerDeng
				resultMessage = fmt.Sprintf(
					"Player%d got %s won to the dealer by %s earn %d Player%d current balance is %d\n",
					player.HandNumber,
					player.CurrentCard(),
					playerSpecial.String(),
					player.Bet*playerDeng,
					player.HandNumber,
					player.Money)
			} else if dealerDeng > playerDeng {
				player.Money += player.Bet * playerDeng
				dealer.Money -= player.Bet * playerDeng
				resultMessage = fmt.Sprintf(
					"Player%d got %s lost to the dealer by %s earn %d Player%d current balance is %d\n",
					player.HandNumber,
					player.CurrentCard(),
					playerSpecial.String(),
					player.Bet*playerDeng,
					player.HandNumber,
					player.Money)
				// if deng is equal and the type of the winning is not the same it's drawn
			} else if dealerDeng == playerDeng {
				if dealerSpecial == playerSpecial {
					if playerPts > dealerPts {
						player.Money += player.Bet * playerDeng
						dealer.Money -= player.Bet * playerDeng
						resultMessage = fmt.Sprintf(
							"Player%d got %s won to the dealer by %s earn %d Player%d current balance is %d\n",
							player.HandNumber,
							player.CurrentCard(),
							playerSpecial.String(),
							player.Bet*playerDeng,
							player.HandNumber,
							player.Money)
					} else if playerPts < dealerPts {
						player.Money -= player.Bet * playerDeng
						dealer.Money += player.Bet * playerDeng
						resultMessage = fmt.Sprintf(
							"Player%d got %s lost to the dealer by %s earn %d Player%d current balance is %d\n",
							player.HandNumber,
							player.CurrentCard(),
							playerSpecial.String(),
							player.Bet*playerDeng,
							player.HandNumber,
							player.Money)
					} else if playerPts == dealerPts {
						if playerSuit > dealerSuit {
							player.Money += player.Bet * playerDeng
							dealer.Money -= player.Bet * playerDeng
							resultMessage = fmt.Sprintf(
								"Player%d got %s won to the dealer by %s earn %d Player%d current balance is %d\n",
								player.HandNumber,
								player.CurrentCard(),
								playerSpecial.String(),
								player.Bet*playerDeng,
								player.HandNumber,
								player.Money)
						} else if dealerSuit > playerSuit {
							player.Money -= player.Bet * playerDeng
							dealer.Money += player.Bet * playerDeng
							resultMessage = fmt.Sprintf(
								"Player%d got %s won to the dealer by %s earn %d Player%d current balance is %d\n",
								player.HandNumber,
								player.CurrentCard(),
								playerSpecial.String(),
								player.Bet*playerDeng,
								player.HandNumber,
								player.Money)
						} else {
							resultMessage = fmt.Sprintf("Player%d draw with the dealer\n", player.HandNumber)

						}

					}
				}
			}

		} else {
			// if player got special and dealer is not special
			if playerSpecial != NotSpecialWinning {
				player.Money += player.Bet * playerDeng
				dealer.Money -= player.Bet * playerDeng
				resultMessage = fmt.Sprintf(
					"Player%d got %s won to the dealer by %s earn %d Player%d current balance is %d\n",
					player.HandNumber,
					player.CurrentCard(),
					playerSpecial.String(),
					player.Bet*playerDeng,
					player.HandNumber,
					player.Money)

			} else {
				if dealerPts > playerPts {
					player.Money -= player.Bet * playerDeng
					dealer.Money += player.Bet * playerDeng
					resultMessage = fmt.Sprintf(
						"Player%d got %s lost to the dealer: %d Player%d current balance is %d\n",
						player.HandNumber,
						player.CurrentCard(),
						player.Bet*playerDeng,
						player.HandNumber,
						player.Money)
				} else if dealerPts == playerPts && dealerDeng == playerDeng {
					resultMessage = fmt.Sprintf("Player%d draw with the dealer\n", player.HandNumber)
				} else if dealerPts == playerPts && dealerDeng > playerDeng {
					player.Money -= player.Bet * (dealerDeng - playerDeng)
					dealer.Money += player.Bet * (dealerDeng - playerDeng)
					resultMessage = fmt.Sprintf(
						"Player%d got %s lost to the dealer deng: %d Player%d current balance is %d\n",
						player.HandNumber,
						player.CurrentCard(),
						player.Bet*(dealerDeng-playerDeng),
						player.HandNumber,
						player.Money)

				} else if dealerPts == playerPts && dealerDeng < playerDeng {
					player.Money += player.Bet * (playerDeng - dealerDeng)
					dealer.Money -= player.Bet * (playerDeng - dealerDeng)
					resultMessage = fmt.Sprintf(
						"Player%d got %s win to the dealer deng: %d Player%d current balance is %d\n",
						player.HandNumber,
						player.CurrentCard(),
						player.Bet*(playerDeng-dealerDeng),
						player.HandNumber,
						player.Money)

				} else if playerPts > dealerPts {
					player.Money += player.Bet * playerDeng
					dealer.Money -= player.Bet * playerDeng
					resultMessage = fmt.Sprintf(
						"Player%d got %s won to the dealer points: %d Player%d current balance is %d\n",
						player.HandNumber,
						player.CurrentCard(),
						player.Bet*playerDeng,
						player.HandNumber,
						player.Money)
				}

			}

		}
		s.broadcastSameMessage(room.RoomId, resultMessage)
	}
	room.GameState.SetStatus(GameStatusEnd, 1)
	s.broadcastSameMessage(room.RoomId, "\nSince the game is end dealer have 2 choices\n1. /start (to continue the game)\n2. /quit to close the room\n")
}

// handle for the case of 2 cards only because if someone got three card
// we need to check for more condition like is it straight or three of a kind and many more
func handleTwoCard(room *Room, dealer *Player) {
	dealerPok, dealerPts, dealerDeng := dealer.CulculateTwoCard()
	for _, player := range room.Players {
		if player.isOwner {
			continue
		}

		playerPok, playerPts, playerDeng := player.CulculateTwoCard()

		if player.PlayerAction.Get() == int32(PlayerActionPok) {
			player.Money += player.Bet * playerDeng
			dealer.Money -= player.Bet * playerDeng
			continue
		}

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
