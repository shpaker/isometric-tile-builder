package itb

import "slices"

// Point holds screen coordinates in pixels.
type Point struct {
	X, Y float64
}

// View defines the tile diamond size, the pixel height of a level, and the
// camera transform. Pan is in screen pixels and is applied after Zoom.
type View struct {
	TileWidth, TileHeight, LevelHeight float64
	Pan                                Point
	Zoom                               float64
}

// DefaultView returns a 96 by 32 pixel diamond and a 32 pixel vertical level.
func DefaultView() View {
	return View{TileWidth: 96, TileHeight: 32, LevelHeight: 32, Zoom: 1}
}

// Side identifies a face under the fixed isometric view.
type Side uint8

const (
	Top Side = iota
	Left
	Right
)

// Face is a visible portion of a tile face. The polygon vertices follow the
// boundary in screen space. Clipping can produce two faces with the same side.
type Face struct {
	Side    Side
	Polygon [4]Point
}

// FrameTile contains the source tile and its screen geometry. Anchor is the
// center of the tile's upper face. Faces are in drawing order.
type FrameTile struct {
	X, Y   int
	Tile   Tile
	Anchor Point
	Faces  []Face
}

// Frame contains projected geometry for a scene and view.
type Frame struct {
	tiles []FrameTile
}

// Build projects the scene in back-to-front drawing order. Zero or invalid
// dimensions and zoom use their DefaultView values. Each non-finite pan
// component is treated as zero.
func Build(scene *Scene, view View) Frame {
	if scene == nil || len(scene.cells) == 0 {
		return Frame{}
	}

	view = view.withDefaults()
	width, height := len(scene.cells[0]), len(scene.cells)
	var frame Frame
	for depth := 0; depth < width+height-1; depth++ {
		for y := max(0, depth-width+1); y <= min(depth, height-1); y++ {
			x := depth - y
			current := scene.cells[y][x]
			if current.present {
				frame.tiles = append(frame.tiles, view.projectTile(scene, x, y, current.tile))
			}
		}
	}
	return frame
}

// Tiles returns a copy of the tiles in back-to-front drawing order.
func (frame Frame) Tiles() []FrameTile {
	tiles := slices.Clone(frame.tiles)
	for index := range tiles {
		tiles[index].Faces = slices.Clone(tiles[index].Faces)
	}
	return tiles
}

// Pick returns the tile and face drawn last at the screen point. Polygon
// edges are included, so drawing order also resolves shared edges.
func (frame Frame) Pick(x, y float64) (FrameTile, Face, bool) {
	if !finite(x) || !finite(y) {
		return FrameTile{}, Face{}, false
	}

	for tileIndex := len(frame.tiles) - 1; tileIndex >= 0; tileIndex-- {
		tile := frame.tiles[tileIndex]
		for faceIndex := len(tile.Faces) - 1; faceIndex >= 0; faceIndex-- {
			face := tile.Faces[faceIndex]
			if contains(face.Polygon, Point{X: x, Y: y}) {
				tile.Faces = slices.Clone(tile.Faces)
				return tile, face, true
			}
		}
	}
	return FrameTile{}, Face{}, false
}

func (view View) withDefaults() View {
	defaults := DefaultView()
	if view.TileWidth <= 0 || !finite(view.TileWidth) {
		view.TileWidth = defaults.TileWidth
	}
	if view.TileHeight <= 0 || !finite(view.TileHeight) {
		view.TileHeight = defaults.TileHeight
	}
	if view.LevelHeight <= 0 || !finite(view.LevelHeight) {
		view.LevelHeight = defaults.LevelHeight
	}
	if view.Zoom <= 0 || !finite(view.Zoom) {
		view.Zoom = defaults.Zoom
	}
	if !finite(view.Pan.X) {
		view.Pan.X = 0
	}
	if !finite(view.Pan.Y) {
		view.Pan.Y = 0
	}
	return view
}

func (view View) projectTile(scene *Scene, x, y int, tile Tile) FrameTile {
	xPos, yPos, top := float64(x), float64(y), tile.top()
	frameTile := FrameTile{
		X: x, Y: y,
		Tile:   tile,
		Anchor: view.project(xPos, yPos, top),
		Faces:  make([]Face, 0, 3),
	}

	frameTile.Faces = view.appendSide(frameTile.Faces, tile, scene.at(x, y+1), Left,
		Point{xPos - 0.5, yPos + 0.5}, Point{xPos + 0.5, yPos + 0.5})
	frameTile.Faces = view.appendSide(frameTile.Faces, tile, scene.at(x+1, y), Right,
		Point{xPos + 0.5, yPos + 0.5}, Point{xPos + 0.5, yPos - 0.5})
	frameTile.Faces = append(frameTile.Faces, Face{
		Side: Top,
		Polygon: [4]Point{
			view.project(xPos-0.5, yPos-0.5, top),
			view.project(xPos+0.5, yPos-0.5, top),
			view.project(xPos+0.5, yPos+0.5, top),
			view.project(xPos-0.5, yPos+0.5, top),
		},
	})
	return frameTile
}

func (view View) appendSide(faces []Face, tile Tile, neighbor cell, side Side, start, end Point) []Face {
	bottom, top := tile.Elevation, tile.top()
	if !neighbor.present || neighbor.tile.top() <= bottom || neighbor.tile.Elevation >= top {
		return append(faces, view.sideFace(side, start, end, bottom, top))
	}
	if neighbor.tile.Elevation > bottom {
		faces = append(faces, view.sideFace(side, start, end, bottom, neighbor.tile.Elevation))
	}
	if neighbor.tile.top() < top {
		faces = append(faces, view.sideFace(side, start, end, neighbor.tile.top(), top))
	}
	return faces
}

func (view View) sideFace(side Side, start, end Point, bottom, top float64) Face {
	return Face{
		Side: side,
		Polygon: [4]Point{
			view.project(start.X, start.Y, top),
			view.project(end.X, end.Y, top),
			view.project(end.X, end.Y, bottom),
			view.project(start.X, start.Y, bottom),
		},
	}
}

func (view View) project(x, y, z float64) Point {
	return Point{
		X: ((x-y)*view.TileWidth/2)*view.Zoom + view.Pan.X,
		Y: ((x+y)*view.TileHeight/2-z*view.LevelHeight)*view.Zoom + view.Pan.Y,
	}
}

func contains(polygon [4]Point, point Point) bool {
	var positive, negative bool
	for index, start := range polygon {
		end := polygon[(index+1)%len(polygon)]
		cross := (end.X-start.X)*(point.Y-start.Y) - (end.Y-start.Y)*(point.X-start.X)
		positive = positive || cross > 0
		negative = negative || cross < 0
	}
	return positive != negative
}
