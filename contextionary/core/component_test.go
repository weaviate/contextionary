// +build sentence

package contextionary

import (
	"fmt"
	"testing"
)

func TestDevelopmentEnvironmentForContextionary(t *testing.T) {

	// Make sure you have run ./tools/dev/gen_simple_contextionary.sh
	// from the project root or downloaded a full contextionary prior
	// to running those tests.

	c11y, err := LoadVectorFromDisk("../../tools/dev/example.knn", "../../tools/dev/example.idx")
	if err != nil {
		t.Fatalf("could not generate c11y: %s", err)
	}

	fmt.Printf("here's the c11y, do whatever you want with it: %#v", c11y)

	t.Errorf("... add whatever you like!")
}
