package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"math"
)

func findNearest(rtreeLines *rtreego.Rtree, newPos sdf.V3, numNeighbors int) (sdf.V3, float64, *Triangle) {
	allNearest := rtreeLines.NearestNeighbors(numNeighbors, rtreego.Point{newPos.X, newPos.Y, newPos.Z})
	closestVertDistSq := math.MaxFloat64
	var closestVert sdf.V3
	var nearestTri *Triangle
	for _, nearest := range allNearest {
		nearestTri = nearest.(*Triangle)
		closestVert = nearestTri[0]
		closestVertDistSq = nearestTri[0].Sub(newPos).Length2()
		closestVertDistSq2 := nearestTri[1].Sub(newPos).Length2()
		closestVertDistSq3 := nearestTri[2].Sub(newPos).Length2()
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
