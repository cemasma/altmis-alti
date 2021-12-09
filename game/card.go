package game

import (
	"math/rand"
	"time"
)

var (
	cardTypes = [4]string{"CLUB", "DIAMOND", "HEART", "SPADE"}
	cards = [6]string{"1", "KING", "QUEEN", "JACK", "10", "9"}
	values = [6]uint8{11, 4, 3, 2, 10, 0}
)

type Card struct {
	Kind string
	CardName string
	Value uint8
}

func PrepareCards() []Card {
	cardArr := make([]Card, 0)

	for _, value := range cardTypes {
		for cardIndex, cardValue := range cards {
			cardArr = append(cardArr, Card{
				Kind:     value,
				CardName: cardValue,
				Value:    values[cardIndex],
			})
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cardArr), func(i, j int) {
		cardArr[i], cardArr[j] = cardArr[j], cardArr[i]
	})

	return cardArr
}