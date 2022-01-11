package surreal2

import (
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"math"
)

func (a *Algorithm) buildSingleSurface(s sdf.SDF2, firstPoint sdf.V2) ([][2]sdf.V2, *rtreego.Rtree) {
	var res [][2]sdf.V2
	rtreeLines := rtreego.NewTree(2, 3, 5)
	remaining := []*toProcess{{firstPoint, false}}
	for len(remaining) > 0 {
		curLineStart := remaining[0]
		remaining = remaining[1:]
		newLineEnd := a.walkAlongSurface(s, curLineStart, &remaining, rtreeLines)
		newLine := [2]sdf.V2{curLineStart.point, newLineEnd}
		rtreeLines.Insert(&line{newLine})
		res = append(res, newLine)
	}
	return res, rtreeLines
}

func (a *Algorithm) walkAlongSurface(s sdf.SDF2, start *toProcess, remaining *[]*toProcess, rtreeLines *rtreego.Rtree) sdf.V2 {
	startNormal := sdf.Normal2(s, start.point, a.normalEps)
	startTangent := tangentForNormal(startNormal, start.dir)
	curPos := start.point
	prevPos := curPos
	firstIter := true
	for {
		// Move by step to check if we are still good at the new point
		// TODO: Is it ok to fall outside the bounding box to continue the surface? (or should we force cut the surface
		//  and start generation in opposite direction from the initial vertex)
		prevPos = curPos
		curPos = curPos.Add(startTangent.MulScalar(a.step))
		curPosInSurface := fallToSurface(s, curPos, a.surfHitEps, a.normalEps, a.surfStepSize, a.surfMaxSteps, a.rng)
		if curPosInSurface == nil {
			panic("[SURREAL2] surface not found (while walking along tangent)")
		}
		curPos = *curPosInSurface
		newNormal := sdf.Normal2(s, curPos, a.normalEps)
		sharpAngle := false
		if firstIter {
			// This "hack" makes sure that the start normal is correct by displacing by epsilon and taking data there
			startNormal = newNormal
			startTangent = tangentForNormal(startNormal, start.dir)
		} else {
			// This "hack" detects sharp angles: triggered when pos does not move as it falls back to the same position
			// (in this case we need a change of direction, as this is a very sharp corner: >= ~90º)
			movedBy := prevPos.Sub(curPos).Length2()
			sharpAngle = movedBy < a.step*a.step/1000
			if sharpAngle {
				//log.Println("[SURREAL2] Found sharp angle on first iteration (moved by", movedBy, "<", a.step*a.step/1000, "), firstIter:", firstIter)
			}
		}
		angle := math.Acos(startNormal.Dot(newNormal))
		//log.Println("[SURREAL2] Pos:", prevPos, "->", curPos, "Normals:", startNormal, "->", newNormal, "Angle:", angle, ">=?", a.minAngle)
		if (!firstIter && angle >= a.minAngle) || sharpAngle { // We need to place a vertex
			if remaining != nil {
				// Try to merge vertices (closing boundary)
				closestVert, closestVertDistSq, nearestLine := findNearest(rtreeLines, curPos, 1)
				canMerge := closestVertDistSq < a.step
				if canMerge {
					// Override the closest vert to the one that is closest to our start! (if they share the closest line, edge case)
					closestVertStart, _, nearestLineStart := findNearest(rtreeLines, start.point, 2) // Skips start itself
					if nearestLineStart == nearestLine {
						closestVert = closestVertStart
					}
				}
				canMerge = canMerge && closestVert != start.point
				//log.Println("[SURREAL2] MERGE INFO:", curPos, "->", closestVert, "--", canMerge, "&&",
				//	closestVertDistSq, "<", a.step, "&&", closestVert != start.point)
				if canMerge {
					//log.Println("[SURREAL2] MERGE!", curPos, "->", closestVert)
					for i, other := range *remaining { // Remove from boundaries to process (should be of len() 1)
						if other.point == closestVert && other.dir != start.dir {
							*remaining = append((*remaining)[:i], (*remaining)[i+1:]...)
							break
						}
					}
					curPos = closestVert // Perfect close
				} else { // Mark as new boundary
					newProc := &toProcess{point: curPos, dir: start.dir}
					*remaining = append(*remaining, newProc)
				}
			}
			//log.Println("[SURREAL2] walkAlongSurface finished with result:", curPos)
			return curPos
		} // Otherwise, continue moving forward without generating a vertex yet
		firstIter = false
	}
}
