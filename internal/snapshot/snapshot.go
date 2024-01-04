package snapshot

// LightDTO is a DTO for light record in the snapshots
type LightDTO struct {
	Board uint8  `json:"board"`
	Pin   string `json:"pin"`
	IsOn  bool   `json:"is_on"`
}
