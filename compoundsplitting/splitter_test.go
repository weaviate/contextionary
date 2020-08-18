package compoundsplitting

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSplitTreeSplitter(t *testing.T) {
	dictMock := &DictMock{
		scores: map[string]float64{
			"drie":     2.0,
			"hoek":     2.0,
			"brood":    4.0,
			"driehoek": 5.0,
			"broodje":  5.0,
		},
	}

	ts := Splitter{
		dict: dictMock,
	}

	// drie hoek brood
	//           broodje
	// driehoek brood
	//          broodje

	ts.findAllWordCombinations("driehoeksbroodje")

	combinations := ts.getAllWordCombinations()
	assert.Equal(t, 4, len(combinations))
	for _, combination := range combinations {
		fmt.Printf("%v\n", combination)
	}

	splited, err := ts.Split("driehoeksbroodje")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(splited))
	assert.Equal(t, "driehoek", splited[0])
	assert.Equal(t, "broodje", splited[1])

	// Test no result
	splited, err = ts.Split("raupenprozessionsspinner")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(splited), "Expected no result since no substring is in the dict")
}

func TestNegativeScore(t *testing.T) {
	dictMock := &DictMock{
		scores: map[string]float64{
			"drie":     -10.0,
			"hoek":     -10.0,
			"brood":    -8.0,
			"driehoek": -2.0,
			"broodje":  -2.0,
		},
	}

	ts := NewSplitter(dictMock)

	splited, err := ts.Split("driehoeksbroodje")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(splited))
	assert.Equal(t, "driehoek", splited[0])
	assert.Equal(t, "broodje", splited[1])
}

func TestInsertCompound(t *testing.T) {

	t.Run("Add a new word", func(t *testing.T) {
		ts := Splitter{}
		ts.insertCompound("test", 0)

		assert.Equal(t, 1, len(ts.combinations))
		assert.Equal(t, "test", ts.combinations[0].name)
	})

	t.Run("Add a two words", func(t *testing.T) {
		ts := Splitter{}
		ts.insertCompound("test", 0)
		ts.insertCompound("testje", 0)

		assert.Equal(t, 2, len(ts.combinations))
		assert.Equal(t, "test", ts.combinations[0].name)
		assert.Equal(t, "testje", ts.combinations[1].name)
	})

	t.Run("Add a two words different index", func(t *testing.T) {
		ts := Splitter{}

		// phrase: testje
		ts.insertCompound("test", 0)
		ts.insertCompound("stje", 2)

		assert.Equal(t, 2, len(ts.combinations))
		assert.Equal(t, "test", ts.combinations[0].name)
		assert.Equal(t, "stje", ts.combinations[1].name)
	})

	t.Run("Add a two words different index", func(t *testing.T) {
		ts := Splitter{}

		// phrase: testjenuttig
		//         123456789111
		//                  012
		ts.insertCompound("test", 0)
		ts.insertCompound("nuttig", 8)

		assert.Equal(t, 1, len(ts.combinations))
		phrase := ts.combinations[0]
		assert.Equal(t, "test", phrase.name)
		assert.Equal(t, "nuttig", phrase.children[0].name)

	})

	t.Run("Add a two combinations", func(t *testing.T) {
		ts := Splitter{}

		// phrase: testjenuttig
		//         123456789111
		//                  012
		ts.insertCompound("test", 0)
		ts.insertCompound("est", 1)
		ts.insertCompound("nuttig", 8)

		assert.Equal(t, 2, len(ts.combinations))
		phrase := ts.combinations[0]
		assert.Equal(t, "test", phrase.name)
		assert.Equal(t, "nuttig", phrase.children[0].name)

		phrase = ts.combinations[1]
		assert.Equal(t, "est", phrase.name)
		assert.Equal(t, "nuttig", phrase.children[0].name)
	})

	t.Run("Add driehoeksbroodje", func(t *testing.T) {
		ts := Splitter{}

		// phrase: driehoeksbroodje
		//         1234567891111111
		//                  0123456
		ts.insertCompound("drie", 0)
		ts.insertCompound("driehoek", 0)
		ts.insertCompound("hoek", 5)
		ts.insertCompound("brood", 10)
		ts.insertCompound("broodje", 10)

		// drie hoek brood
		//           broodje

		// driehoek brood
		//          broodje

		assert.Equal(t, 2, len(ts.combinations))
	})

}

func TestNode(t *testing.T) {

	t.Run("New Node", func(t *testing.T) {
		node := NewNode("test", 2)
		assert.Equal(t, 6, node.endIndex)
	})

	t.Run("Add child", func(t *testing.T) {
		node1 := NewNode("test", 2)
		node2 := NewNode("case", 6)
		node3 := NewNode("ase", 7)
		err := node1.AddChild(node2)
		assert.Nil(t, err)
		err = node1.AddChild(node3)
		assert.Nil(t, err)

		assert.Equal(t, 2, len(node1.children))
	})

	t.Run("Add wrong index", func(t *testing.T) {
		node1 := NewNode("test", 2)
		node2 := NewNode("esting", 3)
		err := node1.AddChild(node2)
		assert.NotNil(t, err)
	})

	t.Run("find children before index", func(t *testing.T) {
		// testcasees
		// 0123456789
		test := NewNode("test", 0)
		caseN := NewNode("case", 4)
		as := NewNode("as", 5)
		see := NewNode("see", 6)
		es := NewNode("es", 8)

		// test case es
		// test  as  es
		// test   see

		test.AddChild(caseN)
		test.AddChild(as)
		test.AddChild(see)
		caseN.AddChild(es)
		as.AddChild(es)

		// no child nodes that end before index 6
		assert.Equal(t, 0, len(test.findChildNodesBeforeIndex(6)))
		// as ends at 7
		assert.Equal(t, 1, len(test.findChildNodesBeforeIndex(7)))
		// case ends at 8
		assert.Equal(t, 2, len(test.findChildNodesBeforeIndex(8)))
		// see ends at 9
		assert.Equal(t, 3, len(test.findChildNodesBeforeIndex(9)))
	})

	t.Run("find leaves before index", func(t *testing.T) {
		// testcasees
		// 0123456789
		test := NewNode("test", 0)
		caseN := NewNode("case", 4)
		as := NewNode("as", 5)
		see := NewNode("see", 6)
		es := NewNode("es", 8)

		// test case es
		// test  as  es
		// test   see

		test.AddChild(caseN)
		test.AddChild(as)
		test.AddChild(see)
		caseN.AddChild(es)
		as.AddChild(es)

		assert.Equal(t, 0, len(test.RecursivelyFindLeavesBeforeIndex(0)))
		assert.Equal(t, 0, len(test.RecursivelyFindLeavesBeforeIndex(3)))
		assert.Equal(t, 1, len(test.RecursivelyFindLeavesBeforeIndex(4)))
		node := test.RecursivelyFindLeavesBeforeIndex(4)[0]
		assert.Equal(t, "test", node.name)

		assert.Equal(t, 1, len(test.RecursivelyFindLeavesBeforeIndex(7)))
		node = test.RecursivelyFindLeavesBeforeIndex(7)[0]
		assert.Equal(t, "as", node.name)

		assert.Equal(t, 2, len(test.RecursivelyFindLeavesBeforeIndex(8)))
	})

}

func TestSplitVeryLongWords(t *testing.T) {
	dictMock := &DictMock{
		scores: map[string]float64{
			"aaaa": 1.0,
			"bbbb": 1.0,
		},
	}

	ts := Splitter{
		dict: dictMock,
	}

	t1 := time.Now()

	split, err := ts.Split("aaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaa")

	t2 := time.Now()
	diff := t2.Sub(t1)

	assert.Nil(t, err)
	assert.Less(t, 0, len(split))

	if diff > time.Millisecond * 200 {
		fmt.Errorf("Splitter took too long")
		t.Fail()
	}
}

func TestSplitTooLongWords(t *testing.T) {
	dictMock := &DictMock{
		scores: map[string]float64{
			"aaaa": 1.0,
			"bbbb": 1.0,
		},
	}

	ts := Splitter{
		dict: dictMock,
	}

	split, err := ts.Split("aaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbb")

	assert.Nil(t, err)
	assert.Equal(t, 0, len(split))
}