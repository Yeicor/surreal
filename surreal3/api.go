package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"log"
	"math"
	"math/rand"
)

// Algorithm see surreal2.Algorithm
type Algorithm struct {
	minAngle, step, normalEps float64
	scanSurfaceCells          sdf.V3i
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
		sdf.V3i{10, 10, 10} /* will scan for multiple surfaces in a XxY uniform grid (it will cancel duplicates) */)
}

// NewSimple values may change at any time. See New.
func NewSimple(minAngle, step float64, scanSurfaceCells sdf.V3i) *Algorithm {
	return New(minAngle, step, 1e-12, scanSurfaceCells, 0.1, 1e-12,
		1, 100, rand.NewSource(0))
}

// New see Algorithm.
func New(minAngle float64, step float64, normalEps float64, scanSurfaceCells sdf.V3i, scanSurfaceDistSq, surfHitEps float64, surfStepSize float64, surfMaxSteps int, randSource rand.Source) *Algorithm {
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

func (a *Algorithm) Run(s sdf.SDF3) []*Triangle {
	//printSdf(s, sdf.V2i{50, 20})
	// Scale some variables to adapt sizes
	bb := s.BoundingBox()
	bbSize := bb.Size()
	bbSizeLength := bbSize.MaxComponent()
	a.step *= bbSizeLength
	// Outputs
	var allSurfaces []*Triangle
	allTrianglesRtree := rtreego.NewTree(3, 3, 5)
	// Scan a uniform grid for surfaces
	cellSize := bbSize.Div(a.scanSurfaceCells.ToV3())
	bbMinCenter := bb.Min.Add(cellSize.DivScalar(2))
	var cellIndex sdf.V3i
	for cellIndex[0] = 0; cellIndex[0] < a.scanSurfaceCells[0]; cellIndex[0]++ {
		for cellIndex[1] = 0; cellIndex[1] < a.scanSurfaceCells[1]; cellIndex[1]++ {
			for cellIndex[2] = 0; cellIndex[2] < a.scanSurfaceCells[2]; cellIndex[2]++ {
				cellCenter := bbMinCenter.Add(cellSize.Mul(cellIndex.ToV3()))
				firstPointOnSurface := fallToSurface(s, cellCenter, a.surfHitEps, a.normalEps, a.surfStepSize, a.surfMaxSteps, a.rng)
				if firstPointOnSurface == nil {
					log.Println("[SURREAL2] WARNING: Surface not found")
					continue
				}
				firstPoint := *firstPointOnSurface
				// Move this point to an "edge" (will make matches easier ---avoid duplicate surfaces--- and reduce the number of lines by 1 and avoid intersections)
				noPoint := sdf.V3{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64}
				firstPoint = a.walkAlongSurface(s, &toProcess{[2]sdf.V3{firstPoint, noPoint}}, nil, nil)
				secondPoint := a.walkAlongSurface(s, &toProcess{[2]sdf.V3{firstPoint, noPoint.Neg()}}, nil, nil)
				// If the found point is not on any previously generated surface...
				_, closestVertDistSq, _ := findNearest(allTrianglesRtree, firstPoint, 2 /* TODO: more? */)
				_, closestVertDistSq2, _ := findNearest(allTrianglesRtree, secondPoint, 2 /* TODO: more? */)
				if closestVertDistSq == math.MaxFloat64 || closestVertDistSq > a.scanSurfaceDistSq && closestVertDistSq2 > a.scanSurfaceDistSq {
					// Build the new surface
					//log.Println("[SURREAL2] Generating surface at", cellIndex, ">", firstPoint, "with closest", closestVertDistSq)
					surface, subRtree := a.buildSingleSurface(s, [2]sdf.V3{firstPoint, secondPoint})
					// Combine results
					allSurfaces = append(allSurfaces, surface...)
					// Combine lines rtree
					allRect, _ := rtreego.NewRect(rtreego.Point{-math.MaxFloat64 / 2, -math.MaxFloat64 / 2, -math.MaxFloat64 / 2},
						[]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64})
					for _, spatial := range subRtree.SearchIntersect(allRect) {
						allTrianglesRtree.Insert(spatial)
					}
				} // Otherwise, skip this as it was already generated
			}
		}
	}
	return allSurfaces
}
