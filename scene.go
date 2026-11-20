// Package itb projects rectangular height fields into isometric tile geometry.
package itb

import (
	"fmt"
	"math"
)

// Tile describes a solid interval in units of one vertical level.
// Elevation is the bottom and Thickness is the distance to the top.
type Tile struct {
	Elevation float64
	Thickness float64
}

type cell struct {
	tile    Tile
	present bool
}

// Scene is an immutable rectangular grid with at most one tile at each (x, y).
type Scene struct {
	cells [][]cell
}

// NewScene copies a non-empty rectangular grid indexed by [y][x]. A nil tile
// leaves the cell empty. Tile elevations must be finite and thicknesses must
// be finite, positive, and large enough to produce a distinct upper elevation.
func NewScene(cells [][]*Tile) (*Scene, error) {
	if len(cells) == 0 || len(cells[0]) == 0 {
		return nil, fmt.Errorf("scene must contain at least one row and column")
	}

	width := len(cells[0])
	scene := &Scene{cells: make([][]cell, len(cells))}
	for y, row := range cells {
		if len(row) != width {
			return nil, fmt.Errorf("row %d has %d cells, expected %d", y, len(row), width)
		}

		scene.cells[y] = make([]cell, width)
		for x, tile := range row {
			if tile == nil {
				continue
			}
			top := tile.top()
			if !finite(tile.Elevation) || !finite(tile.Thickness) || !finite(top) || top <= tile.Elevation {
				return nil, fmt.Errorf("tile (%d, %d) must have finite elevation and positive representable thickness", x, y)
			}
			scene.cells[y][x] = cell{tile: *tile, present: true}
		}
	}
	return scene, nil
}

func (tile Tile) top() float64 {
	return tile.Elevation + tile.Thickness
}

func (scene *Scene) at(x, y int) cell {
	if y >= len(scene.cells) || x >= len(scene.cells[0]) {
		return cell{}
	}
	return scene.cells[y][x]
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
