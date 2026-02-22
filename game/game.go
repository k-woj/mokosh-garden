package game

import (
	"image/color"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	DefaultWidth  = 640
	DefaultHeight = 360

	groundCenterOffsetX = 0
	groundCenterOffsetY = 6
	plantRowUpOffset    = 0
	plantFrameCount     = 16
	defaultGrowRate     = 0.95
)

type Game struct {
	w, h         int
	bg           *ebiten.Image
	growRate     float64
	typeGrowRate map[string]float64
	debugEnabled bool
	plants       []*Plant
}

func New() (*Game, error) {
	rand.Seed(time.Now().UnixNano())

	background, width, height, err := loadBackgroundImage(filepath.Join("assets", "garden.png"))
	if err != nil {
		return nil, err
	}
	plantFrames, err := loadSpriteSheetFrames(filepath.Join("assets", "wheet.png"), plantFrameWidth, plantFrameHeight, plantFrameCount)
	if err != nil {
		return nil, err
	}
	typeGrowRate := map[string]float64{
		"Wheet": 1.0,
	}

	growthTicks := float64(plantGrowthSeconds * int(ebiten.TPS()))
	if growthTicks <= 0 {
		growthTicks = float64(plantGrowthSeconds * defaultPlantTickRate)
	}

	groundCenterX := width/2 + groundCenterOffsetX
	groundCenterY := height/2 + groundCenterOffsetY
	centerGroundX := groundCenterX
	centerGroundY := groundCenterY - plantRowUpOffset

	type plantSlot struct {
		gridX   int
		gridY   int
		offsetX int
		offsetY int
	}

	edgeSlots := []plantSlot{
		{gridX: 0, gridY: 0, offsetX: 1, offsetY: -26},
		{gridX: 1, gridY: 0, offsetX: 30, offsetY: -14},
		{gridX: 2, gridY: 0, offsetX: -30, offsetY: -13},
		{gridX: 0, gridY: 1, offsetX: -61, offsetY: -2},
		{gridX: 1, gridY: 1, offsetX: 2, offsetY: 3},
		{gridX: 2, gridY: 1, offsetX: 64, offsetY: 0},
		{gridX: 0, gridY: 2, offsetX: -30, offsetY: 14},
		{gridX: 1, gridY: 2, offsetX: 33, offsetY: 15},
		{gridX: 2, gridY: 2, offsetX: 2, offsetY: 30},
	}

	plants := make([]*Plant, 0, len(edgeSlots))
	for _, slot := range edgeSlots {
		groundX := centerGroundX + slot.offsetX
		groundY := centerGroundY + slot.offsetY
		plantX, plantY := spritePositionForGroundCenter(
			groundX,
			groundY,
			plantFrameWidth,
			plantFrameHeight,
			plantGroundHeight,
		)
		plant := NewPlant(plantFrames, "Wheet", slot.gridX, slot.gridY, plantX, plantY, growthTicks, typeGrowRate["Wheet"])
		plants = append(plants, plant)
	}

	sort.Slice(plants, func(i, j int) bool {
		if plants[i].y != plants[j].y {
			return plants[i].y < plants[j].y
		}
		return plants[i].x < plants[j].x
	})

	return &Game{w: width, h: height, bg: background, growRate: defaultGrowRate, typeGrowRate: typeGrowRate, debugEnabled: false, plants: plants}, nil
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.debugEnabled = !g.debugEnabled
	}

	for _, plant := range g.plants {
		plant.Update(g.growRate)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursorX, cursorY := ebiten.CursorPosition()
		for _, plant := range g.plants {
			if plant.IsFullyGrown() && plant.ContainsGround(cursorX, cursorY) {
				plant.Reset()
				break
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.bg != nil {
		screen.DrawImage(g.bg, nil)
	} else {
		screen.Fill(color.RGBA{R: 20, G: 24, B: 30, A: 255})
	}

	for _, plant := range g.plants {
		plant.Draw(screen)
	}

	if g.debugEnabled {
		ebitenutil.DebugPrint(screen, g.plantsDebugText())
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.w, g.h
}

func (g *Game) Size() (int, int) {
	return g.w, g.h
}

func spritePositionForGroundCenter(centerX, centerY, frameWidth, frameHeight, groundHeight int) (int, int) {
	groundHalfHeight := groundHeight / 2
	anchorX := frameWidth / 2
	anchorY := frameHeight - groundHalfHeight

	return centerX - anchorX, centerY - anchorY
}

func (g *Game) plantsDebugText() string {
	var builder strings.Builder
	builder.WriteString("Debug:")

	rowIndex := -1
	columnIndex := 0
	lastY := 0

	for index, plant := range g.plants {
		if index == 0 || plant.y != lastY {
			rowIndex++
			columnIndex = 0
			lastY = plant.y
		}

		builder.WriteString("\n")
		builder.WriteString(plant.DebugLine(columnIndex, rowIndex))
		columnIndex++
	}

	return builder.String()
}
