package internal

type LightType int
type Side int

const (
	LightTypeServiceEntrance = iota
	LightTypeServiceNoManLand
	LightTypeShortWindow
	LightTypeLongWindow
	LightTypeWallStub

	SideFront = iota
	SideRight
	SideBack
	SideLeft
)

type Light struct {
	Number int
	Side   Side
	Kind   LightType
	Addr   LightAddress
}

type LightAddress struct {
	Pin   string
	Board uint8
}

type LigtsBuildingMap struct {
	Levels [][]Light
}
