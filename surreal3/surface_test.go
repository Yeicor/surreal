package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"math"
	"math/rand"
	"testing"
)

func TestFallToSurface(t *testing.T) {
	s, _ := sdf.Box3D(sdf.V3{X: 1, Y: 1, Z: 1}, 0)
	hitEps := 1e-3
	hit := fallToSurface(s, sdf.V3{}, hitEps, 1e-10, 1, 100, rand.New(rand.NewSource(0)))
	// Check that it falls in any random surface of the cube
	if hit == nil || math.Abs(math.Abs(hit.X)-0.5) > hitEps && math.Abs(math.Abs(hit.Y)-0.5) > hitEps && math.Abs(math.Abs(hit.Z)-0.5) > hitEps {
		t.Fatal("Hit point not on surface", hit)
	}
}
