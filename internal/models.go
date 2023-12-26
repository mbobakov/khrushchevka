package internal

type LightType int

const (
	LightTypeMaintenanceEntrance = iota
	LightTypeMaintenanceNoManLand
	LightTypeShortWindow1
	LightTypeLongWindow
	LightTypeShortWindow2
	LightTypeShortSideWindow
)

type Light struct {
	Kind LightType
	Addr LightAddress
}

type LightAddress struct {
	Pin   string
	Board uint8
}

type Appartment struct {
	Number  int
	Windows []Light
}

type LigtsBuildingMap struct {
	Appartments map[int][]Appartment //floor to appartments
	Maintenance map[int][]Light      //floor to maintance lights
}
