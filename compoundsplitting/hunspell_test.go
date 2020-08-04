package compoundsplitting

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImplementation(t *testing.T) {

	hsp := Hunspell("../test/compoundsplitting/nl_NL.aff", "../test/compoundsplitting/nl_NL.dic")

	assert.True(t, hsp.Spell("Amsterdam"))
	assert.True(t, hsp.Spell("appellante"))
	assert.True(t, hsp.Spell("appellantes"))

}