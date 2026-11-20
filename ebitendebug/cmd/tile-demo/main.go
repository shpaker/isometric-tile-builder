package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	itb "github.com/shpaker/isometric-tile-builder"
	"github.com/shpaker/isometric-tile-builder/ebitendebug"
)

type game struct {
	scene *itb.Scene
	view  itb.View
	frame itb.Frame

	hoveredTile itb.FrameTile
	hoveredFace itb.Face
	hovered     bool
}

func (game *game) Update() error {
	const panSpeed = 5
	previousView := game.view
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		game.view.Pan.X -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		game.view.Pan.X += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		game.view.Pan.Y -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		game.view.Pan.Y += panSpeed
	}
	x, y := ebiten.CursorPosition()
	_, wheel := ebiten.Wheel()
	if wheel != 0 {
		zoom := min(3, max(0.5, game.view.Zoom*math.Pow(1.12, wheel)))
		ratio := zoom / game.view.Zoom
		game.view.Pan.X = float64(x) + (game.view.Pan.X-float64(x))*ratio
		game.view.Pan.Y = float64(y) + (game.view.Pan.Y-float64(y))*ratio
		game.view.Zoom = zoom
	}
	if game.view != previousView {
		game.frame = itb.Build(game.scene, game.view)
	}
	game.hoveredTile, game.hoveredFace, game.hovered = game.frame.Pick(float64(x), float64(y))
	return nil
}

func (game *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 23, 35, 255})
	ebitendebug.Draw(screen, game.frame, game.style)
	label := "Arrow keys: pan   Mouse wheel: zoom   Hover a tile face"
	if game.hovered {
		tile := game.hoveredTile
		names := [...]string{itb.Top: "top", itb.Left: "left", itb.Right: "right"}
		label += fmt.Sprintf("\nTile (%d, %d)  Base %g  Thickness %g  Face %s",
			tile.X, tile.Y, tile.Tile.Elevation, tile.Tile.Thickness, names[game.hoveredFace.Side])
	}
	ebitenutil.DebugPrint(screen, label)
}

func (game *game) style(tile itb.FrameTile, side itb.Side) ebitendebug.Style {
	style := ebitendebug.DefaultStyle(tile, side)
	if tile.Tile.Thickness < 1 {
		water := [...]color.RGBA{
			itb.Top: {62, 156, 214, 255}, itb.Left: {44, 114, 182, 255}, itb.Right: {37, 87, 153, 255},
		}
		style.Fill = water[side]
	}
	if game.hovered && tile.X == game.hoveredTile.X && tile.Y == game.hoveredTile.Y && side == game.hoveredFace.Side {
		style.Fill = color.RGBA{250, 194, 96, 255}
		style.Outline = color.RGBA{255, 231, 166, 255}
		style.OutlineWidth = 2
	}
	return style
}

func (*game) Layout(_, _ int) (int, int) {
	return 720, 480
}

func main() {
	ground := &itb.Tile{Thickness: 1}
	water := &itb.Tile{Thickness: 0.75}
	raised := &itb.Tile{Elevation: 1, Thickness: 1}
	scene, err := itb.NewScene([][]*itb.Tile{
		{ground, ground, ground, ground, nil},
		{ground, raised, raised, ground, ground},
		{ground, ground, water, water, ground},
		{ground, water, water, water, ground},
		{nil, ground, ground, ground, nil},
	})
	if err != nil {
		log.Fatal(err)
	}

	view := itb.DefaultView()
	view.Pan = itb.Point{X: 360, Y: 190}
	ebiten.SetWindowSize(720, 480)
	ebiten.SetWindowTitle("Isometric Tile Builder")
	if err := ebiten.RunGame(&game{scene: scene, view: view, frame: itb.Build(scene, view)}); err != nil {
		log.Fatal(err)
	}
}
