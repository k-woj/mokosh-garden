package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	pickerLongClickTicks = int(defaultPlantTickRate * 0.5) // hold for 0.5s
	pickerRingRadius     = 40.0                            // distance from plant centre to each option centre
	pickerCircleRadius   = 18.0                            // radius of each option circle
	pickerPreviewSize    = 28.0                            // plant preview drawn inside the circle
)

type pickerOption struct {
	typeName string
	x, y     float64
	preview  *ebiten.Image
}

type PlantPicker struct {
	active     bool
	plant      *Plant
	options    []pickerOption
	hoveredIdx int
}

func newPlantPicker() *PlantPicker {
	return &PlantPicker{hoveredIdx: -1}
}

func (pp *PlantPicker) Open(plant *Plant, plantTypes map[string]*PlantType, screenW, screenH int) {
	pp.active = true
	pp.plant = plant
	pp.hoveredIdx = -1

	cx, cy := plant.Center()

	pp.options = make([]pickerOption, len(plantTypeNames))
	for i, name := range plantTypeNames {
		// Evenly spaced around the ring, starting at top (−π/2).
		angle := float64(i)*math.Pi*2/float64(len(plantTypeNames)) - math.Pi/2
		ox := cx + math.Cos(angle)*pickerRingRadius
		oy := cy + math.Sin(angle)*pickerRingRadius

		// Clamp so circles stay fully on screen.
		margin := pickerCircleRadius + 1
		ox = clampF(ox, margin, float64(screenW)-margin)
		oy = clampF(oy, margin, float64(screenH)-margin)

		var preview *ebiten.Image
		if pt, ok := plantTypes[name]; ok {
			preview = pt.previewImg
		}
		pp.options[i] = pickerOption{typeName: name, x: ox, y: oy, preview: preview}
	}
}

func (pp *PlantPicker) Close() {
	pp.active = false
	pp.plant = nil
	pp.options = nil
	pp.hoveredIdx = -1
}

func (pp *PlantPicker) Update(mx, my int) {
	pp.hoveredIdx = -1
	for i, opt := range pp.options {
		dx := float64(mx) - opt.x
		dy := float64(my) - opt.y
		if dx*dx+dy*dy <= pickerCircleRadius*pickerCircleRadius {
			pp.hoveredIdx = i
			break
		}
	}
}

// HoveredType returns the plant type name under the cursor, if any.
func (pp *PlantPicker) HoveredType() (string, bool) {
	if pp.hoveredIdx < 0 {
		return "", false
	}
	return pp.options[pp.hoveredIdx].typeName, true
}

func (pp *PlantPicker) Draw(screen *ebiten.Image) {
	if !pp.active {
		return
	}
	for i, opt := range pp.options {
		hovered := i == pp.hoveredIdx

		bgCol := color.RGBA{30, 25, 15, 210}
		if hovered {
			bgCol = color.RGBA{80, 65, 20, 230}
		}
		vector.DrawFilledCircle(screen, float32(opt.x), float32(opt.y), float32(pickerCircleRadius), bgCol, true)

		borderCol := color.RGBA{160, 140, 60, 255}
		if hovered {
			borderCol = color.RGBA{255, 220, 80, 255}
		}
		vector.StrokeCircle(screen, float32(opt.x), float32(opt.y), float32(pickerCircleRadius), 1.5, borderCol, true)

		if opt.preview != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(opt.x-pickerPreviewSize/2, opt.y-pickerPreviewSize/2)
			screen.DrawImage(opt.preview, op)
		}
	}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
