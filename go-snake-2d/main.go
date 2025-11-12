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

// ============================================================================
// GAME DESIGN & LOGIC FLOW OVERVIEW
// ============================================================================
//
// This is a classic Snake game built using the Ebiten game engine.
// 
// CORE GAME LOOP:
// 1. Update() - Called every frame (~60 FPS) to handle input and game logic
// 2. Draw() - Called every frame to render the current game state
// 3. Layout() - Defines the screen dimensions
//
// GAME MECHANICS:
// - The snake is represented as a slice of Points (coordinates)
// - The snake moves by adding a new head in the direction of movement
// - If the snake eats food, it grows (old tail stays); otherwise tail is removed
// - Game over occurs when snake hits walls or itself
// - Game speed is controlled independently from frame rate using time-based updates
//
// DATA FLOW:
// Input (WASD keys) → Update direction → Time check → Move snake → 
// Check collisions → Update snake/food → Draw everything
//
// ============================================================================

const (
	// gameSpeed controls how fast the snake moves (6 updates per second)
	gameSpeed = time.Second / 6

	// Screen dimensions in pixels
	screenWidth  = 640
	screenHeight = 480

	// gridSize defines the size of each cell in pixels
	// The game grid is screenWidth/gridSize by screenHeight/gridSize cells
	gridSize = 20
)

// Direction vectors - used to move the snake in 2D space
var (
	dirUp    = Point{x: 0, y: -1}  // Moving up decreases y
	dirDown  = Point{x: 0, y: 1}   // Moving down increases y
	dirRight = Point{x: 1, y: 0}   // Moving right increases x
	dirLeft  = Point{x: -1, y: 0}  // Moving left decreases x

	// Font source for rendering text (loaded from embedded fonts)
	mplusFaceSource *text.GoTextFaceSource
)

// Point represents a position on the game grid
// Note: These are grid coordinates, not pixel coordinates
// To convert to pixels, multiply by gridSize
type Point struct {
	x, y int
}

// Game holds all the state for our snake game
type Game struct {
	// snake is a slice where [0] is the head and [len-1] is the tail
	snake []Point

	// direction is the current movement direction (one of the dir* vectors)
	direction Point

	// lastUpdate tracks when we last moved the snake
	// This allows us to control game speed independent of frame rate
	lastUpdate time.Time

	// food is the current position of the food item
	food Point

	// gameOver flag determines if the game has ended
	gameOver bool
}

// Update is called every frame by Ebiten (~60 times per second)
// This is where we handle input and update game state
func (g *Game) Update() error {
	// GAME OVER STATE HANDLING
	// When game is over, we only check for restart input
	if g.gameOver {
		// Check if player wants to restart
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			// Reset the game to initial state
			g.resetGame()
			return nil
		}
		// If Escape is pressed, we could exit, but for now just stay in game over
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			// Game remains in game over state
			return nil
		}
		return nil
	}

	// INPUT HANDLING
	// We capture input BEFORE the time check so direction changes feel responsive
	// The snake will move in the new direction on the next update tick
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		// Only allow direction change if it's not the opposite direction
		// (prevents snake from reversing into itself)
		if g.direction != dirDown {
			g.direction = dirUp
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.direction != dirUp {
			g.direction = dirDown
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.direction != dirRight {
			g.direction = dirLeft
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		if g.direction != dirLeft {
			g.direction = dirRight
		}
	}

	// TIME-BASED UPDATE
	// Only update game logic at gameSpeed intervals, not every frame
	// This decouples game speed from render speed
	if time.Since(g.lastUpdate) < gameSpeed {
		return nil // Not enough time has passed, skip this update
	}

	// Update the timer for the next movement
	g.lastUpdate = time.Now()

	// CORE GAME LOGIC
	// Move the snake in the current direction
	g.updateSnake(&g.snake, g.direction)

	return nil
}

// updateSnake handles the core snake movement logic
// This is where the "snake grows when eating" mechanic is implemented
func (g *Game) updateSnake(snake *[]Point, direction Point) {
	// Calculate the new head position based on current direction
	head := (*snake)[0]
	newHead := Point{
		x: head.x + direction.x,
		y: head.y + direction.y,
	}

	// COLLISION DETECTION
	// Check if the new head position causes game over
	if g.isBadCollision(newHead, *snake) {
		g.gameOver = true
		return
	}

	// FOOD CONSUMPTION
	// If snake eats food, grow by keeping the tail
	if newHead == g.food {
		// Prepend new head, keep entire body (snake grows)
		*snake = append([]Point{newHead}, *snake...)
		g.spawnFood() // Spawn new food at random location
	} else {
		// NORMAL MOVEMENT
		// Prepend new head, remove tail (snake moves without growing)
		// This creates the illusion of movement
		*snake = append(
			[]Point{newHead},
			(*snake)[:len(*snake)-1]...,
		)
	}
}

// isBadCollision checks if a point causes game over
// Returns true if the point is:
// 1. Outside the game boundaries (wall collision)
// 2. Overlapping with the snake's body (self collision)
func (g *Game) isBadCollision(p Point, snake []Point) bool {
	// BOUNDARY CHECK
	// Check if point is outside the grid
	if p.x < 0 || p.y < 0 || p.x >= screenWidth/gridSize || p.y >= screenHeight/gridSize {
		return true
	}

	// SELF-COLLISION CHECK
	// Check if point overlaps with any part of the snake's body
	for _, sp := range snake {
		if sp == p {
			return true
		}
	}

	return false
}

// Draw renders the current game state to the screen
// Called every frame by Ebiten
func (g *Game) Draw(screen *ebiten.Image) {
	// DRAW SNAKE
	// Render each segment of the snake as a white square
	for _, p := range g.snake {
		vector.FillRect(screen,
			float32(p.x*gridSize), // Convert grid coords to pixels
			float32(p.y*gridSize),
			gridSize,
			gridSize,
			color.White,
			true,
		)
	}

	// DRAW FOOD
	// Render food as a red square
	vector.FillRect(screen,
		float32(g.food.x*gridSize),
		float32(g.food.y*gridSize),
		gridSize,
		gridSize,
		color.RGBA{255, 0, 0, 255}, // Red color (alpha was 0, fixed to 255)
		true,
	)

	// DRAW GAME OVER SCREEN
	if g.gameOver {
		// Create font face for game over text
		face := &text.GoTextFace{
			Source: mplusFaceSource,
			Size:   48,
		}

		// GAME OVER TEXT
		gameOverText := "Game Over!"
		w, h := text.Measure(gameOverText, face, face.Size)

		// Center the text on screen
		op := &text.DrawOptions{}
		op.GeoM.Translate(screenWidth/2-w/2, screenHeight/2-h/2)
		op.ColorScale.ScaleWithColor(color.White)

		text.Draw(screen, gameOverText, face, op)

		// RESTART INSTRUCTIONS
		instructionFace := &text.GoTextFace{
			Source: mplusFaceSource,
			Size:   24,
		}
		instructionText := "Press ENTER or SPACE to restart"
		iw, _ := text.Measure(instructionText, instructionFace, instructionFace.Size)

		instructionOp := &text.DrawOptions{}
		instructionOp.GeoM.Translate(screenWidth/2-iw/2, screenHeight/2+h)
		instructionOp.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})

		text.Draw(screen, instructionText, instructionFace, instructionOp)
	}
}

// Layout defines the screen size
// Called by Ebiten to determine the game's logical screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// spawnFood generates a new food position at a random grid location
// Note: This doesn't check if food spawns on the snake (could be improved)
func (g *Game) spawnFood() {
	g.food = Point{
		x: rand.IntN(screenWidth / gridSize),
		y: rand.IntN(screenHeight / gridSize),
	}
}

// resetGame resets all game state to initial conditions for a new game
func (g *Game) resetGame() {
	// Reset snake to starting position (center of screen, length 2)
	g.snake = []Point{
		{
			x: screenWidth / gridSize / 2,
			y: screenHeight / gridSize / 2,
		},
		{
			x: screenWidth/gridSize/2 - 1,
			y: screenHeight / gridSize / 2,
		},
	}

	// Reset direction to moving right
	g.direction = Point{x: 1, y: 0}

	// Clear game over flag
	g.gameOver = false

	// Reset last update time to prevent immediate movement
	g.lastUpdate = time.Now()

	// Spawn new food
	g.spawnFood()
}

// main is the entry point of the program
// Sets up the game and starts the game loop
func main() {
	// FONT INITIALIZATION
	// Load the embedded font for rendering text
	s, err := text.NewGoTextFaceSource(
		bytes.NewReader(fonts.MPlus1pRegular_ttf),
	)
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s

	// GAME INITIALIZATION
	// Create initial game state with snake in center
	g := &Game{
		snake: []Point{
			{
				x: screenWidth / gridSize / 2,
				y: screenHeight / gridSize / 2,
			},
			{
				x: screenWidth/gridSize/2 - 1,
				y: screenHeight/gridSize/2 - 1,
			},
		},
		direction:  Point{x: 1, y: 0}, // Start moving right
		lastUpdate: time.Now(),        // Initialize timer
	}

	// Spawn initial food
	g.spawnFood()

	// WINDOW SETUP
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake Game - WASD to move")

	// START GAME LOOP
	// This blocks until the game window is closed
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}