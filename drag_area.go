// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// DragArea is an interactive region with custom screen drawing. While the pointer is pressed
// and this widget holds focus, onDrag is called with the pointer position clamped to bounds.
// If onDrag returns true, the returned [EventHandler] runs its [EventHandler.On] callback
// (same pattern as [SliderF] value changes).
//
// A DragArea widget is uniquely determined by its call location. Use [IDScope] in loops.
func (c *Context) DragArea(
	screenDraw func(screen *ebiten.Image, bounds image.Rectangle),
	onDrag func(bounds image.Rectangle, pos image.Point) bool,
) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.widget(id, optionNoFrame, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
			if c.focus == id && c.pointing.pressed() {
				p := c.pointingPosition()
				q := p
				if q.X < bounds.Min.X {
					q.X = bounds.Min.X
				}
				if q.Y < bounds.Min.Y {
					q.Y = bounds.Min.Y
				}
				if q.X > bounds.Max.X-1 {
					q.X = bounds.Max.X - 1
				}
				if q.Y > bounds.Max.Y-1 {
					q.Y = bounds.Max.Y - 1
				}
				if onDrag(bounds, q) {
					return &eventHandler{}
				}
			}
			return nil
		}, func(bounds image.Rectangle) {
			c.setClip(c.clipRect())
			defer c.setClip(unclippedRect)
			cmd := c.appendCommand(commandDraw)
			cmd.draw.f = func(screen *ebiten.Image) {
				screenDraw(screen, bounds)
			}
		})
	})
}
