package main

import (
	"github.com/Yeicor/surreal/surreal3"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"log"
	"math"
	"time"
)

func main() {
	s, _ := sdf.Box3D(sdf.V3{X: 1, Y: 1, Z: 1}, 0)
	// TODO: Apply some transforms to make the problem harder
	//s = sdf.Transform3D(s, sdf.Translate3d(sdf.V3{X: 4, Y: 1, Z: -4}).Mul(sdf.Rotate3d(sdf.V3{X: 1, Y: 1, Z: 1}, math.Pi/8)))
	//s = sdf.ScaleUniform3D(s, 2)

	startTime := time.Now()
	alg := surreal3.NewSimple(math.Pi/4, 1e-1, sdf.V3i{1, 1, 1})
	triangles := alg.Run(s)
	log.Println("Generated", len(triangles), "output triangles in", time.Since(startTime))
	for _, triangle := range triangles {
		log.Println("Output triangle:", triangle)
	}

	// Save boilerplate
	render.ToSTL(s, -1, "render.stl", alg)
}
