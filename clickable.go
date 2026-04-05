// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// Clickable allocates a layout cell with no default frame. The draw callback paints the cell;
// clicks fire the EventHandler from [EventHandler.On].
//
// A Clickable widget is uniquely determined by its call location. Use [IDScope] in loops.
func (c *Context) Clickable(draw func(bounds image.Rectangle)) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.widget(id, optionNoFrame, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
			if c.pointing.justPressed() && c.focus == id {
				return &eventHandler{}
			}
			return nil
		}, draw)
	})
}
