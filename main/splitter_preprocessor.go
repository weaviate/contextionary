package main

import (
	"fmt"
	"github.com/semi-technologies/contextionary/preprocessing"
	"os"
)

func main() {
	if len(os.Args) != 5 {
		missing := fmt.Errorf("Missing arguments requires: [.idx, .dic, .aff, output_file]")
		panic(missing.Error())
	}

	err := preprocessing.GenerateSplittingDictFile(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		panic(err.Error())
	}
}


