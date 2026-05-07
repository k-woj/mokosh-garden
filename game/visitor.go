package game

type Visitor struct {
	name      string
	present   bool
	seen      bool
	condition func([]*Plant) bool
}

func visitorByName(visitors []*Visitor, name string) *Visitor {
	for _, v := range visitors {
		if v.name == name {
			return v
		}
	}
	panic("visitor not found: " + name)
}

func makeVisitors() []*Visitor {
	return []*Visitor{
		{
			name: "Goat",
			condition: func(plants []*Plant) bool {
				n := 0
				for _, p := range plants {
					if p.IsFullyGrown() && p.Name() == PlantWheat {
						n++
					}
				}
				return n >= 4
			},
		},
		{
			name: "Deer",
			condition: func(plants []*Plant) bool {
				grown := 0
				for _, p := range plants {
					if !p.IsFullyGrown() {
						continue
					}
					grown++
					if p.Name() != PlantMint && p.Name() != PlantNettle {
						return false
					}
				}
				return grown > 0
			},
		},
		{
			name: "Butterfly",
			condition: func(plants []*Plant) bool {
				n := 0
				for _, p := range plants {
					if p.IsFullyGrown() && p.Name() == PlantStJohnsWort {
						n++
					}
				}
				return n >= 4
			},
		},
	}
}
