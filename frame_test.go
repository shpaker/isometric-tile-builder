package itb

import (
	"math"
	"reflect"
	"testing"
)

func mustScene(t *testing.T, cells [][]*Tile) *Scene {
	t.Helper()
	scene, err := NewScene(cells)
	if err != nil {
		t.Fatal(err)
	}
	return scene
}

func TestDefaultView(t *testing.T) {
	view := DefaultView()
	if view.TileWidth != 3*view.TileHeight || view.LevelHeight <= 0 || view.Zoom != 1 {
		t.Fatalf("unexpected defaults: %+v", view)
	}
	scene := mustScene(t, [][]*Tile{{{Thickness: 1}}})
	if got, want := Build(scene, View{}), Build(scene, view); !reflect.DeepEqual(got, want) {
		t.Fatalf("Build with a zero View = %+v, want %+v", got, want)
	}
}

func TestSceneValidationAndCopy(t *testing.T) {
	for name, cells := range map[string][][]*Tile{
		"empty":       nil,
		"empty row":   {{}},
		"ragged":      {{nil}, {nil, nil}},
		"zero height": {{{Elevation: 1}}},
		"negative":    {{{Thickness: -1}}},
		"infinite":    {{{Elevation: math.Inf(1), Thickness: 1}}},
		"NaN":         {{{Thickness: math.NaN()}}},
		"overflow":    {{{Elevation: math.MaxFloat64, Thickness: math.MaxFloat64}}},
	} {
		t.Run(name, func(t *testing.T) {
			if scene, err := NewScene(cells); err == nil {
				t.Fatalf("NewScene(%v) = %v, want error", cells, scene)
			}
		})
	}

	cells := [][]*Tile{{{Elevation: -1, Thickness: 2}, nil}}
	scene := mustScene(t, cells)
	cells[0][0].Thickness = 4
	cells[0][1] = &Tile{Thickness: 1}
	frame := Build(scene, View{})
	tiles := frame.Tiles()
	if len(tiles) != 1 || tiles[0].Tile.Thickness != 2 {
		t.Fatalf("unexpected tiles after modifying input: %+v", tiles)
	}
	tiles[0].Faces[0] = Face{}
	if frame.Tiles()[0].Faces[0] == (Face{}) {
		t.Fatal("modifying output changed the frame")
	}
}

func TestTileOrderAndAnchor(t *testing.T) {
	scene := mustScene(t, [][]*Tile{
		{{Thickness: 1}, {Thickness: 0.75}, {Elevation: 1, Thickness: 1}},
		{{Thickness: 1}, nil, {Thickness: 1}},
	})
	view := View{TileWidth: 60, TileHeight: 20, LevelHeight: 10, Zoom: 2, Pan: Point{100, 200}}
	tiles := Build(scene, view).Tiles()
	wantCells := [][2]int{{0, 0}, {1, 0}, {0, 1}, {2, 0}, {2, 1}}
	for index, tile := range tiles {
		if got := [2]int{tile.X, tile.Y}; got != wantCells[index] {
			t.Fatalf("tile %d at %v, want %v", index, got, wantCells[index])
		}
	}
	if len(tiles) != len(wantCells) {
		t.Fatalf("got %d tiles, want %d", len(tiles), len(wantCells))
	}
	if got, want := tiles[3].Anchor, (Point{220, 200}); got != want {
		t.Fatalf("raised tile anchor = %+v, want %+v", got, want)
	}
}

func TestSideClipping(t *testing.T) {
	cases := []struct {
		name     string
		tile     Tile
		neighbor *Tile
		want     [][2]float64
	}{
		{"hole", Tile{0, 1}, nil, [][2]float64{{0, 1}}},
		{"water", Tile{0, 1}, &Tile{0, 0.75}, [][2]float64{{0.75, 1}}},
		{"neighbor covers wall", Tile{1, 1}, &Tile{0, 3}, nil},
		{"neighbor covers middle", Tile{0, 3}, &Tile{1, 1}, [][2]float64{{0, 1}, {2, 3}}},
		{"neighbor covers top", Tile{0, 2}, &Tile{1, 3}, [][2]float64{{0, 1}}},
		{"neighbor reaches base", Tile{1, 1}, &Tile{0, 1}, [][2]float64{{1, 2}}},
		{"neighbor above", Tile{0, 1}, &Tile{2, 1}, [][2]float64{{0, 1}}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			scene := mustScene(t, [][]*Tile{{&test.tile, test.neighbor}, {test.neighbor, nil}})
			view := DefaultView()
			tile := Build(scene, view).Tiles()[0]
			for _, side := range []Side{Left, Right} {
				var got [][2]float64
				for _, face := range tile.Faces {
					if face.Side != side {
						continue
					}
					// Both front faces share the south vertex of tile (0, 0).
					south := 1
					if side == Right {
						south = 0
					}
					lower := 3 - south
					got = append(got, [2]float64{
						(view.TileHeight/2 - face.Polygon[lower].Y) / view.LevelHeight,
						(view.TileHeight/2 - face.Polygon[south].Y) / view.LevelHeight,
					})
				}
				if !reflect.DeepEqual(got, test.want) {
					t.Fatalf("side %d spans %v, want %v", side, got, test.want)
				}
			}
		})
	}
}

func TestPickDrawOrder(t *testing.T) {
	scene := mustScene(t, [][]*Tile{{{Thickness: 1}, {Thickness: 0.75}}})
	for _, view := range []View{
		DefaultView(),
		{TileWidth: 96, TileHeight: 32, LevelHeight: 32, Zoom: 1.5, Pan: Point{52, -22}},
	} {
		frame := Build(scene, view)
		for _, test := range []struct {
			point Point
			x     int
			side  Side
		}{
			{Point{0, -32}, 0, Top},
			{Point{-24, -1}, 0, Left},
			{Point{24, -20}, 0, Right},
			{Point{30, -10}, 1, Top},
			{Point{24, -16}, 1, Top},
			{Point{-24, -24}, 0, Top},
		} {
			point := Point{test.point.X*view.Zoom + view.Pan.X, test.point.Y*view.Zoom + view.Pan.Y}
			tile, face, ok := frame.Pick(point.X, point.Y)
			if !ok || tile.X != test.x || tile.Y != 0 || face.Side != test.side {
				t.Fatalf("Pick(%+v) = tile (%d, %d), face %d, found %t, want tile (%d, 0), face %d",
					point, tile.X, tile.Y, face.Side, ok, test.x, test.side)
			}
		}
		if _, _, ok := frame.Pick(1000, 1000); ok {
			t.Fatal("Pick returned a tile outside the frame")
		}
		if _, _, ok := frame.Pick(math.NaN(), 0); ok {
			t.Fatal("Pick returned a tile for NaN")
		}
	}
}
