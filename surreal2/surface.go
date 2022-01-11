package surreal2

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
func fallToSurface(s sdf.SDF2, from sdf.V2, hitEps, normalEps, stepSize float64, maxSteps int, rng *rand.Rand) *sdf.V2 {
	for step := 0; step < maxSteps; step++ {
		val := s.Evaluate(from)
		if math.Abs(val) < hitEps {
			//log.Println("[fallToSurface] steps", step)
			return &from
		}
		normal := sdf.Normal2(s, from, normalEps)
		if math.IsNaN(normal.X) && math.IsNaN(normal.Y) { // Decide randomly when the normal can't move us
			normal.X = rng.Float64()*0.1 - 0.05
			normal.Y = rng.Float64()*0.1 - 0.05
		} else if math.IsNaN(normal.X) {
			normal.X = 0
			normal.Y = sign(normal.Y)
		} else if math.IsNaN(normal.Y) { // Decide randomly when the normal can't move us
			normal.X = sign(normal.X)
			normal.Y = 0
		}
		//log.Println("[fallToSurface] from", from, "normal", normal, "val", val, -val*stepSize)
		from = from.Add(normal.MulScalar(-val * stepSize))
	}
	return nil // Not found
}
