package extensions

type Extension struct {
	Concept    string         `json:"concept"`
	Vector     []float32      `json:"vector"`
	Occurrence int            `json:"occurrence"`
	Input      ExtensionInput `json:"input"`
}

type ExtensionInput struct {
	Definition string  `json:"definition"`
	Weight     float32 `json:"weight"`
}
