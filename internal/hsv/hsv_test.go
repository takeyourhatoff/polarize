package hsv

import (
	"image/color"
	"testing"
	"testing/quick"
)

func TestHSVModel(t *testing.T) {
	f := func(c1 color.RGBA) bool {
		c1.A = 0xff
		hsv := HSVModel.Convert(c1)
		c2 := color.RGBAModel.Convert(hsv).(color.RGBA)
		t.Logf("c1 = %#v, c2 = %#v\n", c1, c2)
		return c1 == c2
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
