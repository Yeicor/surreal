package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"math"
	"math/rand"
)

// fallToSurface moves any point to a nearby point in the surface (no direction limit of raycasts)
//
// PARAMETERS:
// - hitEps how close to the surface to be in order to consider a hit
// - normalEps should generally be as close to 0 as possible (considering numerical inaccuracies)
// - stepSize indicates how much to move on each step (should be in (0, 1])
// - maxSteps sets a limit to the number of steps (will fail if this number is reached)
func fallToSurface(s sdf.SDF3, from sdf.V3, hitEps, normalEps, stepSize float64, maxSteps int, rng *rand.Rand) *sdf.V3 {
	for step := 0; step < maxSteps; step++ {
		val := s.Evaluate(from)
		if math.Abs(val) < hitEps {
			return &from
		}
		normal := sdf.Normal3(s, from, normalEps)
		normal = normalHandleNaNs(normal, rng, normalEps)
		//log.Println("[fallToSurface] from", from, "normal", normal, "val", val, -val*stepSize)
		from = from.Add(normal.MulScalar(-val * stepSize))
	}
	return nil // Not found
}

func normalHandleNaNs(normal sdf.V3, rng *rand.Rand, normalEps float64) sdf.V3 {
	if math.IsNaN(normal.X) && math.IsNaN(normal.Y) && math.IsNaN(normal.Z) { // Decide randomly when the normal can't move us
		normal.X = rng.Float64()*normalEps - normalEps/2
		normal.Y = rng.Float64()*normalEps - normalEps/2
		normal.Z = rng.Float64()*normalEps - normalEps/2
	} else if math.IsNaN(normal.X) && math.IsNaN(normal.Y) {
		normal.X = 0
		normal.Y = 0
		normal.Z = sign(normal.Z)
	} else if math.IsNaN(normal.X) && math.IsNaN(normal.Z) {
		normal.X = 0
		normal.Y = sign(normal.Y)
		normal.Z = 0
	} else if math.IsNaN(normal.Y) && math.IsNaN(normal.Z) {
		normal.X = sign(normal.X)
		normal.Y = 0
		normal.Z = 0
	} else if math.IsNaN(normal.X) {
		normal.X = 0
		normal.Y = sign(normal.Y)
		normal.Z = sign(normal.Z)
		normal = normal.Normalize()
	} else if math.IsNaN(normal.Y) {
		normal.X = sign(normal.X)
		normal.Y = 0
		normal.Z = sign(normal.Z)
		normal = normal.Normalize()
	} else if math.IsNaN(normal.Z) {
		normal.X = sign(normal.X)
		normal.Y = sign(normal.Y)
		normal.Z = 0
		normal = normal.Normalize()
	}
	return normal
}
