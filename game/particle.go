package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxParticles       = 40
	particleLifetime   = 30
	particleDriftY     = -0.3
	particleEmitChance = 0.2
)

type Particle struct {
	x, y   float64
	vx, vy float64
	life   int
}

func newParticle(x, y float64) *Particle {
	return &Particle{
		x:    x + randomRange(-3, 3),
		y:    y + randomRange(-3, 3),
		vx:   randomRange(-0.3, 0.3),
		vy:   particleDriftY + randomRange(-0.1, 0),
		life: particleLifetime,
	}
}

func (p *Particle) Update() {
	p.x += p.vx
	p.y += p.vy
	p.life--
}

func (p *Particle) Done() bool {
	return p.life <= 0
}

func (p *Particle) Draw(screen *ebiten.Image) {
	alpha := uint8(200 * p.life / particleLifetime)
	screen.Set(int(p.x), int(p.y), color.RGBA{255, 220, 80, alpha})
}
