package main

import (
	"fmt"
	"github.com/frncscsrcc/baassdk"
	"os"
)

type X struct{}

func (x X) Play(cards [3]string, previousCard string) string {
	show(cards)
	show(previousCard)
	cardToPlay := "X" // Here goes the algorithm
	return cardToPlay
}

func main() {
	var x X
	game := baassdk.NewGame(os.Args[1], "http://localhost:8082", x)
	_, err := game.Play()
	if err != nil {
		show(err)
	}
}

func show(i interface{}) {
	fmt.Printf("*** %+v\n", i)
}
