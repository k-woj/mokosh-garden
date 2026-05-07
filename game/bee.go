package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	maxBeeCount         = 22
	beeBaseSpeed        = 0.7
	beeSpeedVariance    = 0.3
	beeWobbleSpeed      = 0.35
	beeTargetRadius     = 1.5
	beeFlowerAttraction = 0.7
	beeGoHomeChance     = 0.0002
	beePollenRadius     = 12.0
	beePollenRate       = 0.02
	beePollenMax        = 2.5

	beeWanderMinX = 50.0
	beeWanderMaxX = 270.0
	beeWanderMinY = 30.0
	beeWanderMaxY = 140.0

	// Visible front entry: small quadrilateral (45,154), (45,152), (48,150), (48,152).
	hiveFrontEntryX      = 46.5
	hiveFrontEntryY      = 152.0
	hiveFrontEntryJitter = 1.0

	// Hidden side entry: vertical line from (54,137) to (54,157).
	hiveHiddenEntryX    = 54.0
	hiveHiddenEntryMinY = 137.0
	hiveHiddenEntryMaxY = 157.0
)

type Bee struct {
	x, y      float64
	targetX   float64
	targetY   float64
	speed     float64
	wobble    float64
	pollen    float64
	goingHome bool
	done      bool
}

func newBee(x, y float64) *Bee {
	return &Bee{
		x:       x,
		y:       y,
		targetX: x,
		targetY: y,
		speed:   beeBaseSpeed + randomRange(-beeSpeedVariance, beeSpeedVariance),
		wobble:  rand.Float64() * math.Pi * 2,
	}
}

// newBeeAtHive spawns a bee emerging from one of the two hive entries.
func newBeeAtHive() *Bee {
	x, y := pickHiveEntry()
	return newBee(x, y)
}

// pickHiveEntry randomly chooses between the front (visible) and hidden side entry.
func pickHiveEntry() (float64, float64) {
	if rand.Float64() < 0.5 {
		return hiveFrontEntryX + randomRange(-hiveFrontEntryJitter, hiveFrontEntryJitter),
			hiveFrontEntryY + randomRange(-hiveFrontEntryJitter, hiveFrontEntryJitter)
	}
	return hiveHiddenEntryX, randomRange(hiveHiddenEntryMinY, hiveHiddenEntryMaxY)
}

func (b *Bee) sendHome() {
	b.goingHome = true
	b.targetX, b.targetY = pickHiveEntry()
}

func (b *Bee) collectPollen(plants []*Plant) {
	for _, p := range plants {
		if !p.IsFullyGrown() {
			continue
		}
		cx, cy := p.Center()
		dx := cx - b.x
		dy := cy - b.y
		if dx*dx+dy*dy < beePollenRadius*beePollenRadius {
			b.pollen += beePollenRate * p.PollenValue()
			if b.pollen >= beePollenMax && !b.goingHome {
				b.sendHome()
			}
			return
		}
	}
}

func (b *Bee) Update(plants []*Plant) {
	if b.done {
		return
	}

	b.collectPollen(plants)

	if !b.goingHome && rand.Float64() < beeGoHomeChance {
		b.sendHome()
	}

	dx := b.targetX - b.x
	dy := b.targetY - b.y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < beeTargetRadius {
		if b.goingHome {
			b.done = true
			return
		}
		b.pickTarget(plants)
	} else {
		b.x += (dx / dist) * b.speed
		b.y += (dy / dist) * b.speed
	}

	b.wobble += beeWobbleSpeed
}

func (b *Bee) pickTarget(plants []*Plant) {
	var grown []*Plant
	for _, p := range plants {
		if p.IsFullyGrown() {
			grown = append(grown, p)
		}
	}

	if len(grown) > 0 && rand.Float64() < beeFlowerAttraction {
		cx, cy := grown[rand.Intn(len(grown))].Center()
		b.targetX = cx + randomRange(-12, 12)
		b.targetY = cy + randomRange(-12, 12)
	} else {
		b.targetX = randomRange(beeWanderMinX, beeWanderMaxX)
		b.targetY = randomRange(beeWanderMinY, beeWanderMaxY)
	}
}

func (b *Bee) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, b.x-1, b.y-1, 2, 2, color.RGBA{240, 195, 0, 255})
}
