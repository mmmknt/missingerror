package main

import (
	"missingerror"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(missingerror.Analyzer) }
