package p2p

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/tgrziminiar/pok-deng-server-engine/deck"
)

func TestPlayerActions(t *testing.T) {
	testCases := []struct {
		name          string
		room          *Room
		expectedMoney map[string]int
	}{
		{
			name: "Dealer Higher Straight",
			room: &Room{
				RoomId: "Room1",
				Players: map[string]*Player{
					"1": {isOwner: true, HandNumber: 1, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Clubs, Value: 4}, {Suit: deck.Clubs, Value: 5}, {Suit: deck.Clubs, Value: 6}},
					},
					"2": {isOwner: false, HandNumber: 2, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Clubs, Value: 1}, {Suit: deck.Clubs, Value: 2}, {Suit: deck.Clubs, Value: 3}},
					},
				},
			},
			expectedMoney: map[string]int{
				"1": 130,
				"2": 70,
			},
		},
		{
			name: "Dealer Lower Straight",
			room: &Room{
				RoomId: "Room1",
				Players: map[string]*Player{
					"1": {isOwner: true, HandNumber: 1, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Clubs, Value: 9}, {Suit: deck.Clubs, Value: 8}, {Suit: deck.Clubs, Value: 7}},
					},
					"2": {isOwner: false, HandNumber: 2, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Clubs, Value: 10}, {Suit: deck.Clubs, Value: 11}, {Suit: deck.Clubs, Value: 12}},
					},
				},
			},
			expectedMoney: map[string]int{
				"1": 70,
				"2": 130,
			},
		},
		{
			name: "Equal Straight Player Higher Suit",
			room: &Room{
				RoomId: "Room1",
				Players: map[string]*Player{
					"1": {isOwner: true, HandNumber: 1, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Clubs, Value: 9}, {Suit: deck.Clubs, Value: 8}, {Suit: deck.Clubs, Value: 7}},
					},
					"2": {isOwner: false, HandNumber: 2, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Spades, Value: 9}, {Suit: deck.Clubs, Value: 8}, {Suit: deck.Clubs, Value: 7}},
					},
				},
			},
			expectedMoney: map[string]int{
				"1": 70,
				"2": 130,
			},
		},
		{
			name: "Equal Straight Dealer Higher Suit",
			room: &Room{
				RoomId: "Room1",
				Players: map[string]*Player{
					"1": {isOwner: true, HandNumber: 1, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Spades, Value: 9}, {Suit: deck.Clubs, Value: 8}, {Suit: deck.Clubs, Value: 7}},
					},
					"2": {isOwner: false, HandNumber: 2, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Clubs, Value: 9}, {Suit: deck.Clubs, Value: 8}, {Suit: deck.Clubs, Value: 7}},
					},
				},
			},
			expectedMoney: map[string]int{
				"1": 130,
				"2": 70,
			},
		},
		{
			name: "Dealer Higher Pts",
			room: &Room{
				RoomId: "Room1",
				Players: map[string]*Player{
					"1": {isOwner: true, HandNumber: 1, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Spades, Value: 7}, {Suit: deck.Clubs, Value: 10}, {Suit: deck.Clubs, Value: 10}},
					},
					"2": {isOwner: false, HandNumber: 2, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Hearts, Value: 1}, {Suit: deck.Clubs, Value: 10}, {Suit: deck.Clubs, Value: 12}},
					},
				},
			},
			expectedMoney: map[string]int{
				"1": 110,
				"2": 90,
			},
		},
		{
			name: "Player Higher Pts",
			room: &Room{
				RoomId: "Room1",
				Players: map[string]*Player{
					"1": {isOwner: true, HandNumber: 1, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Spades, Value: 7}, {Suit: deck.Clubs, Value: 10}, {Suit: deck.Clubs, Value: 10}},
					},
					"2": {isOwner: false, HandNumber: 2, PlayerAction: &AtomicInt{1}, Money: 100, Bet: 10,
						Card: [3]deck.Card{{Suit: deck.Hearts, Value: 9}, {Suit: deck.Clubs, Value: 10}, {Suit: deck.Clubs, Value: 12}},
					},
				},
			},
			expectedMoney: map[string]int{
				"1": 90,
				"2": 110,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dealer := tc.room.Players["1"]
			dealerSpecial, dealerPts, dealerSuit, dealerDeng := dealer.CulculateThreeCard()
			slog.Info("Dealer Card", "Dealer Special", dealerSpecial, "Dealer Point", dealerPts, "Dealer Suit", dealerSuit, "Dealer Deng", dealerDeng)

			for _, player := range tc.room.Players {
				if player.isOwner {
					// s.broadcastSameMessage(room.RoomId, fmt.Sprintf("Dealer (Player1) got %s\n", player.CurrentCard()))
					continue
				}

				playerSpecial, playerPts, playerSuit, playerDeng := player.CulculateThreeCard()
				slog.Info("Player Card", "Player Special", playerSpecial, "Player Point", playerPts, "Player Suit", playerSuit, "Player Deng", playerDeng)
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
									resultMessage = fmt.Sprintf("Player%d draw with the dealer %d %d\n", player.HandNumber, playerSuit, dealerSuit)
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
				fmt.Println(resultMessage)
			}

			for playerID, expectedMoney := range tc.expectedMoney {
				if tc.room.Players[playerID].Money != expectedMoney {
					t.Errorf("Expected Player %s to have money %d, but got %d", playerID, expectedMoney, tc.room.Players[playerID].Money)
				}
			}

		})
	}
}
