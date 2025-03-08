package theme

import (
	"fmt"
	"image/color"

	"github.com/crazy3lf/colorconv"
)

type Color struct {
	col color.Color
}

func (c Color) HSL() (h, s, l float64) {
	return colorconv.ColorToHSL(c.Color())
}
func (c Color) Color() color.Color {
	return c.col
}

func (c Color) String() string {
	r, g, b, _ := c.col.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func (c *Color) UnmarshalYAML(unmarshal func(any) error) error {
	var col string
	err := unmarshal(&col)
	if err != nil {
		return err
	}
	if len(col) == 4 {
		col = fmt.Sprintf("#%c%c%c%c%c%c", col[1], col[1], col[2], col[2], col[3], col[3])
	}

	c.col, err = colorconv.HexToColor(col)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid color: %w", col, err)
	}

	return nil
}
