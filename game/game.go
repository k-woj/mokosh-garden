package game

import (
	"fmt"
	"image/color"
	"io/fs"
	"math/rand"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	DefaultWidth  = 640
	DefaultHeight = 360

	groundCenterOffsetY = 6
	plantFrameCount     = 16
	defaultGrowRate     = 0.95
)

type PlantType struct {
	name        string
	variants    [][]*ebiten.Image
	previewImg  *ebiten.Image
	growRate    float64
	pollenValue float64
}

func (pt *PlantType) randomFrames() []*ebiten.Image {
	return pt.variants[rand.Intn(len(pt.variants))]
}

type Game struct {
	w, h                      int
	bg                        *ebiten.Image
	trophyImg                 *ebiten.Image
	growRate                  float64
	plantTypes                map[string]*PlantType
	debugEnabled              bool
	plants                    []*Plant
	bees                      []*Bee
	particles                 []*Particle
	honey                     float64
	honeyRate                 float64
	honeyRateMax              float64
	honeyMilestones           []float64
	nextMilestone             int
	fullGardenAchieved        bool
	flashText                 string
	flashTick                 int
	visitors                  []*Visitor
	achievements              []*Achievement
	showAchievements          bool
	allAchievementsCelebrated bool
	picker                    *PlantPicker
	pickerUnlocked            bool
	holdPlant                 *Plant
	holdTicks                 int
	activeTouchID             ebiten.TouchID
	lastTouchX, lastTouchY   int
}

func New(fsys fs.FS) (*Game, error) {
	bg, err := loadImage(fsys, "assets/garden.png")
	if err != nil {
		return nil, err
	}
	bgBounds := bg.Bounds()
	width, height := bgBounds.Dx(), bgBounds.Dy()

	trophyImg, err := loadImage(fsys, "assets/trophy.png")
	if err != nil {
		return nil, err
	}

	type plantDef struct {
		key         string
		baseName    string
		growRate    float64
		pollenValue float64
	}
	defs := []plantDef{
		{PlantWheat, "wheat", 1.0, 0.1},
		{PlantChaber, "cornflower", 0.9, 1.0},
		{PlantPoppy, "poppy", 1.1, 1.0},
		{PlantStJohnsWort, "yellow", 1.0, 0.2},
		{PlantNettle, "nettle", 1.2, 0.7},
		{PlantMint, "mint", 0.8, 1.0},
	}

	plantTypes := make(map[string]*PlantType, len(defs))
	for _, d := range defs {
		variants, err := loadPlantVariants(fsys, "assets", d.baseName, plantFrameWidth, plantFrameHeight, plantFrameCount)
		if err != nil {
			return nil, err
		}
		pt := &PlantType{
			name:        d.key,
			variants:    variants,
			growRate:    d.growRate,
			pollenValue: d.pollenValue,
		}
		// Pre-render picker preview from final frame of first variant.
		if len(variants) > 0 && len(variants[0]) > 0 {
			final := variants[0][len(variants[0])-1]
			bounds := final.Bounds()
			scale := pickerPreviewSize / float64(bounds.Dx())
			preview := ebiten.NewImage(int(pickerPreviewSize), int(pickerPreviewSize))
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			preview.DrawImage(final, op)
			pt.previewImg = preview
		}
		plantTypes[d.key] = pt
	}

	// Guard: catch plantTypeNames / defs drift at startup rather than silently misbehaving.
	if len(plantTypes) != len(plantTypeNames) {
		return nil, fmt.Errorf("plantTypes has %d entries but plantTypeNames has %d", len(plantTypes), len(plantTypeNames))
	}
	for _, name := range plantTypeNames {
		if _, ok := plantTypes[name]; !ok {
			return nil, fmt.Errorf("plantTypeNames contains %q but no def exists for it", name)
		}
	}

	growthTicks := float64(plantGrowthSeconds * int(ebiten.TPS()))
	if growthTicks <= 0 {
		growthTicks = float64(plantGrowthSeconds * defaultPlantTickRate)
	}

	centerGroundX := width / 2
	centerGroundY := height/2 + groundCenterOffsetY

	type plantSlot struct {
		gridX, gridY     int
		offsetX, offsetY int
	}
	edgeSlots := []plantSlot{
		{0, 0, 1, -26},
		{1, 0, 30, -14},
		{2, 0, -30, -13},
		{0, 1, -61, -2},
		{1, 1, 2, 3},
		{2, 1, 64, 0},
		{0, 2, -30, 14},
		{1, 2, 33, 15},
		{2, 2, 2, 30},
	}

	// All plants start as one randomly chosen type so Full Garden requires work.
	startType := plantTypeNames[rand.Intn(len(plantTypeNames))]
	plants := make([]*Plant, 0, len(edgeSlots))
	for _, slot := range edgeSlots {
		groundX := centerGroundX + slot.offsetX
		groundY := centerGroundY + slot.offsetY
		plantX, plantY := spritePositionForGroundCenter(groundX, groundY, plantFrameWidth, plantFrameHeight, plantGroundHeight)
		pt := plantTypes[startType]
		plant := NewPlant(PlantConfig{
			Frames:       pt.randomFrames(),
			Name:         startType,
			GridX:        slot.gridX,
			GridY:        slot.gridY,
			X:            plantX,
			Y:            plantY,
			GrowthTicks:  growthTicks,
			TypeGrowRate: pt.growRate,
			PollenValue:  pt.pollenValue,
		})
		plant.RandomizeGrowth()
		plants = append(plants, plant)
	}

	sort.Slice(plants, func(i, j int) bool {
		if plants[i].y != plants[j].y {
			return plants[i].y < plants[j].y
		}
		return plants[i].x < plants[j].x
	})

	visitors := makeVisitors()
	return &Game{
		w:               width,
		h:               height,
		bg:              bg,
		trophyImg:       trophyImg,
		growRate:        defaultGrowRate,
		plantTypes:      plantTypes,
		plants:          plants,
		honeyMilestones: []float64{50, 200, 1000},
		picker:          newPlantPicker(),
		activeTouchID:   -1,
		visitors:        visitors,
		achievements:    makeAchievements(visitors),
	}, nil
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.debugEnabled = !g.debugEnabled
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showAchievements = !g.showAchievements
	}

	for _, plant := range g.plants {
		plant.Update(g.growRate)
	}

	// Bees: update, harvest pollen from returning bees, manage population.
	grownTypes := g.grownTypeNames()
	target := g.targetBeeCount(grownTypes)
	prevHoney := g.honey
	var active []*Bee
	for _, bee := range g.bees {
		bee.Update(g.plants)
		if !bee.done {
			active = append(active, bee)
		} else {
			g.honey += bee.pollen
		}
	}
	// Exponential moving average of per-tick honey gain (α=0.0008, ~20s window).
	g.honeyRate = g.honeyRate*0.9992 + (g.honey-prevHoney)*0.0008
	if g.honeyRate > g.honeyRateMax {
		g.honeyRateMax = g.honeyRate
	}
	g.bees = active
	for i := target; i < len(g.bees); i++ {
		if !g.bees[i].goingHome {
			g.bees[i].sendHome()
		}
	}
	for len(g.bees) < target {
		g.bees = append(g.bees, newBeeAtHive())
	}

	if !g.pickerUnlocked && len(g.bees) >= 20 {
		g.pickerUnlocked = true
		g.flash("Plant picker unlocked!")
	}

	// Particles.
	var activeParticles []*Particle
	for _, p := range g.particles {
		p.Update()
		if !p.Done() {
			activeParticles = append(activeParticles, p)
		}
	}
	g.particles = activeParticles
	g.emitPollenParticles()

	// Milestones.
	if g.nextMilestone < len(g.honeyMilestones) && g.honey >= g.honeyMilestones[g.nextMilestone] {
		g.flash(fmt.Sprintf("%.0f honey!", g.honeyMilestones[g.nextMilestone]))
		g.nextMilestone++
	}
	if !g.fullGardenAchieved && len(grownTypes) >= len(plantTypeNames) {
		g.flash("Full Garden!")
		g.fullGardenAchieved = true
	}
	if g.flashTick > 0 {
		g.flashTick--
	}

	if !g.allAchievementsCelebrated {
		allDone := true
		for _, a := range g.achievements {
			if !a.condition(g) {
				allDone = false
				break
			}
		}
		if allDone {
			g.allAchievementsCelebrated = true
			g.flash("Garden Complete!")
		}
	}

	for _, v := range g.visitors {
		wasPresent := v.present
		v.present = v.condition(g.plants)
		if v.present && !wasPresent && !v.seen {
			v.seen = true
			g.flash("A " + v.name + " visits!")
		}
	}

	// Unified pointer input: first touch takes priority over mouse.
	var pJustPressed, pIsHeld, pJustReleased bool
	var px, py int

	touchPressed := inpututil.AppendJustPressedTouchIDs(nil)
	touchActive := ebiten.AppendTouchIDs(nil)
	touchReleased := inpututil.AppendJustReleasedTouchIDs(nil)

	if len(touchPressed) > 0 {
		g.activeTouchID = touchPressed[0]
		px, py = ebiten.TouchPosition(g.activeTouchID)
		g.lastTouchX, g.lastTouchY = px, py
		pJustPressed, pIsHeld = true, true
	} else if len(touchActive) > 0 {
		for _, id := range touchActive {
			if id == g.activeTouchID {
				px, py = ebiten.TouchPosition(id)
				g.lastTouchX, g.lastTouchY = px, py
				pIsHeld = true
				break
			}
		}
	} else if len(touchReleased) > 0 {
		px, py = g.lastTouchX, g.lastTouchY
		pJustReleased = true
		g.activeTouchID = -1
	} else {
		px, py = ebiten.CursorPosition()
		pJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
		pIsHeld = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
		pJustReleased = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	}

	// Trophy icon — checked first so it doesn't bleed into plant interaction.
	trophyHandled := false
	tx, ty := g.w-trophyIconW-2, trophyIconY
	if pJustPressed && px >= tx && px < tx+trophyIconW && py >= ty && py < ty+trophyIconH {
		g.showAchievements = !g.showAchievements
		g.holdPlant = nil
		trophyHandled = true
	}

	if !trophyHandled {
		if g.picker.active {
			g.picker.Update(px, py)
			if pJustReleased {
				if typeName, ok := g.picker.HoveredType(); ok {
					pt := g.plantTypes[typeName]
					g.picker.plant.ChangeType(pt.randomFrames(), pt)
				}
				g.picker.Close()
			}
		} else {
			if pJustPressed {
				g.holdPlant = nil
				for _, plant := range g.plants {
					if plant.IsFullyGrown() && plant.ContainsGround(px, py) {
						g.holdPlant = plant
						g.holdTicks = 0
						break
					}
				}
			}
			if g.holdPlant != nil {
				if pIsHeld {
					g.holdTicks++
					if g.holdTicks >= pickerLongClickTicks {
						if g.pickerUnlocked {
							g.picker.Open(g.holdPlant, g.plantTypes, g.w, g.h)
						}
						g.holdPlant = nil
						g.holdTicks = 0
					}
				} else {
					// Released before threshold — quick tap: randomise type.
					newTypeName := plantTypeNames[rand.Intn(len(plantTypeNames))]
					newType := g.plantTypes[newTypeName]
					g.holdPlant.ChangeType(newType.randomFrames(), newType)
					g.holdPlant = nil
					g.holdTicks = 0
				}
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
	for _, p := range g.particles {
		p.Draw(screen)
	}
	for _, bee := range g.bees {
		bee.Draw(screen)
	}

	g.picker.Draw(screen)
	g.drawHUD(screen)
	g.drawTrophyIcon(screen)
	g.drawVisitors(screen)

	if g.flashTick > 0 {
		g.drawFlash(screen)
	}
	if g.showAchievements {
		g.drawAchievements(screen)
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
	anchorX := frameWidth / 2
	anchorY := frameHeight - groundHeight/2
	return centerX - anchorX, centerY - anchorY
}

// grownTypeNames returns the set of plant type names that have at least one fully-grown plant.
func (g *Game) grownTypeNames() map[string]struct{} {
	types := make(map[string]struct{})
	for _, p := range g.plants {
		if p.IsFullyGrown() {
			types[p.Name()] = struct{}{}
		}
	}
	return types
}

func (g *Game) targetBeeCount(grownTypes map[string]struct{}) int {
	total := 0.0
	for _, p := range g.plants {
		if p.IsFullyGrown() {
			total += p.PollenValue()
		}
	}
	return int(total) + int(float64(len(grownTypes))*2.5)
}

func (g *Game) drawVisitors(screen *ebiten.Image) {
	var names []string
	for _, v := range g.visitors {
		if v.present {
			names = append(names, v.name)
		}
	}
	if len(names) == 0 {
		return
	}
	text := "Visitors:"
	for _, n := range names {
		text += "  " + n
	}
	ebitenutil.DebugPrintAt(screen, text, 2, 14)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, float64(g.w), 13, color.RGBA{0, 0, 0, 160})
	rate := g.honeyRate * 60
	honeyText := fmt.Sprintf("Honey: %.2f/s (max %.2f/s)", rate, g.honeyRateMax*60)
	ebitenutil.DebugPrintAt(screen, honeyText, 2, 0)
	beeText := fmt.Sprintf("Bees: %d/%d", len(g.bees), maxBeeCount)
	ebitenutil.DebugPrintAt(screen, beeText, g.w-len(beeText)*6-2, 0)
}

func (g *Game) drawFlash(screen *ebiten.Image) {
	textW := len(g.flashText) * 6
	x := (g.w - textW) / 2
	y := g.h/2 - 6
	ebitenutil.DrawRect(screen, float64(x-4), float64(y-2), float64(textW+8), 14, color.RGBA{0, 0, 0, 180})
	ebitenutil.DebugPrintAt(screen, g.flashText, x, y)
}

func (g *Game) flash(text string) {
	g.flashText = text
	g.flashTick = 180
}

func (g *Game) emitPollenParticles() {
	if len(g.particles) >= maxParticles {
		return
	}
	for _, bee := range g.bees {
		if bee.goingHome || bee.done {
			continue
		}
		for _, plant := range g.plants {
			if !plant.IsFullyGrown() {
				continue
			}
			cx, cy := plant.Center()
			dx := cx - bee.x
			dy := cy - bee.y
			if dx*dx+dy*dy < beePollenRadius*beePollenRadius {
				if rand.Float64() < particleEmitChance {
					g.particles = append(g.particles, newParticle(bee.x, bee.y))
				}
				break
			}
		}
	}
}

func (g *Game) plantsDebugText() string {
	var builder strings.Builder
	builder.WriteString("Debug:")
	for _, plant := range g.plants {
		builder.WriteString("\n")
		position := plant.gridY*3 + plant.gridX + 1
		builder.WriteString(plant.DebugLine(position))
	}
	return builder.String()
}
