package main

import (
	"github.com/Yeicor/surreal/surreal3"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"log"
	"math"
	"strconv"
	"time"
)

func main() {
	s, _ := sdf.Box3D(sdf.V3{X: 1, Y: 1, Z: 1}, 0)
	// TODO: Apply some transforms to make the problem harder
	s = sdf.Transform3D(s, sdf.Translate3d(sdf.V3{X: 0.1, Y: 0.144, Z: -0.25234}))
	//s = sdf.Transform3D(s, sdf.Rotate3d(sdf.V3{X: 1, Y: 1, Z: 1}.Normalize(), math.Pi/32))
	s = sdf.ScaleUniform3D(s, 2)

	startTime := time.Now()
	//for step := 0.01; step > 1e-6; step *= 0.1 {
	step := 0.01
	alg := surreal3.NewSimple(math.Pi/4, step, sdf.V3i{1, 1, 1})
	triangles := alg.Run(s)
	log.Println("Generated", len(triangles), "output triangles in", time.Since(startTime))
	for _, triangle := range triangles {
		log.Println("Output triangle:", triangle)
	}

	// Save boilerplate
	render.ToSTL(s, -1, "render.stl", alg)

	if len(triangles) != 12 {
		panic("Cubes (low enough step and minAngle) are expected to render using only 12 triangles, but got " + strconv.Itoa(len(triangles)))
	}
	//}
}
