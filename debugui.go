// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// DebugUI is a debug UI.
//
// The zero value for DebugUI is ready to use.
type DebugUI struct {
	ctx           Context
	styleStorage  Style
	styleActive   bool
}

// InputCapturingState is a bit mask that indicates the input capturing state of the debug UI.
type InputCapturingState int

const (
	// InputCapturingStateHover indicates that a pointing device is hovering over a widget.
	InputCapturingStateHover InputCapturingState = 1 << iota

	// InputCapturingStateFocus indicates that a widget like a text field is focused.
	InputCapturingStateFocus
)

// Update updates the debug UI.
//
// Update returns true if the debug UI is capturing input, e.g. when a widget has focus.
// Otherwise, Update returns false.
//
// Update should be called once in the game's Update function.
func (d *DebugUI) Update(f func(ctx *Context) error) (InputCapturingState, error) {
	inputCapturingState, err := d.ctx.update(f)
	if err != nil {
		return 0, err
	}
	return inputCapturingState, nil
}

// Draw draws the debug UI.
//
// Draw should be called once in the game's Draw function.
func (d *DebugUI) Draw(screen *ebiten.Image) {
	d.ctx.draw(screen)
	d.ctx.screenWidth, d.ctx.screenHeight = screen.Bounds().Dx(), screen.Bounds().Dy()
}

// SetStyle replaces the UI style. Pass nil to restore [DefaultStyle].
// The style is copied; later changes to the argument are ignored.
func (d *DebugUI) SetStyle(s *Style) {
	if s == nil {
		d.styleActive = false
		d.ctx.stylePtr = nil
		return
	}
	d.styleStorage = *s
	d.styleActive = true
	d.ctx.stylePtr = &d.styleStorage
}

// Style returns a copy of the active style, or [DefaultStyle] if none was set.
func (d *DebugUI) Style() Style {
	if !d.styleActive {
		return DefaultStyle()
	}
	return d.styleStorage
}

// LoadStyleFile reads a JSON theme from path and applies it. See [LoadStyleReader] for the format.
func (d *DebugUI) LoadStyleFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s, err := ParseStyleJSON(b)
	if err != nil {
		return err
	}
	d.SetStyle(&s)
	return nil
}

// LoadStyleReader reads a JSON theme from r and applies it. See [ParseStyleJSON].
func (d *DebugUI) LoadStyleReader(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s, err := ParseStyleJSON(b)
	if err != nil {
		return err
	}
	d.SetStyle(&s)
	return nil
}
