package sdf2

import (
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"log"
	"math"
)

// IsoPlanarGen2 builds a surface by moving along it's border and making lines as long as possible.
//
// PARAMETERS:
// - minAngle limits the angle between tangents to generate a new vertex.
// - step is the distance between surface checks and is useful for avoiding skipping small features that return
//   to the original tangent. Keep it as close to 0, as possible (as long as performance is good enough).
// - normalEps should generally be as close to 0 as possible (considering numerical inaccuracies).
// - scanSurfaceCells is required when there are multiple surfaces, as the algorithm needs to find each one of them to
//   properly generate them.
// - surf* see fallToSurface.
type IsoPlanarGen2 struct {
	minAngle, step, normalEps float64
	scanSurfaceCells          sdf.V2i
	surfHitEps, surfStepSize  float64
	surfMaxSteps              int
}

func NewIsoPlanarGen2Default() *IsoPlanarGen2 {
	return NewIsoPlanarGen2Simple(
		math.Pi/30, /* <=X segments for an 180 degree arc */
		1e-3,       /* will not lose features (that go back to having the same normal, e.g., spikes) bigger than this (relative to bounding box) */
		sdf.V2i{25, 25} /* will scan for multiple surfaces in a XxY uniform grid (relatively fast, as it will cancel duplicates) */)
}

func NewIsoPlanarGen2Simple(minAngle, step float64, scanSurfaceCells sdf.V2i) *IsoPlanarGen2 {
	return NewIsoPlanarGen2(minAngle, step, 1e-10, scanSurfaceCells, 1e-10, 1, 100)
}

func NewIsoPlanarGen2(minAngle float64, step float64, normalEps float64, scanSurfaceCells sdf.V2i, surfHitEps float64, surfStepSize float64, surfMaxSteps int) *IsoPlanarGen2 {
	return &IsoPlanarGen2{minAngle: minAngle, step: step, normalEps: normalEps, scanSurfaceCells: scanSurfaceCells, surfHitEps: surfHitEps, surfStepSize: surfStepSize, surfMaxSteps: surfMaxSteps}
}

func (a *IsoPlanarGen2) Run(s sdf.SDF2) []*render.Line {
	// TODO: Find a way to parallelize the algorithm (hard because each line depends on the previous and
	//  each surface depends on each other to know if they are new)
	//printSdf(s, sdf.V2i{50, 20})
	// Scale some variables to adapt sizes
	bb := s.BoundingBox()
	bbSize := bb.Size()
	bbSizeLength := bbSize.MaxComponent()
	a.step *= bbSizeLength
	// Outputs
	var allSurfaces []*render.Line
	allLinesRtree := rtreego.NewTree(2, 3, 5)
	// Scan a uniform grid for surfaces
	cellSize := bbSize.Div(a.scanSurfaceCells.ToV2())
	bbMinCenter := bb.Min.Add(cellSize.DivScalar(2))
	var cellIndex sdf.V2i
	for cellIndex[0] = 0; cellIndex[0] < a.scanSurfaceCells[0]; cellIndex[0]++ {
		for cellIndex[1] = 0; cellIndex[1] < a.scanSurfaceCells[1]; cellIndex[1]++ {
			cellCenter := bbMinCenter.Add(cellSize.Mul(cellIndex.ToV2()))
			firstPointOnSurface := fallToSurface(s, cellCenter, a.surfHitEps, a.normalEps, a.surfStepSize, a.surfMaxSteps)
			if firstPointOnSurface == nil {
				log.Println("[IsoPlanarGen2] WARNING: Surface not found")
				continue
			}
			firstPoint := *firstPointOnSurface
			// Move this point to an "edge" (optional, may reduce the number of lines by 1 and avoid intersections)
			newLine := a.runIter(s, &toProcess{firstPoint, true}, nil, nil)
			if newLine == nil {
				log.Println("[IsoPlanarGen2] WARNING: Surface not found 2")
				continue
			}
			firstPoint = newLine[1]
			// If the found point is not on any previously generated surface...
			_, closestVertDistSq, nearestLine := findNearest(allLinesRtree, firstPoint, 1)
			if closestVertDistSq == math.MaxFloat64 ||
				!rectContainsPoint(nearestLine.Bounds(), firstPoint, a.step) && closestVertDistSq > a.step {
				// Build the new surface
				//log.Println("[IsoPlanarGen2] Generating surface at", cellIndex, ">", firstPoint, "with closest", closestVertDistSq)
				surface, subRtree := a.buildSingleSurface(s, firstPoint)
				// Combine results
				allSurfaces = append(allSurfaces, surface...)
				// Combine lines rtree
				allRect, _ := rtreego.NewRect(rtreego.Point{-math.MaxFloat64 / 2, -math.MaxFloat64 / 2},
					[]float64{math.MaxFloat64, math.MaxFloat64})
				for _, spatial := range subRtree.SearchIntersect(allRect) {
					allLinesRtree.Insert(spatial)
				}
			} // Otherwise, skip this as it was already generated
		}
	}
	return allSurfaces
}
