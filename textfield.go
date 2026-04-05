// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// textSelectionFill is drawn behind selected text in focused TextFields.
var textSelectionFill = color.RGBA{R: 70, G: 120, B: 200, A: 110}

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

// TextField creates a text field to modify the value of a string buf.
//
// TextField returns an EventHandler to handle events when the value is confirmed, such as on blur or Enter key press.
// A returned EventHandler is never nil.
//
// A TextField widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) TextField(buf *string) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.textField(buf, id, 0)
	})
}

func textFieldTextX(c *Context, bounds image.Rectangle, opt option, textw int) int {
	ofx := bounds.Dx() - c.style().Padding - textw - 1
	textx := bounds.Min.X + min(ofx, c.style().Padding)
	switch {
	case opt&optionAlignCenter != 0:
		textx = bounds.Min.X + (bounds.Dx()-textw)/2
	case opt&optionAlignRight != 0:
		textx = bounds.Min.X + bounds.Dx() - textw - c.style().Padding
	}
	return textx
}

func textFieldTextY(bounds image.Rectangle) int {
	texth := lineHeight()
	return bounds.Min.Y + (bounds.Dy()-texth)/2
}

// byteIndexAtX maps a horizontal offset in pixels from the left edge of the text to a UTF-8 byte index.
func byteIndexAtX(text string, xOff int) int {
	if xOff <= 0 || len(text) == 0 {
		return 0
	}
	for i := 0; i < len(text); {
		_, size := utf8.DecodeRuneInString(text[i:])
		if size == 0 {
			break
		}
		next := i + size
		wNext := textAdvancePrefix(text[:next])
		wPrev := textAdvancePrefix(text[:i])
		if xOff < wNext {
			if xOff-wPrev < wNext-xOff {
				return i
			}
			return next
		}
		i = next
	}
	return len(text)
}

func clampTextSel(i, maxB int) int {
	return clamp(i, 0, maxB)
}

func (c *Context) textFieldRaw(buf *string, id widgetID, opt option) (EventHandler, error) {
	return c.widget(id, opt|optionHoldFocus, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler

		f := c.currentContainer().textInputTextField(id, true)
		if c.focus == id {
			f.Focus()
			disp := f.TextForRendering()
			committedMax := len(f.Text())
			pos := c.pointingPosition()
			hover := c.pointingOver(bounds)
			textw := textWidth(disp)
			textx := textFieldTextX(c, bounds, opt, textw)
			texty := textFieldTextY(bounds)
			texth := lineHeight()

			if c.pointing.justPressed() && hover {
				now := time.Now()
				dt := now.Sub(c.textFieldLastClickAt)
				dx := pos.X - c.textFieldLastClickX
				dy := pos.Y - c.textFieldLastClickY
				isDouble := c.textFieldLastClickID == id && dt < 450*time.Millisecond && dx*dx+dy*dy < 100

				clickX := pos.X - textx
				idx := clampTextSel(byteIndexAtX(disp, clickX), committedMax)

				if isDouble {
					t := f.Text()
					f.SetTextAndSelection(t, 0, len(t))
					*buf = t
					c.textFieldDragging = false
				} else {
					f.SetSelection(idx, idx)
					*buf = f.Text()
					c.textFieldDragging = true
					c.textFieldDragWidget = id
					c.textFieldDragAnchor = idx
				}

				c.textFieldLastClickAt = now
				c.textFieldLastClickID = id
				c.textFieldLastClickX = pos.X
				c.textFieldLastClickY = pos.Y
			} else if c.textFieldDragging && c.textFieldDragWidget == id && c.pointing.pressed() {
				// Drag to extend selection (not on the mousedown frame; see justPressed block above).
				clickX := pos.X - textx
				cur := clampTextSel(byteIndexAtX(disp, clickX), committedMax)
				a := c.textFieldDragAnchor
				start, end := a, cur
				if start > end {
					start, end = end, start
				}
				f.SetSelection(start, end)
				*buf = f.Text()
			}

			if !c.pointing.pressed() {
				c.textFieldDragging = false
			}

			selStart, selEnd := f.Selection()
			caretByte := selEnd
			if u := f.UncommittedTextLengthInBytes(); u > 0 {
				caretByte = selStart + u
			}
			if caretByte > len(disp) {
				caretByte = len(disp)
			}
			caretX := textx + textAdvancePrefix(disp[:caretByte])
			imeBounds := image.Rect(caretX, texty, caretX+max(2, textAdvancePrefix("M")), texty+texth)

			handled, err := f.HandleInputWithBounds(imeBounds)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}
			if *buf != f.Text() {
				*buf = f.Text()
			}

			if !handled {
				ctrl := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
				if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyA) {
					t := f.Text()
					f.SetTextAndSelection(t, 0, len(t))
					*buf = t
				} else if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
					s0, s1 := f.Selection()
					t := f.Text()
					if s0 != s1 {
						lo, hi := s0, s1
						if lo > hi {
							lo, hi = hi, lo
						}
						newText := t[:lo] + t[hi:]
						f.SetTextAndSelection(newText, lo, lo)
						*buf = newText
					} else if len(t) > 0 {
						_, size := utf8.DecodeLastRuneInString(t)
						newText := t[:len(t)-size]
						f.SetTextAndSelection(newText, len(newText), len(newText))
						*buf = newText
					}
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
					e = &eventHandler{}
				}
			}
		} else {
			if *buf != f.Text() {
				f.SetTextAndSelection(*buf, len(*buf), len(*buf))
			}
			if wasFocused {
				e = &eventHandler{}
			}
		}
		return e
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorBase, opt)
		if c.focus == id {
			f := c.currentContainer().textInputTextField(id, true)
			disp := f.TextForRendering()
			tc := c.style().widgetColor(colorText)
			textw := textWidth(disp)
			textx := textFieldTextX(c, bounds, opt, textw)
			texty := textFieldTextY(bounds)
			texth := lineHeight()

			c.pushClipRect(bounds)

			if f.UncommittedTextLengthInBytes() == 0 {
				selStart, selEnd := f.Selection()
				if selStart != selEnd {
					lo, hi := selStart, selEnd
					if lo > hi {
						lo, hi = hi, lo
					}
					if hi <= len(disp) {
						x0 := textx + textAdvancePrefix(disp[:lo])
						x1 := textx + textAdvancePrefix(disp[:hi])
						if x1 < x0 {
							x0, x1 = x1, x0
						}
						c.drawRect(image.Rect(x0, texty, x1, texty+texth), textSelectionFill)
					}
				}
			}

			c.drawText(disp, image.Pt(textx, texty), tc)

			selStart, selEnd := f.Selection()
			caretByte := selEnd
			if u := f.UncommittedTextLengthInBytes(); u > 0 {
				caretByte = selStart + u
			}
			if caretByte > len(disp) {
				caretByte = len(disp)
			}
			caretX := textx + textAdvancePrefix(disp[:caretByte])
			if selStart == selEnd {
				c.drawRect(image.Rect(caretX, texty, caretX+1, texty+texth), tc)
			}

			c.popClipRect()
		} else {
			c.drawWidgetText(*buf, bounds, colorText, opt)
		}
	})
}

// SetTextFieldValue sets the value of the current text field.
//
// If the last widget is not a text field, this function does nothing.
func (c *Context) SetTextFieldValue(value string) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if f := c.currentContainer().textInputTextField(c.currentID, false); f != nil {
			f.SetTextAndSelection(value, 0, 0)
		}
		return nil, nil
	})
}

func (c *Context) textField(buf *string, id widgetID, opt option) (EventHandler, error) {
	return c.textFieldRaw(buf, id, opt)
}

// NumberField creates a number field to modify the value of a int value.
//
// step is the amount to increment or decrement the value when the user drags the mouse cursor.
//
// NumberField returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A NumberField widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberField(value *int, step int) EventHandler {
	pc := caller()
	idPart := idPartFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.numberField(value, step, idPart, optionAlignRight)
	})
}

// NumberFieldF creates a number field to modify the value of a float64 value.
//
// step is the amount to increment or decrement the value when the user drags the mouse cursor.
// digits is the number of decimal places to display.
//
// NumberFieldF returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A NumberFieldF widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberFieldF(value *float64, step float64, digits int) EventHandler {
	pc := caller()
	idPart := idPartFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.numberFieldF(value, step, digits, idPart, optionAlignRight)
	})
}

func (c *Context) numberField(value *int, step int, idPart string, opt option) (EventHandler, error) {
	last := *value

	var e EventHandler
	var err error
	c.idScopeFromIDPart(idPart, func(id widgetID) {
		c.GridCell(func(bounds image.Rectangle) {
			c.SetGridLayout([]int{-1, lineHeight()}, nil)

			buf := fmt.Sprintf("%d", *value)
			e1, err1 := c.textFieldRaw(&buf, id, opt)
			if err1 != nil {
				err = err1
				return
			}
			if e1 != nil {
				e1.On(func() {
					c.setFocus(widgetID{})
					v, err := strconv.ParseInt(buf, 10, 64)
					if err != nil {
						v = 0
					}
					*value = int(v)
					if *value != last {
						e = &eventHandler{}
					}
				})
			}
			if c.focus == id {
				var updated bool
				if keyRepeated(ebiten.KeyUp) || keyRepeated(ebiten.KeyDown) {
					v, err := strconv.ParseInt(buf, 10, 64)
					if err != nil {
						v = 0
					}
					*value = int(v)
					updated = true
					if keyRepeated(ebiten.KeyUp) {
						*value += step
					}
					if keyRepeated(ebiten.KeyDown) {
						*value -= step
						updated = true
					}
				}
				if updated {
					buf := fmt.Sprintf("%d", *value)
					if f := c.currentContainer().textInputTextField(id, false); f != nil {
						f.SetTextAndSelection(buf, len(buf), len(buf))
					}
					e = &eventHandler{}
				}
			}
			if c.hover == id && ebiten.IsKeyPressed(ebiten.KeyControl) {
				_, wy := ebiten.Wheel()
				if wy != 0 && step != 0 {
					if wy < 0 {
						*value += step
					} else {
						*value -= step
					}
					buf := fmt.Sprintf("%d", *value)
					if f := c.currentContainer().textInputTextField(id, false); f != nil {
						f.SetTextAndSelection(buf, len(buf), len(buf))
					}
					c.wheelConsumed = true
					e = &eventHandler{}
				}
			}

			c.GridCell(func(bounds image.Rectangle) {
				c.SetGridLayout(nil, []int{-1, -1})
				up, down := c.spinButtons(id)
				up.On(func() {
					*value += step
					e = &eventHandler{}
				})
				down.On(func() {
					*value -= step
					e = &eventHandler{}
				})
			})
		})
	})

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (c *Context) numberFieldF(value *float64, step float64, digits int, idPart string, opt option) (EventHandler, error) {
	last := *value

	var e EventHandler
	var err error
	c.idScopeFromIDPart(idPart, func(id widgetID) {
		c.GridCell(func(bounds image.Rectangle) {
			c.SetGridLayout([]int{-1, lineHeight()}, nil)

			buf := formatNumber(*value, digits)
			e1, err1 := c.textFieldRaw(&buf, id, opt)
			if err1 != nil {
				err = err1
				return
			}
			if e1 != nil {
				e1.On(func() {
					c.setFocus(widgetID{})
					v, err := strconv.ParseFloat(buf, 64)
					if err != nil {
						v = 0
					}
					*value = float64(v)
					if *value != last {
						e = &eventHandler{}
					}
				})
			}
			if c.focus == id {
				var updated bool
				if keyRepeated(ebiten.KeyUp) || keyRepeated(ebiten.KeyDown) {
					v, err := strconv.ParseFloat(buf, 64)
					if err != nil {
						v = 0
					}
					*value = float64(v)
					updated = true
					if keyRepeated(ebiten.KeyUp) {
						*value += step
					}
					if keyRepeated(ebiten.KeyDown) {
						*value -= step
						updated = true
					}
				}
				if updated {
					buf := formatNumber(*value, digits)
					if f := c.currentContainer().textInputTextField(id, false); f != nil {
						f.SetTextAndSelection(buf, len(buf), len(buf))
					}
					e = &eventHandler{}
				}
			}
			if c.hover == id && ebiten.IsKeyPressed(ebiten.KeyControl) {
				_, wy := ebiten.Wheel()
				if wy != 0 && step != 0 {
					if wy < 0 {
						*value += step
					} else {
						*value -= step
					}
					buf := formatNumber(*value, digits)
					if f := c.currentContainer().textInputTextField(id, false); f != nil {
						f.SetTextAndSelection(buf, len(buf), len(buf))
					}
					c.wheelConsumed = true
					e = &eventHandler{}
				}
			}

			c.GridCell(func(bounds image.Rectangle) {
				c.SetGridLayout(nil, []int{-1, -1})
				up, down := c.spinButtons(id)
				up.On(func() {
					*value += step
					e = &eventHandler{}
				})
				down.On(func() {
					*value -= step
					e = &eventHandler{}
				})
			})
		})
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}
