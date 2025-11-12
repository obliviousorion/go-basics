package main

import (
	"bytes"
	"image/color"
	"log"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	gameSpeed    = time.Second / 6
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 20
)

var (
	dirUp           = Point{x: 0, y: -1}
	dirDown         = Point{x: 0, y: 1}
	dirRight        = Point{x: 1, y: 0}
	dirLeft         = Point{x: -1, y: 0}
	mplusFaceSource *text.GoTextFaceSource
)

type Point struct {
	x, y int
}

type Game struct {
	snake      []Point
	direction  Point
	lastUpdate time.Time
	food       Point
	gameOver   bool
}

func (g *Game) Update() error {

	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.gameOver = true
			return nil
		}

		return nil
	}

	// we need to capture the keystroke before everything or else our game might be
	// buggy
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.direction = dirUp
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.direction = dirDown
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.direction = dirLeft
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.direction = dirRight
	}

	if time.Since(g.lastUpdate) < gameSpeed {
		return nil
	}
	g.lastUpdate = time.Now()

	g.updateSnake(&g.snake, g.direction)
	return nil
}

func (g *Game) updateSnake(
	snake *[]Point, direction Point) {
	head := (*snake)[0]
	newHead := Point{
		x: head.x + direction.x,
		y: head.y + direction.y,
	}

	if g.isBadCollision(newHead, *snake) {
		g.gameOver = true
		return
	}

	if newHead == g.food {

		*snake = append([]Point{newHead}, *snake...,
		)
		g.spawnFood()
	} else {
		*snake = append(
			[]Point{newHead},
			(*snake)[:len(*snake)-1]...,
		)
	}

}

func (g *Game) isBadCollision(
	p Point, 
	snake []Point,
) bool {
	if p.x < 0 || p.y < 0 || p.x >= screenWidth/gridSize || p.y >= screenHeight/gridSize{
		return true
	}

	for _, sp := range snake {
		if sp == p {
			return true
		}
	}

	return false
}


func (g *Game) Draw(screen *ebiten.Image) {
	for _, p := range g.snake {
		vector.FillRect(screen,
			float32(p.x*gridSize),
			float32(p.y*gridSize),
			gridSize,
			gridSize,
			color.White,
			true,
		)
	}

	vector.FillRect(screen,
		float32(g.food.x*gridSize),
		float32(g.food.y*gridSize),
		gridSize,
		gridSize,
		color.RGBA{255, 0, 0, 0},
		true,
	)

	if g.gameOver {
		face := &text.GoTextFace{
			Source: mplusFaceSource,
			Size:   48,
		}

		w,h := text.Measure("Game Over!",
	face, 
	face.Size,
	)

		op := &text.DrawOptions{}
		op.GeoM.Translate(screenWidth/2 - w/2, screenHeight/2 - h/2)
		op.ColorScale.ScaleWithColor(color.White)

		text.Draw(
				screen,
				"Game Over",
				face,
				op,
			)
	}
}

func (g *Game) Layout(
	outsideWidth,
	outsideHeight int,
) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) spawnFood() {
	g.food = Point{rand.IntN(screenWidth / gridSize),
		rand.IntN(screenHeight / gridSize),
	}
}

func main() {

	s, err := text.NewGoTextFaceSource(
		bytes.NewReader(
			fonts.MPlus1pRegular_ttf,
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s

	g := &Game{
		snake: []Point{{
			x: screenWidth / gridSize / 2,
			y: screenHeight / gridSize / 2,
		},
			{
				x: screenWidth/gridSize/2 - 1,
				y: screenHeight/gridSize/2 - 1,
			},
		},
		direction: Point{x: 1, y: 0},
	}

	g.spawnFood()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("snake")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
