package deck

import (
	"fmt"
	"math/rand"
	"strconv"
)

type Suit int

func (s Suit) String() string {
	switch s {
	case Spades:
		return "SPADES"
	case Hearts:
		return "Hearts"
	case Diamonds:
		return "DIAMONDS"
	case Clubs:
		return "CLUBS"
	default:
		return "invalid card suit"
		// panic()
	}
}

const (
	Spades   Suit = iota // 0
	Hearts               // 1
	Diamonds             // 2
	Clubs                // 3
)

type Card struct {
	Suit  Suit
	Value int
}

func (c Card) String() string {
	value := strconv.Itoa(c.Value)
	if c.Value == 1 {
		value = "ACE"
	}

	return fmt.Sprintf("%s of %s %s", value, c.Suit, SuitToUnicode(c.Suit))
}
func (c Card) SpecialCardValue(val int) string {
	switch val {
	case 1:
		return "ACE"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	return strconv.Itoa(val)
}

func NewCard(s Suit, v int) Card {
	if v > 13 {
		panic("the value of the card cannot be higher then 13")
	}

	return Card{
		Suit:  s,
		Value: v,
	}
}

type Deck []Card

func NewDeck() Deck {
	var (
		nSuits = 4
		nCards = 13
	)
	d := make([]Card, 52)

	x := 0
	for i := 0; i < nSuits; i++ {
		for j := 0; j < nCards; j++ {
			d[x] = NewCard(Suit(i), j+1)
			x++
		}
	}
	return d
	// return Shuffle(d)
}

func Shuffle(d Deck) Deck {
	for i := 0; i < len(d); i++ {
		r := rand.Intn(i + 1)

		if r != i {
			d[i], d[r] = d[r], d[i]
		}
	}

	return d
}

func SuitToUnicode(s Suit) string {
	switch s {
	case Spades:
		return handleColor(34, "♠")
	case Hearts:
		return handleColor(31, "❤️")
	case Diamonds:
		return handleColor(31, "♦")
	case Clubs:
		return handleColor(34, "♣")
	default:
		panic("invalid card suit")
	}
}

// 30: Black
// 31: Red
// 32: Green
// 33: Yellow
// 34: Blue
// 35: Magenta
// 36: Cyan
// 37: White
func handleColor(ansicode int, text string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ansicode, text)
}
