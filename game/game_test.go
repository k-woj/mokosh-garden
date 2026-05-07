package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// helpers

func grownPlant(name string, pollenValue float64) *Plant {
	return &Plant{name: name, grown: true, pollenValue: pollenValue, growthTicks: 1800, elapsedTick: 1800}
}

func ungrown(name string) *Plant {
	return &Plant{name: name, grown: false, growthTicks: 1800}
}

func gameWithPlants(plants []*Plant) *Game {
	return &Game{plants: plants}
}

// --- targetBeeCount ---

func TestTargetBeeCount_Empty(t *testing.T) {
	g := gameWithPlants(nil)
	if got := g.targetBeeCount(g.grownTypeNames()); got != 0 {
		t.Errorf("empty garden: want 0, got %d", got)
	}
}

func TestTargetBeeCount_SingleWheat(t *testing.T) {
	g := gameWithPlants([]*Plant{grownPlant(PlantWheat, 0.1)})
	// total=0.1 → int=0; uniqueTypes=1 → 1*2.5=2; total=2
	if got := g.targetBeeCount(g.grownTypeNames()); got != 2 {
		t.Errorf("single wheat: want 2, got %d", got)
	}
}

func TestTargetBeeCount_NineFlowers_SixTypes(t *testing.T) {
	plants := []*Plant{
		grownPlant(PlantWheat, 0.1),
		grownPlant(PlantStJohnsWort, 0.2),
		grownPlant(PlantNettle, 0.7),
		grownPlant(PlantChaber, 1.0),
		grownPlant(PlantPoppy, 1.0),
		grownPlant(PlantMint, 1.0),
		grownPlant(PlantChaber, 1.0),
		grownPlant(PlantPoppy, 1.0),
		grownPlant(PlantMint, 1.0),
	}
	g := gameWithPlants(plants)
	// total = 0.1+0.2+0.7+1+1+1+1+1+1 = 7.0 → int=7
	// uniqueTypes = 6 → 6*2.5=15
	// count = 22
	if got := g.targetBeeCount(g.grownTypeNames()); got != 22 {
		t.Errorf("nine flowers six types: want 22, got %d", got)
	}
}

func TestTargetBeeCount_UngrownPlantsIgnored(t *testing.T) {
	g := gameWithPlants([]*Plant{ungrown(PlantChaber), ungrown(PlantPoppy)})
	if got := g.targetBeeCount(g.grownTypeNames()); got != 0 {
		t.Errorf("all ungrown: want 0, got %d", got)
	}
}

// --- grownTypeNames ---

func TestGrownTypeNames(t *testing.T) {
	g := gameWithPlants([]*Plant{
		grownPlant(PlantMint, 1.0),
		grownPlant(PlantMint, 1.0),
		grownPlant(PlantNettle, 0.7),
		ungrown(PlantPoppy),
	})
	types := g.grownTypeNames()
	if len(types) != 2 {
		t.Errorf("want 2 unique grown types, got %d", len(types))
	}
	if _, ok := types[PlantMint]; !ok {
		t.Error("expected Mint in grown types")
	}
	if _, ok := types[PlantNettle]; !ok {
		t.Error("expected Nettle in grown types")
	}
}

// --- visitor conditions ---

func TestGoatVisitor(t *testing.T) {
	visitors := makeVisitors()
	goat := visitorByName(visitors, "Goat")

	plants3Wheat := []*Plant{
		grownPlant(PlantWheat, 0.1), grownPlant(PlantWheat, 0.1), grownPlant(PlantWheat, 0.1),
		grownPlant(PlantMint, 1.0),
	}
	if goat.condition(plants3Wheat) {
		t.Error("goat: should not visit with only 3 wheat")
	}

	plants4Wheat := append(plants3Wheat, grownPlant(PlantWheat, 0.1))
	if !goat.condition(plants4Wheat) {
		t.Error("goat: should visit with 4 wheat")
	}
}

func TestDeerVisitor(t *testing.T) {
	visitors := makeVisitors()
	deer := visitorByName(visitors, "Deer")

	if deer.condition(nil) {
		t.Error("deer: should not visit with no plants")
	}
	if deer.condition([]*Plant{ungrown(PlantMint)}) {
		t.Error("deer: should not visit with no grown plants")
	}

	mintNettle := []*Plant{grownPlant(PlantMint, 1.0), grownPlant(PlantNettle, 0.7)}
	if !deer.condition(mintNettle) {
		t.Error("deer: should visit with only mint and nettle grown")
	}

	mixed := []*Plant{grownPlant(PlantMint, 1.0), grownPlant(PlantPoppy, 1.0)}
	if deer.condition(mixed) {
		t.Error("deer: should not visit when non-mint/nettle plants are grown")
	}
}

func TestButterflyVisitor(t *testing.T) {
	visitors := makeVisitors()
	butterfly := visitorByName(visitors, "Butterfly")

	three := []*Plant{
		grownPlant(PlantStJohnsWort, 0.2), grownPlant(PlantStJohnsWort, 0.2), grownPlant(PlantStJohnsWort, 0.2),
	}
	if butterfly.condition(three) {
		t.Error("butterfly: should not visit with only 3 yellow flowers")
	}

	four := append(three, grownPlant(PlantStJohnsWort, 0.2))
	if !butterfly.condition(four) {
		t.Error("butterfly: should visit with 4 yellow flowers")
	}
}

// --- plant growth ---

func TestPlantFrameIndex_Ungrown(t *testing.T) {
	p := &Plant{frames: make([]*ebiten.Image, 16), growthTicks: 1800, elapsedTick: 0}
	if got := p.currentFrameIndex(); got != 0 {
		t.Errorf("start of growth: want frame 0, got %d", got)
	}
}

func TestPlantFrameIndex_HalfGrown(t *testing.T) {
	p := &Plant{frames: make([]*ebiten.Image, 16), growthTicks: 1800, elapsedTick: 900}
	// index = int(900 * 15 / 1800) = int(7.5) = 7
	if got := p.currentFrameIndex(); got != 7 {
		t.Errorf("half grown: want frame 7, got %d", got)
	}
}

func TestPlantFrameIndex_FullyGrown(t *testing.T) {
	p := &Plant{frames: make([]*ebiten.Image, 16), growthTicks: 1800, elapsedTick: 1800, grown: true}
	if got := p.currentFrameIndex(); got != 15 {
		t.Errorf("fully grown: want frame 15, got %d", got)
	}
}

func TestPlantUpdate_StopsAtGrowthTicks(t *testing.T) {
	p := &Plant{growthTicks: 100, elapsedTick: 98, growFactor: 1.0, typeGrowRate: 1.0}
	p.Update(1.0)
	p.Update(1.0)
	p.Update(1.0)
	if !p.grown {
		t.Error("plant should be fully grown after exceeding growthTicks")
	}
	if p.elapsedTick != p.growthTicks {
		t.Errorf("elapsedTick should be clamped to growthTicks, got %f", p.elapsedTick)
	}
}

func TestPlantUpdate_ZeroGrowRateDoesNothing(t *testing.T) {
	p := &Plant{growthTicks: 100, elapsedTick: 0, growFactor: 1.0, typeGrowRate: 1.0}
	p.Update(0)
	if p.elapsedTick != 0 {
		t.Error("zero growRate should not advance plant")
	}
}
