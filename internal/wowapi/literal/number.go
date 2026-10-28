package literal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// jsonFloat keeps Python's historical shortest float spelling, including 1.0 and
// -0.0, when encoding literal numbers.
type jsonFloat float64

func (v jsonFloat) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
		return nil, fmt.Errorf("non-finite JSON number")
	}
	style := byte('f')
	a := math.Abs(float64(v))
	if a != 0 && (a < 1e-4 || a >= 1e16) {
		style = 'e'
	}
	text := strconv.FormatFloat(float64(v), style, -1, 64)
	if style == 'f' && !strings.Contains(text, ".") {
		text += ".0"
	}
	return []byte(text), nil
}
