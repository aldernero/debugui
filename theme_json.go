// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// themeFileJSON is the on-disk JSON shape. Omitted fields keep values from [DefaultStyle].
type themeFileJSON struct {
	Version       *int           `json:"version"`
	DefaultWidth  *int           `json:"defaultWidth"`
	DefaultHeight *int           `json:"defaultHeight"`
	Padding       *int           `json:"padding"`
	Spacing       *int           `json:"spacing"`
	Indent        *int           `json:"indent"`
	TitleHeight   *int           `json:"titleHeight"`
	ScrollbarSize *int           `json:"scrollbarSize"`
	ThumbSize     *int           `json:"thumbSize"`
	Colors        *colorsFileJSON `json:"colors"`
}

type colorsFileJSON struct {
	Text               *string `json:"text"`
	Border             *string `json:"border"`
	WindowBG           *string `json:"windowBG"`
	TitleBG            *string `json:"titleBG"`
	TitleBGTransparent *string `json:"titleBGTransparent"`
	TitleText          *string `json:"titleText"`
	PanelBG            *string `json:"panelBG"`
	Button             *string `json:"button"`
	ButtonHover        *string `json:"buttonHover"`
	ButtonFocus        *string `json:"buttonFocus"`
	SliderThumb        *string `json:"sliderThumb"`
	SliderThumbHover   *string `json:"sliderThumbHover"`
	SliderThumbFocus   *string `json:"sliderThumbFocus"`
	Base               *string `json:"base"`
	BaseHover          *string `json:"baseHover"`
	BaseFocus          *string `json:"baseFocus"`
	ScrollBase         *string `json:"scrollBase"`
	ScrollThumb        *string `json:"scrollThumb"`
}

// ParseStyleJSON parses a theme JSON document and merges it onto [DefaultStyle].
// Color values are hex strings: "#RRGGBB", "#RRGGBBAA", or "#RGB".
//
// The optional "version" field is ignored for now; it is reserved for future format changes.
func ParseStyleJSON(data []byte) (Style, error) {
	var raw themeFileJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return Style{}, err
	}
	s := DefaultStyle()
	if raw.DefaultWidth != nil {
		s.DefaultWidth = *raw.DefaultWidth
	}
	if raw.DefaultHeight != nil {
		s.DefaultHeight = *raw.DefaultHeight
	}
	if raw.Padding != nil {
		s.Padding = *raw.Padding
	}
	if raw.Spacing != nil {
		s.Spacing = *raw.Spacing
	}
	if raw.Indent != nil {
		s.Indent = *raw.Indent
	}
	if raw.TitleHeight != nil {
		s.TitleHeight = *raw.TitleHeight
	}
	if raw.ScrollbarSize != nil {
		s.ScrollbarSize = *raw.ScrollbarSize
	}
	if raw.ThumbSize != nil {
		s.ThumbSize = *raw.ThumbSize
	}
	if raw.Colors != nil {
		c := &s.Colors
		if err := mergeColor(&c.Text, raw.Colors.Text); err != nil {
			return Style{}, fmt.Errorf("colors.text: %w", err)
		}
		if err := mergeColor(&c.Border, raw.Colors.Border); err != nil {
			return Style{}, fmt.Errorf("colors.border: %w", err)
		}
		if err := mergeColor(&c.WindowBG, raw.Colors.WindowBG); err != nil {
			return Style{}, fmt.Errorf("colors.windowBG: %w", err)
		}
		if err := mergeColor(&c.TitleBG, raw.Colors.TitleBG); err != nil {
			return Style{}, fmt.Errorf("colors.titleBG: %w", err)
		}
		if err := mergeColor(&c.TitleBGTransparent, raw.Colors.TitleBGTransparent); err != nil {
			return Style{}, fmt.Errorf("colors.titleBGTransparent: %w", err)
		}
		if err := mergeColor(&c.TitleText, raw.Colors.TitleText); err != nil {
			return Style{}, fmt.Errorf("colors.titleText: %w", err)
		}
		if err := mergeColor(&c.PanelBG, raw.Colors.PanelBG); err != nil {
			return Style{}, fmt.Errorf("colors.panelBG: %w", err)
		}
		if err := mergeColor(&c.Button, raw.Colors.Button); err != nil {
			return Style{}, fmt.Errorf("colors.button: %w", err)
		}
		if err := mergeColor(&c.ButtonHover, raw.Colors.ButtonHover); err != nil {
			return Style{}, fmt.Errorf("colors.buttonHover: %w", err)
		}
		if err := mergeColor(&c.ButtonFocus, raw.Colors.ButtonFocus); err != nil {
			return Style{}, fmt.Errorf("colors.buttonFocus: %w", err)
		}
		if err := mergeColor(&c.SliderThumb, raw.Colors.SliderThumb); err != nil {
			return Style{}, fmt.Errorf("colors.sliderThumb: %w", err)
		}
		if err := mergeColor(&c.SliderThumbHover, raw.Colors.SliderThumbHover); err != nil {
			return Style{}, fmt.Errorf("colors.sliderThumbHover: %w", err)
		}
		if err := mergeColor(&c.SliderThumbFocus, raw.Colors.SliderThumbFocus); err != nil {
			return Style{}, fmt.Errorf("colors.sliderThumbFocus: %w", err)
		}
		if err := mergeColor(&c.Base, raw.Colors.Base); err != nil {
			return Style{}, fmt.Errorf("colors.base: %w", err)
		}
		if err := mergeColor(&c.BaseHover, raw.Colors.BaseHover); err != nil {
			return Style{}, fmt.Errorf("colors.baseHover: %w", err)
		}
		if err := mergeColor(&c.BaseFocus, raw.Colors.BaseFocus); err != nil {
			return Style{}, fmt.Errorf("colors.baseFocus: %w", err)
		}
		if err := mergeColor(&c.ScrollBase, raw.Colors.ScrollBase); err != nil {
			return Style{}, fmt.Errorf("colors.scrollBase: %w", err)
		}
		if err := mergeColor(&c.ScrollThumb, raw.Colors.ScrollThumb); err != nil {
			return Style{}, fmt.Errorf("colors.scrollThumb: %w", err)
		}
	}
	return s, nil
}

func mergeColor(dst *color.RGBA, src *string) error {
	if src == nil {
		return nil
	}
	c, err := ParseHexColor(*src)
	if err != nil {
		return err
	}
	*dst = c
	return nil
}

// ParseHexColor parses s as a CSS-style hex color: "#RGB", "#RRGGBB", or "#RRGGBBAA" (optional leading "#").
func ParseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	if s == "" {
		return color.RGBA{}, fmt.Errorf("debugui: empty color")
	}
	var r, g, b, a uint32 = 0, 0, 0, 255
	switch len(s) {
	case 3:
		for i := 0; i < 3; i++ {
			v, err := hexNibble(s[i])
			if err != nil {
				return color.RGBA{}, err
			}
			vv := v*16 + v
			switch i {
			case 0:
				r = uint32(vv)
			case 1:
				g = uint32(vv)
			case 2:
				b = uint32(vv)
			}
		}
	case 6:
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("debugui: parse hex color: %w", err)
		}
		r = uint32((v >> 16) & 0xff)
		g = uint32((v >> 8) & 0xff)
		b = uint32(v & 0xff)
	case 8:
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("debugui: parse hex color: %w", err)
		}
		r = uint32((v >> 24) & 0xff)
		g = uint32((v >> 16) & 0xff)
		b = uint32((v >> 8) & 0xff)
		a = uint32(v & 0xff)
	default:
		return color.RGBA{}, fmt.Errorf("debugui: hex color %q: want 3, 6, or 8 hex digits", s)
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}, nil
}

func hexNibble(b byte) (byte, error) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, nil
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, nil
	default:
		return 0, fmt.Errorf("debugui: invalid hex digit %q", b)
	}
}
