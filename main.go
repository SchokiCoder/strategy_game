// SPDX-License-Identifier: GPL-2.0-only
// Copyright (C) 2025  Andy Frank Schoknecht

//go:generate stringer -type=FieldBiome
package main

import (
	"embed"
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type FieldUnit int
const (
	None FieldUnit = iota

	Farm  // generates 3 food
	House // increases pop cap by 4

	Tower    // has sight and passively attacks
	Barracks // increases the units you can build per round

	Halberdier // good against paladin
	Knight     // good against halberdier ?
	Paladin    // faster, good against buildings ?
)

type FieldBiome int
const (
	Sea FieldBiome = iota // can't be controlled
	Plains                // no effect when controlled
	Forest                // generates 1 wood, 1 food
	Ores                  // generates 2 gold

	FieldBiomeCount
)

type World struct {
	W      int
	H      int
	_biome []FieldBiome
	Biome [][]FieldBiome
	_team  []int
	Team  [][]int
	_unit  []FieldUnit
	Unit  [][]FieldUnit
}

func NewWorld(
	w, h int,
) World {
	var ret = World{
		W: w,
		H: h,
		_biome: make([]FieldBiome, w * h),
		Biome: make([][]FieldBiome, w),
		_team: make([]int, w * h),
		Team: make([][]int, w),
		_unit: make([]FieldUnit, w * h),
		Unit: make([][]FieldUnit, w),
	}

	for x := 0; x < w; x++ {
		ret.Biome[x] = ret._biome[x * h:h + (x * h)]
		ret.Team[x] = ret._team[x * h:h + (x * h)]
		ret.Unit[x] = ret._unit[x * h:h + (x * h)]
	}

	return ret
}

type StratGame struct {
	BiomeImg      [FieldBiomeCount]*ebiten.Image
	Scroll        bool
	ScrollOldX    int
	ScrollOldY    int
	ScrollOriginX int
	ScrollOriginY int
	ScrollX       int
	ScrollY       int
	TeamColor     []color.Color
	TerrImg       *ebiten.Image
	Tilesize      int
	World         World
	WorldImg      *ebiten.Image
	Zoom          float64
}

func NewStratGame(
) StratGame {
	var (
		i FieldBiome
		ret = StratGame{
			TeamColor: []color.Color{
				color.RGBA{
					R: 0xD4,
					G: 0xF4,
					B: 0xD4,
					A: 0xFF,
				},
				color.RGBA{
					R: 0xBD,
					G: 0x3B,
					B: 0x3B,
					A: 0xFF,
				},
				color.RGBA{
					R: 0x5B,
					G: 0x7A,
					B: 0x8C,
					A: 0xFF,
				},
			},
			Tilesize: 16,
			Zoom: 1.0,
		}
	)

	for i = Sea; i <= Plains; i++ {
		ret.BiomeImg[i] = ebiten.NewImage(1, 1)
	}
	for ; i < FieldBiomeCount; i++ {
		var err error
		ret.BiomeImg[i], _, err = ebitenutil.NewImageFromFileSystem(
			assets,
			"assets/" + i.String() + ".png")
		if err != nil {
			panic(err)
		}
	}

	return ret
}

func (g StratGame) Draw(
	screen *ebiten.Image,
) {
	var opt ebiten.DrawImageOptions

	for x := 0; x < g.World.W; x++ {
		for y := 0; y < g.World.H; y++ {
			if g.World.Biome[x][y] != Sea {
				g.TerrImg.Set(x, y,
					g.TeamColor[g.World.Team[x][y]])
			} else {
				g.TerrImg.Set(x, y,
					color.RGBA{R: 0, G:0, B:0, A:0})
			}
		}
	}
	opt.GeoM.Scale(float64(g.Tilesize), float64(g.Tilesize))
	g.WorldImg.DrawImage(g.TerrImg, &opt)

	for x := 0; x < g.World.W; x++ {
		vector.StrokeLine(
			g.WorldImg,
			float32(x * g.Tilesize), 0,
			float32(x * g.Tilesize), float32(g.World.H * g.Tilesize),
			1,
			color.RGBA{R: 0x69, G: 0x69, B: 0x69, A: 0xFF},
			false)
	}
	for y := 0; y < g.World.H; y++ {
		vector.StrokeLine(
			g.WorldImg,
			0, float32(y * g.Tilesize),
			float32(g.World.W * g.Tilesize), float32(y * g.Tilesize),
			1,
			color.RGBA{R: 0x69, G: 0x69, B: 0x69, A: 0xFF},
			false)
	}

	for x := 0; x < g.World.W; x++ {
		for y := 0; y < g.World.H; y++ {
			opt.GeoM.Reset()
			opt.GeoM.Translate(
				float64(x * g.Tilesize + g.Tilesize / 2 - 1),
				float64(y * g.Tilesize))
			g.WorldImg.DrawImage(g.BiomeImg[g.World.Biome[x][y]], &opt)
		}
	}

	opt.GeoM.Reset()
	opt.GeoM.Translate(float64(g.ScrollX), float64(g.ScrollY))
	opt.GeoM.Scale(g.Zoom, g.Zoom)
	screen.DrawImage(g.WorldImg, &opt)
}

func (g* StratGame) GenerateSkirmish(
) {
	const (
		w, h = 16, 16
	)

	g.World = NewWorld(w, h)
	g.TerrImg = ebiten.NewImage(w, h)
	g.WorldImg = ebiten.NewImage(
		g.Tilesize * g.World.W,
		g.Tilesize * g.World.H)

	for x := 1; x < w - 1; x++ {
		for y := 1; y < h - 1; y++ {
			g.World.Biome[x][y] = Plains
		}
	}

	g.World.Team[3][3] = 1
	g.World.Unit[3][3] = Knight

	g.World.Biome[3][5] = Ores

	g.World.Biome[4][3] = Forest
	g.World.Biome[5][3] = Forest
	g.World.Biome[4][2] = Forest

	g.World.Team[12][12] = 2
	g.World.Unit[12][12] = Knight

	g.World.Biome[10][12] = Ores

	g.World.Biome[12][11] = Forest
	g.World.Biome[12][10] = Forest
	g.World.Biome[13][11] = Forest
}

func (g StratGame) Layout(
	outsideWidth, outsideHeight int,
) (int, int) {
	return g.Tilesize * g.World.W, g.Tilesize * g.World.H
}

func (g *StratGame) Update(
) error {
	const (
		minDistanceForScroll = 5
		mwheelOffsetZoomMod = 0.25
	)
	const (
		maxZoom = mwheelOffsetZoomMod * 16
		minZoom = mwheelOffsetZoomMod * 4
	)

	var mX, mY = ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.ScrollOriginX = mX
		g.ScrollOriginY = mY
		g.ScrollOldX = g.ScrollX
		g.ScrollOldY = g.ScrollY
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		diffX := g.ScrollOriginX - mX
		diffY := g.ScrollOriginY - mY
		if diffX >= minDistanceForScroll ||
		   diffX <= minDistanceForScroll ||
		   diffY >= minDistanceForScroll ||
		   diffY <= minDistanceForScroll {
			g.Scroll = true
		}
	}

	if g.Scroll {
		g.ScrollX = g.ScrollOldX + int(float64(mX - g.ScrollOriginX) / g.Zoom)
		g.ScrollY = g.ScrollOldY + int(float64(mY - g.ScrollOriginY) / g.Zoom)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.Scroll = false
	}

	var wX, wY = ebiten.Wheel()

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		newZoom := g.Zoom + wY * mwheelOffsetZoomMod
		if newZoom < minZoom ||
		   newZoom > maxZoom {
		} else {
			viewW := float64(g.World.W) * float64(g.Tilesize) / g.Zoom
			viewH := float64(g.World.H) * float64(g.Tilesize) / g.Zoom
			newViewW := float64(g.World.W) * float64(g.Tilesize) / newZoom
			newViewH := float64(g.World.H) * float64(g.Tilesize) / newZoom
			scrollmodX := int((viewW - newViewW) * 0.5)
			scrollmodY := int((viewH - newViewH) * 0.5)

			g.ScrollX -= scrollmodX
			g.ScrollY -= scrollmodY
			g.Zoom = newZoom
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.ScrollX += int(wY * float64(g.Tilesize))
	} else {
		g.ScrollY += int(wY * float64(g.Tilesize))
		g.ScrollX += int(wX * float64(g.Tilesize))
	}

	if g.ScrollX < int(0.0 - float64(g.World.W) * float64(g.Tilesize) * (g.Zoom - 1.0)) {
		g.ScrollX = int(0.0 - float64(g.World.W) * float64(g.Tilesize) * (g.Zoom - 1.0))
	} else if g.ScrollX > 0 {
		g.ScrollX = 0
	}

	if g.ScrollY < int(0.0 - float64(g.World.H) * float64(g.Tilesize) * (g.Zoom - 1.0)) {
		g.ScrollY = int(0.0 - float64(g.World.H) * float64(g.Tilesize) * (g.Zoom - 1.0))
	} else if g.ScrollY > 0 {
		g.ScrollY = 0
	}

	return nil
}

var (
	AppName string
)

//go:embed assets
var assets embed.FS

func main(
) {
	var (
		g = NewStratGame()
	)

	ebiten.SetWindowTitle(AppName)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetTPS(30)

	g.GenerateSkirmish()
	ebiten.RunGame(&g)
}

