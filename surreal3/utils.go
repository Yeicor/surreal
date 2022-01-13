package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"math"
)

func findNearest(rtreeLines *rtreego.Rtree, pos sdf.V3, startEdge [2]sdf.V3, numNeighbors int) (sdf.V3, float64, *Triangle) {
	const tieEps = 1e-23
	triangle := &Triangle{startEdge[0], startEdge[1], pos}
	triCenter := triangle.Center()
	allNearest := rtreeLines.NearestNeighbors(numNeighbors, rtreego.Point{pos.X, pos.Y, pos.Z})
	closestVertDistSq := math.MaxFloat64
	var closestVert sdf.V3
	var nearestTri *Triangle
	for _, nearest := range allNearest {
		// TODO: Use startEdge metadata for actually detecting the nearest triangle
		nearestTri = nearest.(*Triangle)
		closestVert = nearestTri[0]
		closestVertDistSq1 := nearestTri[0].Sub(pos).Length2() + nearestTri[0].Sub(triCenter).Length2()*tieEps
		closestVertDistSq2 := nearestTri[1].Sub(pos).Length2() + nearestTri[1].Sub(triCenter).Length2()*tieEps
		closestVertDistSq3 := nearestTri[2].Sub(pos).Length2() + nearestTri[2].Sub(triCenter).Length2()*tieEps
		if closestVertDistSq1 < closestVertDistSq {
			closestVertDistSq = closestVertDistSq1
			closestVert = nearestTri[0]
		}
		if closestVertDistSq2 < closestVertDistSq {
			closestVertDistSq = closestVertDistSq2
			closestVert = nearestTri[1]
		}
		if closestVertDistSq3 < closestVertDistSq {
			closestVertDistSq = closestVertDistSq3
			closestVert = nearestTri[2]
		}
	}
	return closestVert, closestVertDistSq, nearestTri
}

type toProcess struct {
	edge [2]sdf.V3
}

func tangentForNormal(startNormal, cross sdf.V3) sdf.V3 {
	return startNormal.Cross(cross).Normalize()
}
