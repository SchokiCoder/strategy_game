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
	BiomeImg  [FieldBiomeCount]*ebiten.Image
	TeamColor []color.Color
	Tilesize  int
	World     World
	WorldImg  *ebiten.Image
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
				g.WorldImg.Set(x, y,
					g.TeamColor[g.World.Team[x][y]])
			} else {
				g.WorldImg.Set(x, y,
					color.RGBA{R: 0, G:0, B:0, A:0})
			}
		}
	}
	opt.GeoM.Scale(float64(g.Tilesize), float64(g.Tilesize))
	screen.DrawImage(g.WorldImg, &opt)

	for x := 0; x < g.World.W; x++ {
		vector.StrokeLine(
			screen,
			float32(x * g.Tilesize), 0,
			float32(x * g.Tilesize), float32(g.World.H * g.Tilesize),
			1,
			color.RGBA{R: 0x69, G: 0x69, B: 0x69, A: 0xFF},
			false)
	}
	for y := 0; y < g.World.H; y++ {
		vector.StrokeLine(
			screen,
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
			screen.DrawImage(g.BiomeImg[g.World.Biome[x][y]], &opt)
		}
	}
}

func (g* StratGame) GenerateSkirmish(
) {
	const (
		w, h = 16, 16
	)

	g.World = NewWorld(w, h)
	g.WorldImg = ebiten.NewImage(w, h)

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

