package surreal2

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"math"
)

func sign(v float64) float64 {
	if v >= 0 {
		return 1
	}
	return -1
}

func findNearest(rtreeLines *rtreego.Rtree, newPos sdf.V2, numNeighbors int) (sdf.V2, float64, *line) {
	allNearest := rtreeLines.NearestNeighbors(numNeighbors, rtreego.Point{newPos.X, newPos.Y})
	closestVertDistSq := math.MaxFloat64
	var closestVert sdf.V2
	var nearestLine *line
	for _, nearest := range allNearest {
		nearestLine = nearest.(*line)
		closestVert = nearestLine.vertices[0]
		//if closestVert == newPos {
		//	continue
		//}
		closestVertDistSq = nearestLine.vertices[0].Sub(newPos).Length2()
		closestVertDistSq2 := nearestLine.vertices[1].Sub(newPos).Length2()
		if closestVertDistSq2 < closestVertDistSq {
			closestVertDistSq = closestVertDistSq2
			closestVert = nearestLine.vertices[1]
		}
	}
	return closestVert, closestVertDistSq, nearestLine
}

// containsPoint tests whether p is located inside or on the boundary of r.
func rectContainsPoint(r *rtreego.Rect, p sdf.V2, eps float64) bool {
	if p.X+eps < r.PointCoord(0) || p.X-eps > r.PointCoord(0)+r.LengthsCoord(0) {
		return false
	}
	if p.Y+eps < r.PointCoord(1) || p.Y-eps > r.PointCoord(1)+r.LengthsCoord(1) {
		return false
	}
	return true
}

type line struct {
	vertices [2]sdf.V2
}

func (p *line) Bounds() *rtreego.Rect {
	p1 := rtreego.Point{p.vertices[0].X, p.vertices[0].Y}
	p2 := rtreego.Point{p.vertices[1].X, p.vertices[1].Y}
	rect, _ := rtreego.NewRectFromPoints(p1, p2)
	return rect
}

type toProcess struct {
	point sdf.V2
	dir   bool
}

func tangentForNormal(startNormal sdf.V2, clockwise bool) (tangent sdf.V2) {
	if clockwise {
		tangent = sdf.V2{X: startNormal.Y, Y: -startNormal.X}
	} else {
		tangent = sdf.V2{X: -startNormal.Y, Y: startNormal.X}
	}
	return
}
