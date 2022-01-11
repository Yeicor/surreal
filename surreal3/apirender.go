//go:build render
// +build render

package surreal3

import (
	"fmt"
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
)

func (a *Algorithm) Info(sdf3 sdf.SDF3, meshCells int) string {
	return fmt.Sprintf("%#+v", a) //TODO: implement me
}

func (a *Algorithm) Render(sdf3 sdf.SDF3, meshCells int, output chan<- *render.Triangle3) {
	for _, triangle := range a.Run(sdf3) {
		output <- &render.Triangle3{V: *triangle}
	}
}
