// Package watermark provides video branding and watermark overlay filter graph generation.
package watermark

import (
	"fmt"
)

// Position defines the placement of a watermark overlay on video frames.
type Position string

const (
	// PositionTopRight places watermark at top-right corner.
	PositionTopRight Position = "top-right"
	// PositionTopLeft places watermark at top-left corner.
	PositionTopLeft Position = "top-left"
	// PositionBottomRight places watermark at bottom-right corner.
	PositionBottomRight Position = "bottom-right"
	// PositionBottomLeft places watermark at bottom-left corner.
	PositionBottomLeft Position = "bottom-left"
	// PositionCenter places watermark in the center.
	PositionCenter Position = "center"
)

// Config specifies watermark overlay parameters.
type Config struct {
	// Path is the filesystem path to the watermark image (PNG, WebP, JPEG).
	Path string
	// Position is the screen corner/anchor (default: PositionTopRight).
	Position Position
	// OffsetX is horizontal padding in pixels from the edge (default: 20).
	OffsetX int
	// OffsetY is vertical padding in pixels from the edge (default: 20).
	OffsetY int
	// Opacity is transparency factor from 0.0 (invisible) to 1.0 (fully opaque, default: 1.0).
	Opacity float64
	// ScaleFraction scales watermark width relative to rendition width (e.g. 0.15 = 15% width, default: 0.15).
	ScaleFraction float64
}

// Normalize validates and applies sensible defaults.
func (c *Config) Normalize() {
	if c.Position == "" {
		c.Position = PositionTopRight
	}
	if c.OffsetX <= 0 {
		c.OffsetX = 20
	}
	if c.OffsetY <= 0 {
		c.OffsetY = 20
	}
	if c.Opacity <= 0 || c.Opacity > 1.0 {
		c.Opacity = 1.0
	}
	if c.ScaleFraction <= 0 || c.ScaleFraction > 1.0 {
		c.ScaleFraction = 0.15
	}
}

// OverlayExpression returns the FFmpeg overlay filter expression coordinates based on Position.
func (c *Config) OverlayExpression() string {
	c.Normalize()
	switch c.Position {
	case PositionTopLeft:
		return fmt.Sprintf("x=%d:y=%d", c.OffsetX, c.OffsetY)
	case PositionBottomRight:
		return fmt.Sprintf("x=main_w-overlay_w-%d:y=main_h-overlay_h-%d", c.OffsetX, c.OffsetY)
	case PositionBottomLeft:
		return fmt.Sprintf("x=%d:y=main_h-overlay_h-%d", c.OffsetX, c.OffsetY)
	case PositionCenter:
		return "x=(main_w-overlay_w)/2:y=(main_h-overlay_h)/2"
	case PositionTopRight:
		fallthrough
	default:
		return fmt.Sprintf("x=main_w-overlay_w-%d:y=%d", c.OffsetX, c.OffsetY)
	}
}
