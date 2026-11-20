package itb_test

import (
	"fmt"
	"testing"

	itb "github.com/shpaker/isometric-tile-builder"
)

func TestRaisedTilePicking(t *testing.T) {
	defaults := itb.DefaultView()
	for _, aspect := range []float64{2, 3, 4} {
		t.Run(fmt.Sprintf("%g-to-1", aspect), func(t *testing.T) {
			view := defaults
			view.TileWidth = view.TileHeight * aspect
			view.Pan = itb.Point{X: 100, Y: 180}
			view.Zoom = 1.25

			for _, test := range []struct {
				name      string
				elevation float64
				wantX     int
				wantSide  itb.Side
			}{
				{"raised side covers the far top", 1, 1, itb.Left},
				{"gap exposes the far top", 2, 0, itb.Top},
			} {
				t.Run(test.name, func(t *testing.T) {
					scene, err := itb.NewScene([][]*itb.Tile{
						{{Thickness: 1}, {Elevation: test.elevation, Thickness: 1}},
					})
					if err != nil {
						t.Fatal(err)
					}
					frame := itb.Build(scene, view)
					x := view.Pan.X + view.TileWidth*3/8*view.Zoom
					y := view.Pan.Y - view.LevelHeight*view.Zoom
					tile, face, ok := frame.Pick(x, y)
					if !ok || tile.X != test.wantX || tile.Y != 0 || face.Side != test.wantSide {
						t.Fatalf("Pick(%g, %g) = (%d, %d), side %v, hit %t; want (%d, 0), side %v",
							x, y, tile.X, tile.Y, face.Side, ok, test.wantX, test.wantSide)
					}

					raised := frame.Tiles()[1]
					tile, face, ok = frame.Pick(raised.Anchor.X, raised.Anchor.Y)
					if !ok || tile.X != 1 || tile.Y != 0 || tile.Tile.Elevation != test.elevation || face.Side != itb.Top {
						t.Fatalf("Pick on raised anchor = %+v, side %v, hit %t", tile, face.Side, ok)
					}
				})
			}
		})
	}
}

func TestPickWithCustomProjection(t *testing.T) {
	scene, err := itb.NewScene([][]*itb.Tile{{{Elevation: 1, Thickness: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	view := itb.View{
		TileWidth: 100, TileHeight: 40, LevelHeight: 17,
		Pan: itb.Point{X: 180.25, Y: 120.5}, Zoom: 1.5,
	}
	frame := itb.Build(scene, view)
	tile := frame.Tiles()[0]
	if want := (itb.Point{X: 180.25, Y: 44}); tile.Anchor != want {
		t.Fatalf("top center = %+v, want %+v", tile.Anchor, want)
	}
	for _, test := range []struct {
		name string
		x, y float64
		side itb.Side
	}{
		{"top", 180.25, 44, itb.Top},
		{"left", 142.75, 84.5, itb.Left},
		{"right", 217.75, 84.5, itb.Right},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, face, ok := frame.Pick(test.x, test.y)
			if !ok || got.X != 0 || got.Y != 0 || got.Tile.Elevation != 1 || face.Side != test.side {
				t.Fatalf("Pick(%g, %g) = %+v, side %v, hit %t", test.x, test.y, got, face.Side, ok)
			}
		})
	}
}
