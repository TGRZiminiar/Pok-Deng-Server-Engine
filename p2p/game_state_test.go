package p2p

import (
	"log/slog"
	"testing"

	"github.com/tgrziminiar/pok-deng-server-engine/deck"
)

func TestCalculatePok(t *testing.T) {
	tests := []struct {
		cards   []deck.Card
		wantPts int
		wantDng int
	}{
		{[]deck.Card{{deck.Spades, 4}, {deck.Hearts, 4}}, 8, 2},
		{[]deck.Card{{deck.Spades, 10}, {deck.Hearts, 8}}, 8, 1},
		{[]deck.Card{{deck.Spades, 8}, {deck.Spades, 1}}, 9, 2},
		{[]deck.Card{{deck.Diamonds, 9}, {deck.Diamonds, 9}}, 8, 2},
		{[]deck.Card{{deck.Clubs, 7}, {deck.Clubs, 11}}, 7, 2},
	}

	for _, tt := range tests {
		t.Run("CalculatePok", func(t *testing.T) {
			gotpoint, gotDeng := calculatePok(tt.cards)
			if gotpoint != tt.wantPts || gotDeng != tt.wantDng {
				t.Errorf("got %d, %d; want %d, %d", gotpoint, gotDeng, tt.wantPts, tt.wantDng)
			}
		})
	}
}

func TestCulculateThreeCard(t *testing.T) {
	testCases := []struct {
		name        string
		playerCards []deck.Card
		expected    struct {
			isSpecial bool
			value     int
			suit      deck.Suit
			deng      int
		}
	}{
		{
			name: "Three of a Kind",
			playerCards: []deck.Card{
				{Value: 11, Suit: deck.Spades},
				{Value: 11, Suit: deck.Diamonds},
				{Value: 11, Suit: deck.Clubs},
			},
			expected: struct {
				isSpecial bool
				value     int
				suit      deck.Suit
				deng      int
			}{isSpecial: true, value: 11, suit: deck.Spades, deng: 5},
		},
		{
			name: "Normal Card",
			playerCards: []deck.Card{
				{Value: 6, Suit: deck.Spades},
				{Value: 2, Suit: deck.Diamonds},
				{Value: 3, Suit: deck.Clubs},
			},
			expected: struct {
				isSpecial bool
				value     int
				suit      deck.Suit
				deng      int
			}{isSpecial: false, value: 1, suit: deck.Spades, deng: 1},
		},
		{
			name: "Three Deng Normal Card",
			playerCards: []deck.Card{
				{Value: 6, Suit: deck.Spades},
				{Value: 1, Suit: deck.Spades},
				{Value: 11, Suit: deck.Spades},
			},
			expected: struct {
				isSpecial bool
				value     int
				suit      deck.Suit
				deng      int
			}{isSpecial: false, value: 7, suit: deck.Spades, deng: 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			player := NewPlayer(false, &Peer{}, 0)
			player.Card = [3]deck.Card(tc.playerCards)

			isSpecial, value, suit, deng := player.CulculateThreeCard()

			if isSpecial != tc.expected.isSpecial || value != tc.expected.value ||
				suit != tc.expected.suit || deng != tc.expected.deng {
				t.Errorf("Test case %q failed: got (%v, %v, %v, %v), expected (%v, %v, %v, %v)",
					tc.name, isSpecial, value, suit, deng,
					tc.expected.isSpecial, tc.expected.value, tc.expected.suit, tc.expected.deng)
			}
			slog.Info("Data -> ", "isSpecial", isSpecial, "value", value, "suit", suit, "deng", deng)
		})
	}
}

func TestIsThreeOfAKind(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []deck.Card
		expected struct {
			value   int
			suit    deck.Suit
			isThree bool
		}
	}{
		{
			name:  "Three of a Kind",
			cards: []deck.Card{{Value: 10, Suit: deck.Hearts}, {Value: 10, Suit: deck.Diamonds}, {Value: 10, Suit: deck.Clubs}},
			expected: struct {
				value   int
				suit    deck.Suit
				isThree bool
			}{value: 10, suit: deck.Hearts, isThree: true},
		},
		{
			name:  "Three of a Kind",
			cards: []deck.Card{{Value: 8, Suit: deck.Hearts}, {Value: 8, Suit: deck.Diamonds}, {Value: 8, Suit: deck.Clubs}},
			expected: struct {
				value   int
				suit    deck.Suit
				isThree bool
			}{value: 8, suit: deck.Hearts, isThree: true},
		},
		{
			name:  "Not Three of a Kind",
			cards: []deck.Card{{Value: 10, Suit: deck.Hearts}, {Value: 10, Suit: deck.Diamonds}, {Value: 9, Suit: deck.Clubs}},
			expected: struct {
				value   int
				suit    deck.Suit
				isThree bool
			}{value: 10, suit: deck.Hearts, isThree: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, suit, isThreeOfAKind := isThreeOfAKind(tc.cards)

			if value != tc.expected.value || suit != tc.expected.suit || isThreeOfAKind != tc.expected.isThree {
				t.Errorf("Test case %q failed: got (%v, %v, %v), expected (%v, %v, %v)",
					tc.name, value, suit, isThreeOfAKind, tc.expected.value, tc.expected.suit, tc.expected.isThree)
			}
			slog.Info("Data -> ", "name", tc.name, "value", value, "suit", suit, "isThreeOfAKind", isThreeOfAKind)
		})
	}
}

func TestIsThreeOfJQK(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []deck.Card
		expected struct {
			value      int
			suit       deck.Suit
			isThreeJQK bool
		}
	}{
		{
			name:  "Three of JQK",
			cards: []deck.Card{{Value: 11, Suit: deck.Hearts}, {Value: 12, Suit: deck.Diamonds}, {Value: 13, Suit: deck.Clubs}},
			expected: struct {
				value      int
				suit       deck.Suit
				isThreeJQK bool
			}{value: 13, suit: deck.Clubs, isThreeJQK: true},
		},
		{
			name:  "Three of JQK",
			cards: []deck.Card{{Value: 11, Suit: deck.Hearts}, {Value: 12, Suit: deck.Diamonds}, {Value: 12, Suit: deck.Clubs}},
			expected: struct {
				value      int
				suit       deck.Suit
				isThreeJQK bool
			}{value: 12, suit: deck.Clubs, isThreeJQK: true},
		},
		{
			name:  "Three of JQK",
			cards: []deck.Card{{Value: 11, Suit: deck.Hearts}, {Value: 11, Suit: deck.Diamonds}, {Value: 13, Suit: deck.Clubs}},
			expected: struct {
				value      int
				suit       deck.Suit
				isThreeJQK bool
			}{value: 13, suit: deck.Clubs, isThreeJQK: true},
		},
		{
			name:  "Not Three of JQK",
			cards: []deck.Card{{Value: 10, Suit: deck.Hearts}, {Value: 12, Suit: deck.Diamonds}, {Value: 13, Suit: deck.Clubs}},
			expected: struct {
				value      int
				suit       deck.Suit
				isThreeJQK bool
			}{value: 13, suit: deck.Clubs, isThreeJQK: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, suit, isThreeJQK := isThreeOfJQK(tc.cards)

			if value != tc.expected.value || suit != tc.expected.suit || isThreeJQK != tc.expected.isThreeJQK {
				t.Errorf("Test case %q failed: got (%v, %v, %v), expected (%v, %v, %v)",
					tc.name, value, suit, isThreeJQK, tc.expected.value, tc.expected.suit, tc.expected.isThreeJQK)
			}
			slog.Info("Data -> ", "name", tc.name, "value", value, "suit", suit, "isThreeOfAKind", isThreeJQK)
		})
	}
}

func TestIsStraight(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []deck.Card
		expected struct {
			value      int
			suit       deck.Suit
			isStraight bool
		}
	}{
		{
			name:  "Straight",
			cards: []deck.Card{{Value: 9, Suit: deck.Hearts}, {Value: 10, Suit: deck.Diamonds}, {Value: 11, Suit: deck.Clubs}},
			expected: struct {
				value      int
				suit       deck.Suit
				isStraight bool
			}{value: 11, suit: deck.Clubs, isStraight: true},
		},
		{
			name:  "Not Straight",
			cards: []deck.Card{{Value: 8, Suit: deck.Hearts}, {Value: 10, Suit: deck.Diamonds}, {Value: 11, Suit: deck.Clubs}},
			expected: struct {
				value      int
				suit       deck.Suit
				isStraight bool
			}{value: 11, suit: deck.Clubs, isStraight: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, suit, isStraight := isStraight(tc.cards)

			if value != tc.expected.value || suit != tc.expected.suit || isStraight != tc.expected.isStraight {
				t.Errorf("Test case %q failed: got (%v, %v, %v), expected (%v, %v, %v)",
					tc.name, value, suit, isStraight, tc.expected.value, tc.expected.suit, tc.expected.isStraight)
			}
			slog.Info("Data -> ", "name", tc.name, "value", value, "suit", suit, "isStraight", isStraight)

		})
	}
}
