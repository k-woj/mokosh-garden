//go:build js && wasm

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vistustan/mokosh-garden/game"
)

func main() {
	g, err := game.New(assets)
	if err != nil {
		log.Fatal(err)
	}

	width, height := g.Size()
	ebiten.SetWindowSize(width*2, height*2)
	ebiten.SetWindowTitle("Mokosh Garden")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
