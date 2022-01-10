package main

import (
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"iso-planar-gen/sdf2"
	"log"
	"math"
	"time"
)

func main() {
	s, _ := sdf.Circle2D(0.5)
	sDiff, _ := sdf.Circle2D(0.25)
	s = sdf.Difference2D(s, sdf.Transform2D(sDiff, sdf.Translate2d(sdf.V2{X: 0.25, Y: 0.25})))

	startTime := time.Now()
	lines := sdf2.NewIsoPlanarGen2Simple(math.Pi/12, 1e-3, sdf.V2i{1, 1}).Run(s)
	log.Println("Generated", len(lines), "output lines in", time.Since(startTime))

	// Save boilerplate
	svg := render.NewSVG("output.svg", "fill:none;stroke:black;stroke-width:0.01")
	for _, line := range lines {
		//log.Println("Output line:", line)
		svg.Line(line[0], line[1])
	}
	err := svg.Save()
	if err != nil {
		panic(err)
	}
}
