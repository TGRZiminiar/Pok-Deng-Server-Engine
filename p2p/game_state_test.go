package p2p

import (
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
			gotpoint, gotDeng := CalculatePok(tt.cards)
			if gotpoint != tt.wantPts || gotDeng != tt.wantDng {
				t.Errorf("got %d, %d; want %d, %d", gotpoint, gotDeng, tt.wantPts, tt.wantDng)
			}
		})
	}
}
