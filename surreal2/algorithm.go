package surreal2

import (
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	"github.com/dhconnelly/rtreego"
	"log"
	"math"
)

func (a *Algorithm) buildSingleSurface(s sdf.SDF2, firstPoint sdf.V2) ([]*render.Line, *rtreego.Rtree) {
	var res []*render.Line
	rtreeLines := rtreego.NewTree(2, 3, 5)
	remaining := []*toProcess{{firstPoint, false}}
	for len(remaining) > 0 {
		curLineStart := remaining[0]
		remaining = remaining[1:]
		newLine := a.runIter(s, curLineStart, &remaining, rtreeLines)
		if newLine != nil {
			rtreeLines.Insert(&line{newLine})
			res = append(res, newLine)
		}
	}
	return res, rtreeLines
}

func (a *Algorithm) runIter(s sdf.SDF2, start *toProcess, remaining *[]*toProcess, rtreeLines *rtreego.Rtree) *render.Line {
	startNormal := sdf.Normal2(s, start.point, a.normalEps)
	startTangent := tangentForNormal(startNormal, start.dir)
	newPos := start.point
	prevPos := newPos
	for {
		firstIter := newPos == start.point
		// Move by step to check if we are still good at the new point
		// TODO: Is it ok to fall outside the bounding box to continue the surface?
		newPos = newPos.Add(startTangent.MulScalar(a.step))
		newPosInSurface := fallToSurface(s, newPos, a.surfHitEps, a.normalEps, a.surfStepSize, a.surfMaxSteps, a.rng)
		if newPosInSurface == nil {
			log.Println("[SURREAL2] WARNING: Surface not found (while walking along tangent)")
			return nil
		}
		newPos = *newPosInSurface
		newNormal := sdf.Normal2(s, newPos, a.normalEps)
		sharpAngle := false
		if firstIter {
			// This "hack" makes sure that the start normal is correct by displacing by epsilon and taking data there
			startNormal = newNormal
			startTangent = tangentForNormal(startNormal, start.dir)
		} else {
			// This other "hack" triggered when pos does not move as it falls back to the same position over and over
			// (in this case we need a change of direction, as this is a very sharp corner)
			sharpAngle = prevPos.Sub(newPos).Length2() < a.step*a.step/1000
			//log.Println("[SURREAL2] WARNING: Stuck triggered (while walking along tangent)")
		}
		angle := math.Acos(startNormal.Dot(newNormal))
		//log.Println("newPos:", newPos, "Normals:", startNormal, newNormal, "Angle:", angle, ">=?", a.minAngle)
		if angle >= a.minAngle || sharpAngle { // We need to place a vertex
			if remaining != nil {
				// Try to merge vertices (closing boundary)
				closestVert, closestVertDistSq, nearestLine := findNearest(rtreeLines, newPos, 1)
				canMerge := closestVertDistSq < a.step && rectContainsPoint(nearestLine.Bounds(), newPos, a.step)
				if canMerge {
					// Override the closest vert to the one that is closest to our start! (if they share the closest line, edge case)
					closestVertStart, _, nearestLineStart := findNearest(rtreeLines, start.point, 2) // Skips start itself
					if nearestLineStart == nearestLine {
						closestVert = closestVertStart
					}
				}
				if canMerge && closestVert != start.point {
					//log.Println("MERGE!")
					for i, other := range *remaining { // Remove from boundaries to process (should be of len() 1)
						if other.point == closestVert && other.dir != start.dir {
							*remaining = append((*remaining)[:i], (*remaining)[i+1:]...)
							break
						}
					}
					newPos = closestVert // Perfect close
				} else { // Mark as new boundary
					newProc := &toProcess{point: newPos, dir: start.dir}
					*remaining = append(*remaining, newProc)
				}
			}
			return &render.Line{start.point, newPos}
		} // Otherwise, continue moving forward without generating a vertex yet
	}
}
