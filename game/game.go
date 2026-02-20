package game

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	DefaultWidth  = 640
	DefaultHeight = 360
)

type Game struct {
	ticks uint64
	w, h  int
}

func New() *Game {
	return &Game{
		w: DefaultWidth,
		h: DefaultHeight,
	}
}

func (g *Game) Update() error {
	g.ticks++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 24, B: 30, A: 255})

	ebitenutil.DebugPrint(screen, "Mokosh Garden\nTicks: "+strconv.FormatUint(g.ticks, 10))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if outsideWidth > 0 && outsideHeight > 0 {
		g.w, g.h = outsideWidth, outsideHeight
		return outsideWidth, outsideHeight
	}
	return g.w, g.h
}
