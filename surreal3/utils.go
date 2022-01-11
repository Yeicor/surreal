package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
)

func sign(v float64) float64 {
	if v >= 0 {
		return 1
	}
	return -1
}

func bestTangentForNormal(startNormal sdf.V3) (tangent sdf.V3) {
	// TODO: Find the best tangent by sampling normals nearby to find an edge (largest normal change)
	//if clockwise {
	//	tangent = sdf.V2{X: startNormal.Y, Y: -startNormal.X}
	//} else {
	//	tangent = sdf.V2{X: -startNormal.Y, Y: startNormal.X}
	//}
	return
}
