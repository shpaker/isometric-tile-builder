# Isometric Tile Builder

Isometric Tile Builder converts a rectangular tile grid into screen-space
polygons and picks the visible tile face at a screen point. The root Go
module uses the standard library. The separate `ebitendebug` module contains
an Ebitengine renderer and interactive example.

```sh
go get github.com/shpaker/isometric-tile-builder@v0.1.0
```

```go
import itb "github.com/shpaker/isometric-tile-builder"

func pick(x, y float64) (itb.FrameTile, itb.Face, bool, error) {
    shore := &itb.Tile{Elevation: 0, Thickness: 1}
    water := &itb.Tile{Elevation: 0, Thickness: 0.75}
    raised := &itb.Tile{Elevation: 1, Thickness: 1}
    scene, err := itb.NewScene([][]*itb.Tile{
        {shore, water, raised},
        {shore, water, nil},
    })
    if err != nil {
        return itb.FrameTile{}, itb.Face{}, false, err
    }

    view := itb.DefaultView()
    view.Pan = itb.Point{X: 300, Y: 200}
    frame := itb.Build(scene, view)
    tile, face, ok := frame.Pick(x, y)
    return tile, face, ok, nil
}
```

## Geometry and picking

- `NewScene` copies a non-empty rectangular grid indexed by `[y][x]`.
  Each cell holds one tile or a `nil` hole. An all-hole grid is valid.
- `Elevation` is the bottom of a tile. `Thickness` is positive and the top
  is `Elevation + Thickness`. Both are measured in levels and may be
  fractional. In the example, water and shore share a bottom, and the
  water surface is 0.25 levels below the shore. The raised tile spans
  levels 1 through 2.
- The default view uses a 96 × 32 pixel diamond and 32 pixels per level.
  The camera takes positive zoom and screen-space pan. At `(0, 0, 0)`
  the projection is the pan position. Normal zoom is 1.
- A tile exposes its top and the side toward each of its front neighbors.
  A side next to a hole or field edge extends to the tile's bottom.
  The part sharing a vertical interval with its neighbor is omitted. A
  neighbor can remove the entire side, trim it, or split it into two
  strips. The renderer handles additional screen-space occlusion by
  drawing in frame order.
- Draw `frame.Tiles()` in slice order and each tile's `Faces` in slice
  order. A tile's `Anchor` is the screen-space center of its top face.
  Store the returned tile slice if you plan to reuse a frame. Each call
  returns a defensive copy.
- Pass screen coordinates to `frame.Pick(x, y)`. It tests the same
  polygons in reverse drawing order. Shared edges belong to the last
  drawn face. A hit includes the source tile's XY and elevation, and the
  polygon for the specific face or side strip.

## Ebitengine demo

```sh
go run github.com/shpaker/isometric-tile-builder/ebitendebug/cmd/tile-demo@v0.1.0
```

Hover to highlight a face and inspect the tile. Use the arrow keys to pan
and the mouse wheel to zoom about the cursor. In your own renderer, call
`ebitendebug.Draw(dst, frame, style)`. The optional style function receives
a projected tile and a side and returns fill and outline colors.

## Development checks

Run the following both at the repository root and inside `ebitendebug/`.
The debug module requires a graphical desktop to launch its demo.

```sh
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

Distributed under the MIT license. See [LICENSE](LICENSE).
