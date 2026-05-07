package game

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	achievePanelW = 112
	achieveRowH   = 13

	trophyIconW = 14
	trophyIconH = 14
	trophyIconY = 14
)

type Achievement struct {
	label     string
	condition func(*Game) bool
}

func makeAchievements(visitors []*Visitor) []*Achievement {
	goat := visitorByName(visitors, "Goat")
	deer := visitorByName(visitors, "Deer")
	butterfly := visitorByName(visitors, "Butterfly")
	return []*Achievement{
		{"50 Honey", func(g *Game) bool { return g.honey >= 50 }},
		{"200 Honey", func(g *Game) bool { return g.honey >= 200 }},
		{"1000 Honey", func(g *Game) bool { return g.honey >= 1000 }},
		{"Full Garden", func(g *Game) bool { return g.fullGardenAchieved }},
		{"20 Bees", func(g *Game) bool { return g.pickerUnlocked }},
		{"Goat Visit", func(*Game) bool { return goat.seen }},
		{"Deer Visit", func(*Game) bool { return deer.seen }},
		{"Butterfly", func(*Game) bool { return butterfly.seen }},
	}
}

func (g *Game) drawTrophyIcon(screen *ebiten.Image) {
	if g.trophyImg == nil {
		return
	}
	b := g.trophyImg.Bounds()
	x := float64(g.w - trophyIconW - 2)
	y := float64(trophyIconY)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(trophyIconW)/float64(b.Dx()), float64(trophyIconH)/float64(b.Dy()))
	op.GeoM.Translate(x, y)
	if g.showAchievements {
		op.ColorScale.Scale(1.5, 1.5, 1.0, 1.0)
	}
	screen.DrawImage(g.trophyImg, op)
}

func (g *Game) drawAchievements(screen *ebiten.Image) {
	panelX := float64(g.w - achievePanelW)
	panelY := 14.0
	panelH := float64(achieveRowH*(len(g.achievements)+1) + 4)

	ebitenutil.DrawRect(screen, panelX, panelY, float64(achievePanelW), panelH, color.RGBA{0, 0, 0, 190})

	tx := g.w - achievePanelW + 3
	ebitenutil.DebugPrintAt(screen, "Achievements", tx, int(panelY)+1)

	for i, a := range g.achievements {
		y := int(panelY) + achieveRowH*(i+1) + 2
		if a.condition(g) {
			ebitenutil.DebugPrintAt(screen, "[x] "+a.label, tx, y)
		} else {
			ebitenutil.DebugPrintAt(screen, "[?] "+strings.Repeat("?", len(a.label)), tx, y)
		}
	}
}
