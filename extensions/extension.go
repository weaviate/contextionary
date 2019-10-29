package extensions

import core "github.com/semi-technologies/contextionary/contextionary/core"

type Extension struct {
	Concept    string
	Vector     core.Vector
	Occurrence int
	Input      ExtensionInput
}

type ExtensionInput struct {
	Definition string
	Weight     float32
}
