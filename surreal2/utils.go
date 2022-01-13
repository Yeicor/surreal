package surreal2

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"math"
)

func findNearest(rtreeLines *rtreego.Rtree, newPos sdf.V2, from sdf.V2, numNeighbors int) (sdf.V2, float64, *line) {
	const tieEps = 1e-23
	center := newPos.Add(from.DivScalar(2))
	allNearest := rtreeLines.NearestNeighbors(numNeighbors, rtreego.Point{newPos.X, newPos.Y})
	closestVertDistSq := math.MaxFloat64
	var closestVert sdf.V2
	var closestLine *line
	for _, nearest := range allNearest {
		sampledLine := nearest.(*line)
		closestVertDistSq1 := sampledLine.vertices[0].Sub(newPos).Length2() + sampledLine.vertices[0].Sub(center).Length2()*tieEps
		closestVertDistSq2 := sampledLine.vertices[1].Sub(newPos).Length2() + sampledLine.vertices[1].Sub(center).Length2()*tieEps
		if closestVertDistSq1 < closestVertDistSq {
			closestVertDistSq = closestVertDistSq1
			closestVert = sampledLine.vertices[0]
			closestLine = sampledLine
		}
		if closestVertDistSq2 < closestVertDistSq {
			closestVertDistSq = closestVertDistSq2
			closestVert = sampledLine.vertices[1]
			closestLine = sampledLine
		}
	}
	return closestVert, closestVertDistSq, closestLine
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
