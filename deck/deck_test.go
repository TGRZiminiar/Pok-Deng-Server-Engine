package deck

import (
	"fmt"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	for _, v := range deck {
		_ = v
		fmt.Println(SuitToUnicode(v.Suit), v.SpecialCardValue(v.Value))
	}
	// colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", 34, "Hello")
	// fmt.Println(colored)
}

// func TestSuitString(t *testing.T) {
// 	tests := []struct {
// 		suit Suit
// 		want string
// 	}{
// 		{Spades, "SPADES"},
// 		{Harts, "HARTS"},
// 		{Diamonds, "DIAMONDS"},
// 		{Clubs, "CLUBS"},
// 		{4, "invalid card suit"},
// 	}

// 	for _, tt := range tests {
// 		got := tt.suit.String()
// 		if got != tt.want {
// 			t.Errorf("String() = %s; want %s", got, tt.want)
// 		}
// 	}
// }
