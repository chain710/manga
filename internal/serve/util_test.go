package serve

import (
	"github.com/stretchr/testify/require"
	"image"
	"testing"
)

func Test_parseRect(t *testing.T) {
	tests := []struct {
		name        string
		r           string
		expect      image.Rectangle
		expectError bool
	}{
		{
			name:   "normal",
			r:      "0000003e83e8",
			expect: image.Rect(0, 0, 1000, 1000),
		},
		{
			name:   "left:200,top:250,w:100,h:140",
			r:      "0c80fa06408c",
			expect: image.Rect(200, 250, 300, 390),
		},
		{
			name:        "left < 0",
			r:           "-c80fa06408c",
			expectError: true,
		},
		{
			name:        "left:200,top:250,w:100,h:751 height exceed",
			r:           "0c80fa0642ef",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseRect(tt.r)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, actual)
			}
		})
	}
}

func Test_bestFit(t *testing.T) {
	tests := []struct {
		name              string
		x, y, fitX, fitY  int
		expectX, expecctY int
	}{
		{
			name:     "same aspect",
			x:        100,
			y:        100,
			fitX:     10,
			fitY:     10,
			expectX:  10,
			expecctY: 10,
		},
		{
			name:     "same aspect, but small",
			x:        10,
			y:        10,
			fitX:     100,
			fitY:     100,
			expectX:  10,
			expecctY: 10,
		},
		{
			name:     "> aspect",
			x:        200,
			y:        100,
			fitX:     120,
			fitY:     120,
			expectX:  120,
			expecctY: 60,
		},
		{
			name:     "< aspect",
			x:        100,
			y:        200,
			fitX:     120,
			fitY:     120,
			expectX:  60,
			expecctY: 120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ax, ay := bestFit(tt.x, tt.y, tt.fitX, tt.fitY)
			require.Equal(t, tt.expectX, ax)
			require.Equal(t, tt.expecctY, ay)
		})
	}
}

func Test_sanitizeTextSearchQuery(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "name",
			input:  "q&a  adachi",
			expect: "q or a or adachi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sanitizeTextSearchQuery(tt.input)
			require.Equal(t, tt.expect, actual)
		})
	}
}
