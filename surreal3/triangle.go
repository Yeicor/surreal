package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"math"
)

type Triangle [3]sdf.V3

func (t *Triangle) Bounds() *rtreego.Rect {
	b := sdf.Box3{
		Min: sdf.V3{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64},
		Max: sdf.V3{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64},
	}
	b = b.Include(t[0])
	b = b.Include(t[1])
	b = b.Include(t[2])
	rect, _ := rtreego.NewRectFromPoints(
		rtreego.Point{b.Min.X, b.Min.Y, b.Min.Z},
		rtreego.Point{b.Max.X, b.Max.Y, b.Max.Z},
	)
	return rect
}

// Normal returns the normal vector to the plane defined by the 3D Triangle.
func (t *Triangle) Normal() sdf.V3 {
	e1 := t[1].Sub(t[0])
	e2 := t[2].Sub(t[0])
	return e1.Cross(e2).Normalize()
}

// Center returns the center point of the 3D Triangle.
func (t *Triangle) Center() sdf.V3 {
	return t[0].Add(t[1]).Add(t[2]).DivScalar(3)
}
