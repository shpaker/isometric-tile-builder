// Package ebitendebug draws isometric tile geometry on Ebitengine images.
package ebitendebug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	itb "github.com/shpaker/isometric-tile-builder"
)

// Style describes the fill and outline of one face. A nil color omits that
// part of the drawing. OutlineWidth is measured in pixels, with zero using 1.
type Style struct {
	Fill, Outline color.Color
	OutlineWidth  float32
}

// StyleFunc chooses a style for a tile and side. Parts of the same side
// receive the same tile and side arguments.
type StyleFunc func(tile itb.FrameTile, side itb.Side) Style

// DefaultStyle gives each side a distinct shade and a one-pixel outline.
func DefaultStyle(_ itb.FrameTile, side itb.Side) Style {
	style := Style{Outline: color.RGBA{22, 30, 40, 255}}
	switch side {
	case itb.Left:
		style.Fill = color.RGBA{87, 108, 126, 255}
	case itb.Right:
		style.Fill = color.RGBA{65, 83, 103, 255}
	default:
		style.Fill = color.RGBA{136, 162, 180, 255}
	}
	return style
}

// Draw fills and outlines the frame in drawing order. A nil style uses
// DefaultStyle. dst and the frame use the same screen coordinate system.
func Draw(dst *ebiten.Image, frame itb.Frame, style StyleFunc) {
	if style == nil {
		style = DefaultStyle
	}
	for _, tile := range frame.Tiles() {
		for _, face := range tile.Faces {
			paint := style(tile, face.Side)
			var path vector.Path
			first := face.Polygon[0]
			path.MoveTo(float32(first.X), float32(first.Y))
			for _, point := range face.Polygon[1:] {
				path.LineTo(float32(point.X), float32(point.Y))
			}
			path.Close()

			if paint.Fill != nil {
				options := &vector.DrawPathOptions{AntiAlias: true}
				options.ColorScale.ScaleWithColor(paint.Fill)
				vector.FillPath(dst, &path, nil, options)
			}
			if paint.Outline != nil {
				width := paint.OutlineWidth
				if width <= 0 {
					width = 1
				}
				options := &vector.DrawPathOptions{AntiAlias: true}
				options.ColorScale.ScaleWithColor(paint.Outline)
				vector.StrokePath(dst, &path, &vector.StrokeOptions{
					Width: width, LineJoin: vector.LineJoinRound,
				}, options)
			}
		}
	}
}
