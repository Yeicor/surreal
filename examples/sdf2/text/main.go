package main

import (
	"fmt"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"iso-planar-gen/sdf2"
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
	lines := sdf2.NewIsoPlanarGen2Default().Run(s)
	log.Println("Generated", len(lines), "output lines in", time.Since(startTime))

	// Save boilerplate
	svg := render.NewSVG("output.svg", "fill:none;stroke:black;stroke-width:0.1")
	for _, line := range lines {
		//log.Println("Output line:", line)
		svg.Line(line[0], line[1])
	}
	err = svg.Save()
	if err != nil {
		panic(err)
	}
}
