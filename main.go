//go:build !js && !wasm && !android && !ios

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/vistustan/mokosh-garden/game"
)

func main() {
	ebiten.SetWindowTitle("Mokosh Garden")
	ebiten.SetWindowSize(game.DefaultWidth*2, game.DefaultHeight*2)

	if err := ebiten.RunGame(game.New()); err != nil {
		log.Fatal(err)
	}
}
