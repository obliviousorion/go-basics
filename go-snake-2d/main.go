package main

import (
	"image/color" // The standard Go package for defining colors.
	"log"         // Used for simple error logging.

	"github.com/hajimehoshi/ebiten/v2" // The core Ebitengine game library.
)

// Game is the main struct that holds the game state and implements Ebitengine's Game interface.
// Ebitengine requires any game object to have three methods: Update, Draw, and Layout.
type Game struct{}

// Update is called 60 times per second (60 FPS) and is where all game logic goes.
// This includes handling input, updating object positions, collision detection, etc.
// The primary job is to change the state of the game.
// Parameters:
//   - (g *Game): The receiver, meaning this method operates on the Game struct instance.
// Returns:
//   - error: Returns an error if the game loop should terminate (e.g., a critical failure).
func (g *Game) Update() error {
	// Currently, this function does nothing, so we just return nil (no error).
	return nil
}

// Draw is responsible for rendering the current state of the game to the screen.
// This is typically called less frequently than Update, depending on the monitor's refresh rate.
// NOTE: Only drawing operations should happen here, not game logic updates.
// Parameters:
//   - (g *Game): The receiver (the Game instance).
//   - screen (*ebiten.Image): The main canvas (the target image) where all drawing must occur.
func (g *Game) Draw(screen *ebiten.Image) {
	// screen.Fill is a utility method that fills the entire 'screen' canvas with a single color.
	// color.RGBA{R, G, B, A} defines a color using 8-bit Red, Green, Blue, and Alpha (transparency) values.
	// 0xff in hexadecimal is 255 in decimal.
	// color.RGBA{0xff, 0, 0, 0xff} sets the color to:
	// R: 255 (Full Red)
	// G: 0   (No Green)
	// B: 0   (No Blue)
	// A: 255 (Fully Opaque)
	screen.Fill(color.RGBA{0xff, 0, 0, 0xff}) // This fills the window with solid RED.
}

// Layout determines the virtual (logical) screen size of the game, independent of the window size.
// Ebitengine automatically scales the virtual screen to fit the physical window/monitor size.
// Parameters:
//   - (g *Game): The receiver (the Game instance).
//   - outsideWidth (int): The physical width of the window/monitor in pixels.
//   - outsideHeight (int): The physical height of the window/monitor in pixels.
// Returns:
//   - screenWidth (int): The logical game width. Coordinates (0, 0) to (screenWidth-1, screenHeight-1) are used for drawing.
//   - screenHeight (int): The logical game height.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// We set the logical game resolution to 320x240.
	// If the physical window is 640x480, the game will be scaled up 2x (320*2 = 640, 240*2 = 480).
	return 320, 240
}

// main is the entry point of the Go program.
func main() {
	// SetWindowSize sets the physical dimensions of the window that opens on the screen.
	// Parameters:
	//   - 640 (int): The width of the physical window in pixels.
	//   - 480 (int): The height of the physical window in pixels.
	ebiten.SetWindowSize(640, 480)

	// SetWindowTitle sets the text that appears in the window's title bar.
	ebiten.SetWindowTitle("fill")

	// RunGame starts the main game loop, passing in our custom Game struct implementation.
	// This function blocks until the game window is closed.
	// It returns an error if the game loop fails to start or encounters a critical runtime issue.
	if err := ebiten.RunGame(&Game{}); err != nil {
		// If an error is returned, log.Fatal prints the error message and terminates the program immediately.
		log.Fatal(err)
	}
}