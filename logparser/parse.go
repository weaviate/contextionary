package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type logEntry struct {
	Action string `json:"action"`
	Words  []word `json:"words"`
}

type word struct {
	Occurrence int     `json:"occurrence"`
	Weight     float64 `json:"weight"`
	Word       string  `json:"word"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var results []logEntry

	for scanner.Scan() {
		var current logEntry
		err := json.Unmarshal(scanner.Bytes(), &current)
		if err != nil {
			log.Fatal(err)
		}

		if current.Action == "debug_vector_weights" {
			results = append(results, current)
		}
	}

	marshalled, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(string(marshalled))
}
