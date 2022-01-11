package main

import (
	"fmt"
	"github.com/Yeicor/surreal/surreal2"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"log"
	"time"
)

// This example also tests support for multiple closed surfaces!
// WARNING: Text is relatively slow to render (limitation of the SDFX library)
func main() {
	f, err := sdf.LoadFont("cmr10.ttf")
	if err != nil {
		panic(fmt.Sprintf("can't read font file %s\n", err))
	}
	t := sdf.NewText("Hello,\nWorld!")
	s, err := sdf.TextSDF2(f, t, 10.0)
	if err != nil {
		log.Fatalf("can't generate text sdf2 %s\n", err)
	}

	startTime := time.Now()
	lines := surreal2.NewDefault().Run(s)
	log.Println("Generated", len(lines), "output lines in", time.Since(startTime))
	// 17 closed surfaces generate 814 lines with default settings at 11/01/2022 (sensible to small parameter and source changes)

	// Save boilerplate
	svg := render.NewSVG("render.svg", "fill:none;stroke:black;stroke-width:0.1")
	for _, line := range lines {
		//log.Println("Output line:", line)
		svg.Line(line[0], line[1])
	}
	err = svg.Save()
	if err != nil {
		panic(err)
	}
}
