package main

import (
	"github.com/Yeicor/surreal/surreal2"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"log"
	"math"
	"time"
)

func main() {
	s := sdf.Box2D(sdf.V2{X: 1, Y: 1}, 0)
	// Apply some transforms to make the problem harder
	s = sdf.Transform2D(s, sdf.Translate2d(sdf.V2{X: 4, Y: 1}).Mul(sdf.Rotate2d(math.Pi/8)))
	s = sdf.ScaleUniform2D(s, 2)

	startTime := time.Now()
	lines := surreal2.NewSimple(math.Pi/4, 1e-3, sdf.V2i{1, 1}).Run(s)
	log.Println("Generated", len(lines), "output lines in", time.Since(startTime))
	if len(lines) != 4 {
		panic("Squares (low enough step and minAngle) are expected to render using only 4 lines")
	}

	// Save boilerplate
	svg := render.NewSVG("render.svg", "fill:none;stroke:black;stroke-width:0.1")
	for _, line := range lines {
		log.Println("Output line:", line)
		svg.Line(line[0], line[1])
	}
	err := svg.Save()
	if err != nil {
		panic(err)
	}
}
