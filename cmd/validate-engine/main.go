package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/oneworkspacestudio/my-slicer/internal/engine"
	"os"
	"time"
)

func main() {
	root := flag.String("engine", "", "Verified native runtime root")
	out := flag.String("out", "", "Fresh validation directory")
	flag.Parse()
	if *root == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "Provide --engine and --out")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	if e := engine.Validate(ctx, *root, *out, func(s string) { fmt.Println(s) }); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Println("PASS both real machine profiles, matching model metadata and non-empty extrusion previews")
}
