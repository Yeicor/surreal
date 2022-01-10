package sdf2

import (
	"fmt"
	"github.com/deadsy/sdfx/sdf"
)

func printSdf(s sdf.SDF2, cells sdf.V2i) {
	bb := s.BoundingBox().ScaleAboutCenter(1 + 2/cells.ToV2().MaxComponent()) // Add 1 extra cell
	cellSize := bb.Size().Div(cells.ToV2())
	var cellIndex sdf.V2i
	for cellIndex[0] = 0; cellIndex[0] < cells[0]; cellIndex[0]++ {
		for cellIndex[1] = 0; cellIndex[1] < cells[1]; cellIndex[1]++ {
			cellCenter := bb.Min.Add(cellIndex.ToV2().Mul(cellSize).Add(cellSize.DivScalar(2)))
			val := s.Evaluate(cellCenter)
			if val > 0 {
				fmt.Printf("+%07.2f ", val)
			} else {
				fmt.Printf("%08.2f ", val)
			}
		}
		fmt.Println()
	}
}
