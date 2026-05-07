package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	plantFrameWidth      = 64
	plantFrameHeight     = 64
	plantGroundHeight    = 32
	plantGrowthSeconds   = 30
	defaultPlantTickRate = 60
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
	elapsedTick  float64
	growthTicks  float64
	typeGrowRate float64
	growFactor   float64
	grown        bool
	pollenValue  float64
}

type PlantConfig struct {
	Frames       []*ebiten.Image
	Name         string
	GridX, GridY int
	X, Y         int
	GrowthTicks  float64
	TypeGrowRate float64
	PollenValue  float64
}

func NewPlant(cfg PlantConfig) *Plant {
	growthTicks := cfg.GrowthTicks
	if growthTicks <= 0 {
		growthTicks = float64(plantGrowthSeconds * defaultPlantTickRate)
	}
	typeGrowRate := cfg.TypeGrowRate
	if typeGrowRate <= 0 {
		typeGrowRate = 1.0
	}

	return &Plant{
		frames:       cfg.Frames,
		name:         cfg.Name,
		gridX:        cfg.GridX,
		gridY:        cfg.GridY,
		x:            cfg.X,
		y:            cfg.Y,
		growthTicks:  growthTicks,
		typeGrowRate: typeGrowRate,
		growFactor:   randomRange(minGrowthVariance, maxGrowthVariance),
		pollenValue:  cfg.PollenValue,
	}
}

func (p *Plant) Update(growRate float64) {
	if p.grown || growRate <= 0 {
		return
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

func (p *Plant) RandomizeGrowth() {
	// Power transform skews toward early stages: ~76% land below mid-growth.
	progress := math.Pow(rand.Float64(), 2.5)
	p.elapsedTick = progress * p.growthTicks
	p.grown = p.elapsedTick >= p.growthTicks
}

func (p *Plant) ChangeType(frames []*ebiten.Image, pt *PlantType) {
	p.frames = frames
	p.name = pt.name
	p.typeGrowRate = pt.growRate
	p.growFactor = randomRange(minGrowthVariance, maxGrowthVariance)
	p.pollenValue = pt.pollenValue
	p.elapsedTick = 0
	p.grown = false
}

func (p *Plant) IsFullyGrown() bool   { return p.grown }
func (p *Plant) Name() string         { return p.name }
func (p *Plant) PollenValue() float64 { return p.pollenValue }

func (p *Plant) Center() (float64, float64) {
	return float64(p.x + plantFrameWidth/2), float64(p.y + plantFrameHeight/2)
}

func (p *Plant) ContainsGround(x, y int) bool {
	centerX := p.x + plantFrameWidth/2
	centerY := p.y + plantFrameHeight - (plantGroundHeight / 2)
	halfWidth := float64(plantFrameWidth) / 2.0
	halfHeight := float64(plantGroundHeight) / 2.0

	deltaX := math.Abs(float64(x - centerX))
	deltaY := math.Abs(float64(y - centerY))

	return (deltaX/halfWidth)+(deltaY/halfHeight) <= 1.0
}

func (p *Plant) currentFrameIndex() int {
	frameCount := len(p.frames)
	if frameCount <= 1 {
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

func (p *Plant) DebugLine(position int) string {
	total := len(p.frames)
	if total == 0 {
		total = 1
	}
	return fmt.Sprintf("[%d] %s %d/%d", position, p.name, p.currentFrameIndex()+1, total)
}

func randomRange(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + rand.Float64()*(max-min)
}
