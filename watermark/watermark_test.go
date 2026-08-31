package watermark

import (
	"testing"
)

func TestConfigNormalize(t *testing.T) {
	cfg := Config{}
	cfg.Normalize()

	if cfg.Position != PositionTopRight {
		t.Errorf("expected PositionTopRight, got %s", cfg.Position)
	}
	if cfg.OffsetX != 20 {
		t.Errorf("expected OffsetX 20, got %d", cfg.OffsetX)
	}
	if cfg.OffsetY != 20 {
		t.Errorf("expected OffsetY 20, got %d", cfg.OffsetY)
	}
	if cfg.Opacity != 1.0 {
		t.Errorf("expected Opacity 1.0, got %f", cfg.Opacity)
	}
	if cfg.ScaleFraction != 0.15 {
		t.Errorf("expected ScaleFraction 0.15, got %f", cfg.ScaleFraction)
	}
}

func TestOverlayExpression(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		cfg      Config
	}{
		{
			name:     "top-right default",
			cfg:      Config{Position: PositionTopRight, OffsetX: 15, OffsetY: 15},
			expected: "x=main_w-overlay_w-15:y=15",
		},
		{
			name:     "top-left",
			cfg:      Config{Position: PositionTopLeft, OffsetX: 10, OffsetY: 10},
			expected: "x=10:y=10",
		},
		{
			name:     "bottom-right",
			cfg:      Config{Position: PositionBottomRight, OffsetX: 25, OffsetY: 30},
			expected: "x=main_w-overlay_w-25:y=main_h-overlay_h-30",
		},
		{
			name:     "bottom-left",
			cfg:      Config{Position: PositionBottomLeft, OffsetX: 12, OffsetY: 18},
			expected: "x=12:y=main_h-overlay_h-18",
		},
		{
			name:     "center",
			cfg:      Config{Position: PositionCenter},
			expected: "x=(main_w-overlay_w)/2:y=(main_h-overlay_h)/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := tt.cfg.OverlayExpression()
			if expr != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr)
			}
		})
	}
}
