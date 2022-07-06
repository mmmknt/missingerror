package main

import (
	"github.com/mmmknt/missingerror"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(missingerror.Analyzer) }
