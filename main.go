// SPDX-License-Identifier: GPL-2.0-only
// Copyright (C) 2025  Andy Frank Schoknecht

//go:generate stringer -type=FieldBiome
package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	_ "image/png"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
	Font          text.GoTextFace
	FoodImg       *ebiten.Image
	GoldImg       *ebiten.Image
	PopCapImg     *ebiten.Image
	Scroll        bool
	ScrollOldX    float64
	ScrollOldY    float64
	ScrollOriginX float64
	ScrollOriginY float64
	ScrollX       float64
	ScrollY       float64
	TeamColor     []color.Color
	TeamFood      []int
	TeamGold      []int
	TeamPopCap    []int
	TeamWood      []int
	TerrImg       *ebiten.Image
	Tilesize      float64
	WoodImg       *ebiten.Image
	World         World
	WorldImg      *ebiten.Image
	WorldImgW     float64
	WorldImgH     float64
	Zoom          float64
}

func NewStratGame(
) StratGame {
	var (
		i   FieldBiome
		err error
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
			TeamFood: []int{
				0,
				InitFood,
				InitFood,
			},
			TeamGold: []int{
				0,
				InitGold,
				InitGold,
			},
			TeamPopCap: []int{
				0,
				InitPopCap,
				InitPopCap,
			},
			TeamWood: []int{
				0,
				InitWood,
				InitWood,
			},
			Tilesize: 16,
			Zoom: 1.0,
		}
	)

	ret.Font = text.GoTextFace{
		Source: dejavusansmonoSource,
		Direction: text.DirectionLeftToRight,
		Size: ret.Tilesize - 2,
		Language: language.English,
	}

	ret.FoodImg, _, err = ebitenutil.NewImageFromFileSystem(
		images,
		"assets/Food.png")
	if err != nil {
		panic(err)
	}

	ret.GoldImg, _, err = ebitenutil.NewImageFromFileSystem(
		images,
		"assets/Gold.png")
	if err != nil {
		panic(err)
	}

	ret.PopCapImg, _, err = ebitenutil.NewImageFromFileSystem(
		images,
		"assets/PopCap.png")
	if err != nil {
		panic(err)
	}

	ret.WoodImg, _, err = ebitenutil.NewImageFromFileSystem(
		images,
		"assets/Wood.png")
	if err != nil {
		panic(err)
	}

	for i = Sea; i <= Plains; i++ {
		ret.BiomeImg[i] = ebiten.NewImage(1, 1)
	}
	for ; i < FieldBiomeCount; i++ {
		ret.BiomeImg[i], _, err = ebitenutil.NewImageFromFileSystem(
			images,
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
	var (
		edio ebiten.DrawImageOptions
		tdo  text.DrawOptions
	)

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
	edio.GeoM.Scale(g.Tilesize, g.Tilesize)
	g.WorldImg.DrawImage(g.TerrImg, &edio)

	for x := 0; x < g.World.W; x++ {
		vector.StrokeLine(
			g.WorldImg,
			float32(float64(x) * g.Tilesize),
			0,
			float32(float64(x) * g.Tilesize),
			float32(g.WorldImgH),
			1,
			color.RGBA{R: 0x69, G: 0x69, B: 0x69, A: 0xFF},
			false)
	}
	for y := 0; y < g.World.H; y++ {
		vector.StrokeLine(
			g.WorldImg,
			0,
			float32(float64(y) * g.Tilesize),
			float32(g.WorldImgW),
			float32(float64(y) * g.Tilesize),
			1,
			color.RGBA{R: 0x69, G: 0x69, B: 0x69, A: 0xFF},
			false)
	}

	for x := 0; x < g.World.W; x++ {
		for y := 0; y < g.World.H; y++ {
			edio.GeoM.Reset()
			edio.GeoM.Translate(
				float64(x) * g.Tilesize + g.Tilesize / 2.0 - 1.0,
				float64(y) * g.Tilesize)
			g.WorldImg.DrawImage(g.BiomeImg[g.World.Biome[x][y]], &edio)
		}
	}

	edio.GeoM.Reset()
	edio.GeoM.Translate(g.ScrollX, g.ScrollY)
	edio.GeoM.Scale(g.Zoom, g.Zoom)
	edio.GeoM.Translate(0, g.Tilesize)
	screen.DrawImage(g.WorldImg, &edio)

	vector.DrawFilledRect(
		screen,
		0,
		0,
		float32(g.WorldImgW),
		float32(g.Tilesize),
		color.RGBA{0x11, 0x11, 0x11, 0xFF},
		true)

	drawCursor := 0.0
	headerImgs := [...]*ebiten.Image{
		g.PopCapImg,
		g.FoodImg,
		g.GoldImg,
		g.WoodImg,
	}
	headerTxts := [...]string{
		fmt.Sprintf("%v/%v", 0, g.TeamPopCap[1]),
		fmt.Sprintf("%v", g.TeamFood[1]),
		fmt.Sprintf("%v", g.TeamGold[1]),
		fmt.Sprintf("%v", g.TeamWood[1]),
	}

	for i := 0; i < len(headerImgs); i++ {
		edio.GeoM.Reset()
		edio.GeoM.Translate(drawCursor, 0)
		screen.DrawImage(headerImgs[i], &edio)
		drawCursor += g.Tilesize + 1

		tdo.GeoM.Reset()
		tdo.GeoM.Translate(drawCursor, 1)
		text.Draw(screen, headerTxts[i], &g.Font, &tdo)
		tw, _ := text.Measure(headerTxts[i], &g.Font, 0)
		drawCursor += tw + 10
	}
}

func (g* StratGame) GenerateSkirmish(
) {
	const (
		w, h = 16, 16
	)

	g.World = NewWorld(w, h)
	g.TerrImg = ebiten.NewImage(w, h)
	g.WorldImgW = g.Tilesize * float64(g.World.W)
	g.WorldImgH = g.Tilesize * float64(g.World.H)
	g.WorldImg = ebiten.NewImage(int(g.WorldImgW), int(g.WorldImgH))

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
	_, _ int,
) (int, int) {
	return -1, -1
}

func (g StratGame) LayoutF(
	outsideWidth, outsideHeight float64,
) (float64, float64) {
	return g.WorldImgW, g.WorldImgH + g.Tilesize
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

	_mX, _mY := ebiten.CursorPosition()
	mX, mY := float64(_mX), float64(_mY)

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
		g.ScrollX = g.ScrollOldX + (mX - g.ScrollOriginX) / g.Zoom
		g.ScrollY = g.ScrollOldY + (mY - g.ScrollOriginY) / g.Zoom
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
			viewW := g.WorldImgW / g.Zoom
			viewH := g.WorldImgH / g.Zoom
			newViewW := g.WorldImgW / newZoom
			newViewH := g.WorldImgH / newZoom
			scrollmodX := (viewW - newViewW) * 0.5
			scrollmodY := (viewH - newViewH) * 0.5

			g.ScrollX -= scrollmodX
			g.ScrollY -= scrollmodY
			g.Zoom = newZoom
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.ScrollX += wY * g.Tilesize
	} else {
		g.ScrollY += wY * g.Tilesize
		g.ScrollX += wX * g.Tilesize
	}

	scrollcapX := g.WorldImgW / g.Zoom - g.WorldImgW
	if g.ScrollX < scrollcapX {
		g.ScrollX = scrollcapX
	} else if g.ScrollX > 0 {
		g.ScrollX = 0
	}

	scrollcapY := g.WorldImgH / g.Zoom - g.WorldImgH
	if g.ScrollY < scrollcapY {
		g.ScrollY = scrollcapY
	} else if g.ScrollY > 0 {
		g.ScrollY = 0
	}

	return nil
}

const (
	InitFood = 3
	InitGold = 3
	InitWood = 3
	InitPopCap = 5
)

var (
	AppName string
)

//go:embed assets/DejaVuSansMono.ttf
var dejavusansmonoTTF []byte

var dejavusansmonoSource *text.GoTextFaceSource

func init(
) {
	s, err := text.NewGoTextFaceSource(
		bytes.NewReader(dejavusansmonoTTF))
	if err != nil {
		panic(err)
	}
	dejavusansmonoSource = s
}

//go:embed assets/*.png
var images embed.FS

func main(
) {
	var (
		g = NewStratGame()
	)

	ebiten.SetWindowTitle(AppName)
	ebiten.SetWindowSize(512, 512 + int(g.Tilesize * 2))
	ebiten.SetTPS(30)

	g.GenerateSkirmish()
	ebiten.RunGame(&g)
}
