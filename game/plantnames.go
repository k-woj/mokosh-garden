package game

// Plant type name constants — use these instead of raw string literals
// so typos become compile errors rather than silent runtime mismatches.
const (
	PlantWheat       = "Wheat"
	PlantChaber      = "Chaber"
	PlantPoppy       = "Poppy"
	PlantStJohnsWort = "StJohnsWort"
	PlantNettle      = "Nettle"
	PlantMint        = "Mint"
)

// plantTypeNames is the single authoritative ordered list of plant types.
// Defined here alongside the constants so they stay in sync.
var plantTypeNames = []string{
	PlantWheat, PlantChaber, PlantPoppy, PlantStJohnsWort, PlantNettle, PlantMint,
}
