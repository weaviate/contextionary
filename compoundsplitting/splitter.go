package compoundsplitting

import (
	"context"
	"fmt"
	"time"
)

// minCompoundWordLength prevents the splitting into very small (often not real) words
//  to prevent a bloated tree
const minCompoundWordLength = 4

// maxWordLength prevents a tree from growing too big when adding very long strings
const maxWordLength = 100

// maxNumberTreeNodes
const maxNumberTreeNodes = 20

const cancelSplittingAfter = 500 * time.Millisecond

type Dictionary interface {
	// Score receives a phrase of words and gives a score on how "good" this phrase is.
	//  If a compound word can be splitted into multiple phrases it will choose the one with the highest score.
	Score(phrase []string) float64
	// Contains is true if the word is in the dictionary
	Contains(word string) bool
}

// Splitter builds a tree of compound splits and selects
//  the best option based on a scoring mechanism
type Splitter struct {
	dict Dictionary
	cancelAfter       time.Duration
}

// New Splitter recognizing words given by dict and
//  selecting split phrases based on scoring
func NewSplitter(dict Dictionary) *Splitter {
	return &Splitter{
		dict:              dict,
		cancelAfter:       cancelSplittingAfter,
	}
}

type CompoundSplit struct {
	// Combinations of compound combinations in a phrase
	combinations      []*Node
}

// Split a compound word into its compounds
func (sp *Splitter) Split(word string) ([]string, error) {

	if len(word) > maxWordLength {
		return []string{}, nil
	}

	compoundSplit := CompoundSplit{}

	// spawn a new context that cancels the recursion if we are spending too much
	// time on it
	ctx, cancel := context.WithTimeout(context.Background(), sp.cancelAfter)
	defer cancel()

	err := sp.findAllWordCombinations(ctx, word, &compoundSplit)
	if err != nil {
		return nil, err
	}
	combinations := compoundSplit.getAllWordCombinations(ctx)
	maxScore := 0.0
	maxPhrase := []string{}
	for _, combination := range combinations {
		currentScore := sp.dict.Score(combination)
		if len(maxPhrase) == 0 {
			// Initialize if score is negative
			maxScore = currentScore
			maxPhrase = combination
		}
		if currentScore > maxScore {
			maxScore = currentScore
			maxPhrase = combination
		}
	}
	return maxPhrase, nil
}

func (cs *CompoundSplit) insertCompound(ctx context.Context, word string,
	startIndex int) error {
	compound := NewNode(word, startIndex)
	appended := false
	for _, combination := range cs.combinations {
		// For all possible combinations

		leaves := combination.RecursivelyFindLeavesBeforeIndex(ctx, startIndex)
		for _, leave := range leaves {
			// Append the new compound to the leaves

			appended = true
			err := leave.AddChild(compound)
			if err != nil {
				return err
			}
		}
	}
	if !appended {
		// if compound was not added to any leave add it to combinations
		cs.combinations = append(cs.combinations, compound)
	}
	return nil
}

func (sp *Splitter) findAllWordCombinations(ctx context.Context, str string, compoundSplit *CompoundSplit) error {
	compoundsUsed := 0
	for offset, _ := range str {
		// go from left to right and choose offsetted substring
		offsetted := str[offset:]

		for i := 1; i <= len(offsetted); i++ {
			// go from left to right to find a word
			word := offsetted[:i]
			if len(word) < minCompoundWordLength {
				continue
			}

			if sp.dict.Contains(word) {
				compoundsUsed += 1
				if compoundsUsed == maxNumberTreeNodes {
					// Tree is getting out of bounds stopping for performance
					return nil
				}
				err := compoundSplit.insertCompound(ctx, word, offset)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (cs *CompoundSplit) getAllWordCombinations(ctx context.Context) [][]string {
	wordCombinations := [][]string{}

	for _, combination := range cs.combinations {
		wordCombinations = append(wordCombinations,
			combination.RecursivelyBuildNames(ctx)...)
	}

	return wordCombinations
}

// Node for of the word tree
type Node struct {
	name       string
	children   []*Node
	startIndex int // inclusiv
	endIndex   int // exclusive
}

// NewNode from node name and in compoundword index
func NewNode(name string, startIndex int) *Node {
	return &Node{
		name:       name,
		children:   []*Node{},
		startIndex: startIndex,
		endIndex:   startIndex + len(name),
	}
}

// AddChild node to node
func (node *Node) AddChild(newChildNode *Node) error {
	if newChildNode.startIndex < node.endIndex {
		return fmt.Errorf("Child starts at %v but this node ends at %v can't add as child", newChildNode.startIndex, node.endIndex)
	}
	node.children = append(node.children, newChildNode)
	return nil
}

func (node *Node) findChildNodesBeforeIndex(index int) []*Node {
	childrensThatEndBeforeIndex := []*Node{}

	for _, child := range node.children {
		if child.endIndex <= index {
			childrensThatEndBeforeIndex = append(childrensThatEndBeforeIndex, child)
		}
	}

	return childrensThatEndBeforeIndex
}

// RecursivelyBuildNames of compounds
func (node *Node) RecursivelyBuildNames(ctx context.Context) [][]string {
	compoundName := [][]string{}
	if ctx.Err() != nil {
		// we've been going recursively too long, abort!
		compoundName = append(compoundName, []string{node.name})
		return compoundName
	}

	for _, child := range node.children {
		childNames := child.RecursivelyBuildNames(ctx)

		for _, childName := range childNames {
			// Add the name of this node first
			fullName := []string{node.name}
			fullName = append(fullName, childName...)
			compoundName = append(compoundName, fullName)
		}
	}
	if len(compoundName) == 0 {
		// This is a leave node
		compoundName = append(compoundName, []string{node.name})
	}

	return compoundName
}

// RecursivelyFindLeavesBeforeIndex where to add a new node
func (node *Node) RecursivelyFindLeavesBeforeIndex(ctx context.Context, index int) []*Node {
	foundLeaves := []*Node{}
	if ctx.Err() != nil {
		// we've been going recursively too long, abort!
		return foundLeaves
	}

	children := node.findChildNodesBeforeIndex(index)
	for _, child := range children {
		leaves := child.RecursivelyFindLeavesBeforeIndex(ctx, index)
		if len(leaves) == 0 {
			// There are no leaves this means the child node is already a leave
			foundLeaves = append(foundLeaves, child)
		} else {
			// Found leaves use them instead of direct child
			foundLeaves = append(foundLeaves, leaves...)
		}
	}

	if len(foundLeaves) == 0 && node.endIndex <= index {
		// This node is the leave
		foundLeaves = append(foundLeaves, node)
	}

	return foundLeaves
}

// NewEmptyTestSplitter creates a splitter,
//  that does not know any words and
//  thus is not able to split any words
func NewEmptyTestSplitter() *Splitter {
	dictMock := &DictMock{
		scores: map[string]float64{},
	}
	return &Splitter{
		dict: dictMock,
	}
}

func NewTestSplitter(wordScoreMapping map[string]float64) *Splitter {
	dict := &DictMock{
		scores: wordScoreMapping,
	}
	return &Splitter{
		dict: dict,
	}
}
