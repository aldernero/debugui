// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ConsumeSecondaryClick reports whether the right mouse button was just pressed while the
// cursor lies inside bounds (layout coordinates) for the current root container.
//
// Typical use: call at the end of a [GridCell] callback after laying out hit-testable widgets,
// so bounds match the cell the widgets occupy.
func (c *Context) ConsumeSecondaryClick(bounds image.Rectangle) bool {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		return false
	}
	return c.pointingOver(bounds)
}
