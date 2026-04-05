package debugui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aldernero/debugui"
)

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		in   string
		want [4]uint8
	}{
		{"#fff", [4]uint8{255, 255, 255, 255}},
		{"#FFFFFF", [4]uint8{255, 255, 255, 255}},
		{"#102030", [4]uint8{16, 32, 48, 255}},
		{"#10203040", [4]uint8{16, 32, 48, 64}},
		{"102030", [4]uint8{16, 32, 48, 255}},
	}
	for _, tc := range tests {
		c, err := debugui.ParseHexColor(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if c.R != tc.want[0] || c.G != tc.want[1] || c.B != tc.want[2] || c.A != tc.want[3] {
			t.Fatalf("%q: got %+v want %v", tc.in, c, tc.want)
		}
	}
}

func TestParseStyleJSONPartial(t *testing.T) {
	data := []byte(`{"thumbSize": 12, "colors": {"text": "#000000"}}`)
	s, err := debugui.ParseStyleJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if s.ThumbSize != 12 {
		t.Fatalf("ThumbSize: got %d", s.ThumbSize)
	}
	def := debugui.DefaultStyle()
	if s.DefaultWidth != def.DefaultWidth {
		t.Fatalf("expected default width preserved")
	}
	if s.Colors.Text.R != 0 || s.Colors.Text.G != 0 || s.Colors.Text.B != 0 {
		t.Fatalf("text color: %+v", s.Colors.Text)
	}
	if s.Colors.Border != def.Colors.Border {
		t.Fatalf("border should stay default")
	}
}

func TestBuiltInThemeMenu(t *testing.T) {
	menu := debugui.BuiltInThemeMenu()
	if len(menu) < 2 {
		t.Fatalf("menu too short: %d", len(menu))
	}
	for _, o := range menu {
		if _, err := debugui.BuiltInTheme(o.Key); err != nil {
			t.Fatalf("key %q: %v", o.Key, err)
		}
		if o.Label == "" {
			t.Fatalf("empty label for key %q", o.Key)
		}
	}
}

func TestBuiltInTheme(t *testing.T) {
	if _, err := debugui.BuiltInTheme("light"); err != nil {
		t.Fatal(err)
	}
	if _, err := debugui.BuiltInTheme("dark"); err != nil {
		t.Fatal(err)
	}
	if _, err := debugui.BuiltInTheme(""); err != nil {
		t.Fatal(err)
	}
	if _, err := debugui.BuiltInTheme("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestExampleThemeFile(t *testing.T) {
	path := filepath.Join("themes", "light-partial.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("theme file not in cwd:", err)
	}
	if _, err := debugui.ParseStyleJSON(b); err != nil {
		t.Fatal(err)
	}
}
