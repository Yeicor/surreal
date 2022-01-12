package surreal3

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"log"
	"math"
)

func (a *Algorithm) buildSingleSurface(s sdf.SDF3, firstEdge [2]sdf.V3) ([]*Triangle, *rtreego.Rtree) {
	var res []*Triangle
	rtreeTriangles := rtreego.NewTree(3, 3, 5)
	remaining := []*toProcess{{edge: firstEdge}, {edge: [2]sdf.V3{firstEdge[1], firstEdge[0]}}}
	for len(remaining) > 0 {
		curTriangleEdge := remaining[0]
		remaining = remaining[1:]
		curTriangleEnd := a.walkAlongSurface(s, curTriangleEdge, &remaining, rtreeTriangles)
		newTriangle := &Triangle{curTriangleEdge.edge[0], curTriangleEdge.edge[1], curTriangleEnd}
		rtreeTriangles.Insert(newTriangle)
		res = append(res, newTriangle)
	}
	return res, rtreeTriangles
}

func (a *Algorithm) walkAlongSurface(s sdf.SDF3, start *toProcess, remaining *[]*toProcess, rtreeTriangles *rtreego.Rtree) sdf.V3 {
	curPos := start.edge[0]
	tangentCross := start.edge[0]
	if math.Abs(start.edge[1].X) < math.MaxFloat64 {
		curPos = start.edge[0].Add(start.edge[1]).DivScalar(2)
		tangentCross = start.edge[1].Sub(start.edge[0])
	} else { // Otherwise, searching for the first edge
		if !math.Signbit(start.edge[1].X) { // Make sure the second search looks in a different direction
			tangentCross = tangentCross.Neg()
		}
	}
	startNormal := sdf.Normal3(s, curPos, a.normalEps)
	startTangent := sdf.V3{X: math.NaN()}
	moveCross := false
	for math.IsNaN(startTangent.X) || math.IsNaN(startTangent.Y) || math.IsNaN(startTangent.Z) {
		if moveCross { // Randomize a little to avoid NaNs
			tangentCross = tangentCross.Add(sdf.V3{X: a.rng.Float64(), Y: a.rng.Float64(), Z: a.rng.Float64()}.
				SubScalar(0.5).DivScalar(5)).Normalize()
		}
		moveCross = true
		startTangent = tangentForNormal(startNormal, tangentCross)
	}
	prevPos := curPos
	firstIter := true
	foundFirstAngle := false
	for {
		// Move by step to check if we are still good at the new point
		// TODO: Is it ok to fall outside the bounding box to continue the surface? (or should we force cut the surface
		//  and start generation in opposite direction from the initial vertex)
		prevPos = curPos
		curPos = curPos.Add(startTangent.MulScalar(a.step))
		curPosInSurface := fallToSurface(s, curPos, a.surfHitEps, a.normalEps, a.surfStepSize, a.surfMaxSteps, a.rng)
		if curPosInSurface == nil {
			panic("[SURREAL3] surface not found (while walking along tangent)")
		}
		curPos = *curPosInSurface
		newNormal := sdf.Normal3(s, curPos, a.normalEps)
		sharpAngle := false
		if firstIter {
			// This "hack" makes sure that the start normal is correct by displacing by epsilon and taking data there
			startNormal = newNormal
			startTangent = tangentForNormal(startNormal, tangentCross)
		} else {
			// This "hack" detects sharp angles: triggered when pos does not move as it falls back to the same position
			// (in this case we need a change of direction, as this is a very sharp corner: >= ~90º)
			movedBy := prevPos.Sub(curPos).Length2()
			sharpAngle = movedBy < a.step*a.step/1000
			if sharpAngle {
				log.Println("[SURREAL3] Found sharp angle (moved by", movedBy, "<", a.step*a.step/1000, "), firstIter:", firstIter)
			}
		}
		angle := math.Acos(startNormal.Dot(newNormal))
		shouldGenNewVertex := (!firstIter && angle >= a.minAngle) || sharpAngle
		if shouldGenNewVertex {
			if !foundFirstAngle {
				// Follow a second tangent once the first one reaches a valid angle (better vertex positioning)
				newTangent := startNormal.Cross(newNormal).Normalize()
				foundFirstAngle = true
				if !(math.IsNaN(newTangent.X) || math.IsNaN(newTangent.Y) || math.IsNaN(newTangent.Z)) {
					startTangent = newTangent
					startNormal = newNormal
					//log.Println("New tangent", startTangent, "and normal", startNormal)
					continue
				}
			}
		}
		//	log.Println("[SURREAL3] Pos:", prevPos, "->", curPos, "Normals:", startNormal, "->", newNormal,
		//		"Angle:", angle, ">=?", a.minAngle, "foundFirstAngle:", foundFirstAngle)
		if shouldGenNewVertex { // We need to place a vertex
			if remaining != nil {
				// Try to merge vertices (closing boundary)
				// FIXME: Proper merge generating more triangles and removing merged edge
				closestVert, closestVertDistSq, nearestTriangle := findNearest(rtreeTriangles, curPos, 3)
				canMerge := closestVertDistSq < a.step
				if canMerge {
					// Override the closest vert to the one that is closest to our start! (if they share the closest triangle, edge case)
					closestVertStart, _, nearestTriangleStart := findNearest(rtreeTriangles, start.edge[0], 2) // Skips start itself
					if nearestTriangleStart == nearestTriangle {
						closestVert = closestVertStart
					}
					closestVertStart, _, nearestTriangleStart = findNearest(rtreeTriangles, start.edge[1], 2) // Skips start itself
					if nearestTriangleStart == nearestTriangle {
						closestVert = closestVertStart
					}
				}
				//canMerge = canMerge /* && closestVert != start.edge[0] && closestVert != start.edge[1]*/
				log.Println("[SURREAL3] MERGE INFO:", curPos, "->", closestVert, "--", canMerge, "&&",
					closestVertDistSq, "<", a.step, "&& ...")
				if canMerge {
					// TODO: Generate 2 triangles (if not degenerate), fully connecting both separate triangles with a quad.
					off := 0
					for i, other := range *remaining { // Remove from boundaries to process (should be of len() 1)
						if other.edge[0] == closestVert || other.edge[1] == closestVert {
							*remaining = append((*remaining)[:i-off], (*remaining)[i-off+1:]...)
							off++
							//break
						}
					}
					log.Println("[SURREAL3] MERGE!", curPos, "->", closestVert, "removed", off)
					curPos = closestVert // Perfect close
				} else { // Mark as new boundary
					newProc := &toProcess{edge: [2]sdf.V3{start.edge[0], curPos}}
					*remaining = append(*remaining, newProc)
					newProc = &toProcess{edge: [2]sdf.V3{curPos, start.edge[1]}}
					*remaining = append(*remaining, newProc)
				}
			}
			//log.Println("[SURREAL3] walkAlongSurface finished with result:", curPos)
			return curPos
		} // Otherwise, continue moving forward without generating a vertex yet
		firstIter = false
	}
}
