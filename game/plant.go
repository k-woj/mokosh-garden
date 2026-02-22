package game

import (
	"fmt"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	plantFrameWidth      = 64
	plantFrameHeight     = 64
	plantGroundHeight    = 32
	plantGrowthSeconds   = 30
	defaultPlantTickRate = 60
	defaultTypeGrowRate  = 1.0
	minGrowthVariance    = 0.9
	maxGrowthVariance    = 1.1
)

type Plant struct {
	frames       []*ebiten.Image
	name         string
	gridX        int
	gridY        int
	x            int
	y            int
	width        int
	height       int
	elapsedTick  float64
	growthTicks  float64
	typeGrowRate float64
	growFactor   float64
	grown        bool
}

func NewPlant(frames []*ebiten.Image, name string, gridX, gridY, x, y int, growthTicks, typeGrowRate float64) *Plant {
	if growthTicks <= 0 {
		growthTicks = float64(plantGrowthSeconds * defaultPlantTickRate)
	}
	if typeGrowRate <= 0 {
		typeGrowRate = defaultTypeGrowRate
	}

	return &Plant{
		frames:       frames,
		name:         name,
		gridX:        gridX,
		gridY:        gridY,
		x:            x,
		y:            y,
		width:        plantFrameWidth,
		height:       plantFrameHeight,
		growthTicks:  growthTicks,
		typeGrowRate: typeGrowRate,
		growFactor:   randomRange(minGrowthVariance, maxGrowthVariance),
	}
}

func (p *Plant) Update(growRate float64) {
	if p.grown {
		return
	}
	if growRate <= 0 {
		growRate = 1
	}

	p.elapsedTick += p.growFactor * growRate * p.typeGrowRate
	if p.elapsedTick >= p.growthTicks {
		p.elapsedTick = p.growthTicks
		p.grown = true
	}
}

func (p *Plant) Draw(screen *ebiten.Image) {
	if len(p.frames) == 0 {
		return
	}

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(float64(p.x), float64(p.y))
	screen.DrawImage(p.frames[p.currentFrameIndex()], options)
}

func (p *Plant) Reset() {
	p.elapsedTick = 0
	p.grown = false
}

func (p *Plant) IsFullyGrown() bool {
	return p.grown
}

func (p *Plant) Contains(x, y int) bool {
	return x >= p.x && x < p.x+p.width && y >= p.y && y < p.y+p.height
}

func (p *Plant) ContainsGround(x, y int) bool {
	centerX := p.x + plantFrameWidth/2
	centerY := p.y + plantFrameHeight - (plantGroundHeight / 2)
	halfWidth := float64(plantFrameWidth) / 2.0
	halfHeight := float64(plantGroundHeight) / 2.0

	deltaX := absFloat(float64(x - centerX))
	deltaY := absFloat(float64(y - centerY))

	return (deltaX/halfWidth)+(deltaY/halfHeight) <= 1.0
}

func (p *Plant) currentFrameIndex() int {
	frameCount := len(p.frames)
	if frameCount == 0 {
		return 0
	}
	if frameCount == 1 {
		return 0
	}
	if p.grown {
		return frameCount - 1
	}
	if p.growthTicks <= 0 {
		return 0
	}

	index := int((p.elapsedTick * float64(frameCount-1)) / p.growthTicks)
	if index < 0 {
		return 0
	}
	if index >= frameCount {
		return frameCount - 1
	}

	return index
}

func (p *Plant) DebugLine(debugX, debugY int) string {
	current := p.currentFrameIndex() + 1
	total := len(p.frames)
	if total == 0 {
		total = 1
	}

	return fmt.Sprintf("[%d,%d] %s %d/%d", debugX, debugY, p.name, current, total)
}

func randomRange(min, max float64) float64 {
	if min >= max {
		return min
	}

	return min + rand.Float64()*(max-min)
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}

	return value
}
