package surreal2

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"log"
	"math"
	"math/rand"
)

// Algorithm builds a surface by moving along it's border and making lines as long as possible.
//
//  PARAMETERS:
//  - minAngle limits the angle between tangents to generate a new vertex (maximum value is 90º, expressed in radians).
//  - step is the distance between surface checks and is useful for avoiding skipping small features that return
//    to the original tangent. Keep it as close to 0, as possible (as long as performance is good enough).
//  - normalEps should generally be as close to 0 as possible (considering numerical inaccuracies).
//  - scanSurfaceCells is required when there are multiple surfaces, as the algorithm needs to find each one of them to
//    properly generate them.
//  - scanSurfaceDistSq is the minimum distance between surfaces to consider them as different (keep as high as your
//    surfaces allow in order to avoid double surfaces)
//  - surf* see fallToSurface.
//  - rng is only used for fallToSurface (to solve 0 gradient by pushing in a random direction).
type Algorithm struct {
	minAngle, step, normalEps float64
	scanSurfaceCells          sdf.V2i
	scanSurfaceDistSq         float64
	surfHitEps, surfStepSize  float64
	surfMaxSteps              int
	rng                       *rand.Rand
}

// NewDefault values may change at any time. See NewSimple.
func NewDefault() *Algorithm {
	return NewSimple(
		math.Pi/30, /* <=X segments for an 180 degree arc */
		1e-3,       /* will not lose features (that go back to having the same normal, e.g., spikes) bigger than this (relative to bounding box) */
		sdf.V2i{10, 10} /* will scan for multiple surfaces in a XxY uniform grid (it will cancel duplicates) */)
}

// NewSimple values may change at any time. See New.
func NewSimple(minAngle, step float64, scanSurfaceCells sdf.V2i) *Algorithm {
	return New(minAngle, step, 1e-12, scanSurfaceCells, 0.1, 1e-12,
		1, 100, rand.NewSource(0))
}

// New see Algorithm.
func New(minAngle float64, step float64, normalEps float64, scanSurfaceCells sdf.V2i, scanSurfaceDistSq, surfHitEps float64, surfStepSize float64, surfMaxSteps int, randSource rand.Source) *Algorithm {
	return &Algorithm{
		minAngle:          minAngle,
		step:              step,
		normalEps:         normalEps,
		scanSurfaceCells:  scanSurfaceCells,
		scanSurfaceDistSq: scanSurfaceDistSq,
		surfHitEps:        surfHitEps,
		surfStepSize:      surfStepSize,
		surfMaxSteps:      surfMaxSteps,
		rng:               rand.New(randSource),
	}
}

func (a *Algorithm) Run(s sdf.SDF2) [][2]sdf.V2 {
	// TODO: Find a way to parallelize the algorithm (hard because each line depends on the previous and
	//  each surface depends on each other to know if they are new)
	//printSdf(s, sdf.V2i{50, 20})
	// Scale some variables to adapt sizes
	bb := s.BoundingBox()
	bbSize := bb.Size()
	bbSizeLength := bbSize.MaxComponent()
	a.step *= bbSizeLength
	// Outputs
	var allSurfaces [][2]sdf.V2
	allLinesRtree := rtreego.NewTree(2, 3, 5)
	// Scan a uniform grid for surfaces
	cellSize := bbSize.Div(a.scanSurfaceCells.ToV2())
	bbMinCenter := bb.Min.Add(cellSize.DivScalar(2))
	var cellIndex sdf.V2i
	for cellIndex[0] = 0; cellIndex[0] < a.scanSurfaceCells[0]; cellIndex[0]++ {
		for cellIndex[1] = 0; cellIndex[1] < a.scanSurfaceCells[1]; cellIndex[1]++ {
			cellCenter := bbMinCenter.Add(cellSize.Mul(cellIndex.ToV2()))
			firstPointOnSurface := fallToSurface(s, cellCenter, a.surfHitEps, a.normalEps, a.surfStepSize, a.surfMaxSteps, a.rng)
			if firstPointOnSurface == nil {
				log.Println("[SURREAL2] WARNING: Surface not found")
				continue
			}
			firstPoint := *firstPointOnSurface
			// Move this point to an "edge" (optional, may reduce the number of lines by 1 and avoid intersections)
			firstPoint = a.walkAlongSurface(s, &toProcess{firstPoint, true}, nil, nil)
			// If the found point is not on any previously generated surface...
			_, closestVertDistSq, _ := findNearest(allLinesRtree, firstPoint, 2 /* TODO: more? */)
			if closestVertDistSq == math.MaxFloat64 || closestVertDistSq > a.scanSurfaceDistSq {
				// Build the new surface
				//log.Println("[SURREAL2] Generating surface at", cellIndex, ">", firstPoint, "with closest", closestVertDistSq)
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
