package extensions

import core "github.com/semi-technologies/contextionary/contextionary/core"

type Extension struct {
	Concept    string         `json:"concept"`
	Vector     core.Vector    `json:"vector"`
	Occurrence int            `json:"occurrence"`
	Input      ExtensionInput `json:"input"`
}

type ExtensionInput struct {
	Definition string  `json:"definition"`
	Weight     float32 `json:"weight"`
}
