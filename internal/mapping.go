package internal

var (
	BuildingMap = LigtsBuildingMap{
		Maintenance: map[int][]Light{
			1: {Light{Kind: LightTypeMaintenanceEntrance, Addr: LightAddress{Board: 0x25, Pin: "A1"}}},
			2: {Light{Kind: LightTypeMaintenanceNoManLand, Addr: LightAddress{Board: 0x25, Pin: "A2"}}},
			3: {Light{Kind: LightTypeMaintenanceNoManLand, Addr: LightAddress{Board: 0x25, Pin: "A3"}}},
			4: {Light{Kind: LightTypeMaintenanceNoManLand, Addr: LightAddress{Board: 0x25, Pin: "A4"}}},
			5: {Light{Kind: LightTypeMaintenanceNoManLand, Addr: LightAddress{Board: 0x25, Pin: "A5"}}},
		},
		Appartments: map[int][]Appartment{
			1: {
				{Number: 1, Windows: []Light{
					{Kind: LightTypeShortWindow1, Addr: LightAddress{Board: 0x24, Pin: "A1"}},
					{Kind: LightTypeLongWindow, Addr: LightAddress{Board: 0x24, Pin: "A2"}},
					{Kind: LightTypeShortWindow2, Addr: LightAddress{Board: 0x24, Pin: "A3"}},
					{Kind: LightTypeShortSideWindow, Addr: LightAddress{Board: 0x24, Pin: "A4"}},
				},
				},
				{Number: 2},
				{Number: 3},
				{Number: 4},
			},
		},
	}
)
