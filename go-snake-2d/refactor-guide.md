# 🐍 Complete Snake Game Refactoring Guide

### Visit this for some more info https://www.youtube.com/watch?v=ZpHqL5ykCyw&list=PL10piHcP2kVKr0Uokt215JmhgGEgw1dE-&index=1


## 📚 Table of Contents
1. [Current Problems](#current-problems)
2. [Refactoring Strategy](#refactoring-strategy)
3. [Step-by-Step Implementation](#step-by-step-implementation)
4. [Final Architecture](#final-architecture)
5. [Testing Strategy](#testing-strategy)

---

## 🔴 Current Problems

### Problem 1: Mixed Responsibilities
```go
func (g *Game) Update() error {
    // Handles input ❌
    if ebiten.IsKeyPressed(ebiten.KeyW) { ... }
    
    // Handles timing ❌
    if time.Since(g.lastUpdate) < gameSpeed { ... }
    
    // Handles game logic ❌
    g.updateSnake(&g.snake, g.direction)
}
```
**Issue**: One function does THREE different things. Hard to test and maintain.

### Problem 2: Global Variables
```go
var mplusFaceSource *text.GoTextFaceSource
```
**Issue**: Makes testing difficult, creates hidden dependencies.

### Problem 3: Magic Numbers
```go
if p.x < 0 || p.y < 0 || p.x >= screenWidth/gridSize ...
```
**Issue**: Hard to change game size, unclear what numbers mean.

### Problem 4: Poor Testability
No interfaces = can't mock dependencies = hard to unit test.

### Problem 5: Food Can Spawn on Snake
```go
func (g *Game) spawnFood() {
    g.food = Point{rand.IntN(screenWidth / gridSize), ...}
}
```
**Issue**: No validation if food spawns on snake body.

---

## 🎯 Refactoring Strategy

### Philosophy
**"Make it work, make it right, make it fast"** - Kent Beck

We'll refactor in phases:
1. ✅ Extract configuration
2. ✅ Create domain types
3. ✅ Separate systems (components)
4. ✅ Add interfaces
5. ✅ Improve error handling
6. ✅ Add tests

### Design Principles We'll Follow
1. **Single Responsibility** - One function = one job
2. **Dependency Injection** - Pass what you need
3. **Interface Segregation** - Small, focused interfaces
4. **Don't Repeat Yourself (DRY)**
5. **Composition over Inheritance**

---

## 📁 Step 1: New Project Structure

Create this file organization:

```
snake-game/
├── main.go              # Entry point
├── game/
│   ├── game.go          # Game orchestration
│   ├── config.go        # Configuration
│   └── state.go         # Game state enum
├── domain/
│   ├── point.go         # Point type and methods
│   ├── direction.go     # Direction type and constants
│   └── snake.go         # Snake type with methods
├── systems/
│   ├── input.go         # Input handling
│   ├── movement.go      # Snake movement logic
│   ├── collision.go     # Collision detection
│   └── food.go          # Food spawning
├── rendering/
│   └── renderer.go      # All drawing code
└── tests/
    ├── collision_test.go
    ├── movement_test.go
    └── food_test.go
```

**Why this structure?**
- Clear separation of concerns
- Easy to find specific functionality
- Logical grouping by responsibility
- Supports testing

---

## 📝 Step 2: Extract Configuration

### File: `game/config.go`

```go
package game

import "time"

// Config holds all game configuration
// Makes it easy to adjust game parameters without touching code
type Config struct {
    // Display settings
    ScreenWidth  int
    ScreenHeight int
    GridSize     int
    WindowTitle  string
    
    // Game mechanics
    GameSpeed          time.Duration
    InitialSnakeLength int
    
    // Colors (could be loaded from config file)
    SnakeColor color.Color
    FoodColor  color.Color
    BGColor    color.Color
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
    return Config{
        ScreenWidth:        640,
        ScreenHeight:       480,
        GridSize:           20,
        WindowTitle:        "Snake Game",
        GameSpeed:          time.Second / 6,
        InitialSnakeLength: 2,
        SnakeColor:         color.White,
        FoodColor:          color.RGBA{255, 0, 0, 255},
        BGColor:            color.Black,
    }
}

// GridWidth returns the number of grid cells horizontally
func (c Config) GridWidth() int {
    return c.ScreenWidth / c.GridSize
}

// GridHeight returns the number of grid cells vertically
func (c Config) GridHeight() int {
    return c.ScreenHeight / c.GridSize
}

// IsInBounds checks if a point is within grid boundaries
func (c Config) IsInBounds(x, y int) bool {
    return x >= 0 && y >= 0 && x < c.GridWidth() && y < c.GridHeight()
}
```

**Benefits:**
- ✅ All configuration in one place
- ✅ Easy to load from JSON/YAML later
- ✅ Helper methods for common calculations
- ✅ Type-safe and documented

---

## 📝 Step 3: Create Domain Types

### File: `domain/point.go`

```go
package domain

// Point represents a grid position
type Point struct {
    X, Y int
}

// Add returns a new Point by adding another Point
func (p Point) Add(other Point) Point {
    return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Equals checks if two points are the same
func (p Point) Equals(other Point) bool {
    return p.X == other.X && p.Y == other.Y
}
```

### File: `domain/direction.go`

```go
package domain

// Direction represents a movement direction
type Direction struct {
    DX, DY int
}

// Predefined directions as constants
var (
    DirUp    = Direction{DX: 0, DY: -1}
    DirDown  = Direction{DX: 0, DY: 1}
    DirLeft  = Direction{DX: -1, DY: 0}
    DirRight = Direction{DX: 1, DY: 0}
)

// IsOpposite checks if this direction is opposite to another
func (d Direction) IsOpposite(other Direction) bool {
    return d.DX == -other.DX && d.DY == -other.DY
}

// ToPoint converts direction to a Point (for addition)
func (d Direction) ToPoint() Point {
    return Point{X: d.DX, Y: d.DY}
}
```

### File: `domain/snake.go`

```go
package domain

// Snake represents the player's snake
type Snake struct {
    segments  []Point
    direction Direction
}

// NewSnake creates a snake at the given starting position
func NewSnake(start Point, length int, direction Direction) *Snake {
    segments := make([]Point, length)
    segments[0] = start
    
    // Create body segments going backwards from head
    for i := 1; i < length; i++ {
        segments[i] = Point{
            X: start.X - (i * direction.DX),
            Y: start.Y - (i * direction.DY),
        }
    }
    
    return &Snake{
        segments:  segments,
        direction: direction,
    }
}

// Head returns the head position
func (s *Snake) Head() Point {
    return s.segments[0]
}

// Body returns all segments except the head
func (s *Snake) Body() []Point {
    if len(s.segments) <= 1 {
        return []Point{}
    }
    return s.segments[1:]
}

// AllSegments returns a copy of all segments
func (s *Snake) AllSegments() []Point {
    result := make([]Point, len(s.segments))
    copy(result, s.segments)
    return result
}

// Length returns the number of segments
func (s *Snake) Length() int {
    return len(s.segments)
}

// Direction returns the current direction
func (s *Snake) Direction() Direction {
    return s.direction
}

// SetDirection changes the snake's direction
// Only allows changes that aren't opposite to current direction
func (s *Snake) SetDirection(newDir Direction) {
    if !s.direction.IsOpposite(newDir) {
        s.direction = newDir
    }
}

// Move moves the snake in its current direction
// If grow is true, the tail stays (snake grows)
// If grow is false, the tail is removed (normal movement)
func (s *Snake) Move(grow bool) {
    newHead := s.Head().Add(s.direction.ToPoint())
    
    if grow {
        // Prepend new head, keep tail
        s.segments = append([]Point{newHead}, s.segments...)
    } else {
        // Prepend new head, remove tail
        s.segments = append([]Point{newHead}, s.segments[:len(s.segments)-1]...)
    }
}

// Contains checks if the snake contains a specific point
func (s *Snake) Contains(p Point) bool {
    for _, segment := range s.segments {
        if segment.Equals(p) {
            return true
        }
    }
    return false
}
```

**Why create these types?**
- ✅ Encapsulates behavior with data
- ✅ Provides clear API
- ✅ Easy to test in isolation
- ✅ Self-documenting code

---

## 📝 Step 4: Create System Components

### File: `systems/input.go`

```go
package systems

import (
    "github.com/hajimehoshi/ebiten/v2"
    "yourproject/domain"
)

// InputSystem handles keyboard input
type InputSystem struct{}

// NewInputSystem creates a new input system
func NewInputSystem() *InputSystem {
    return &InputSystem{}
}

// ReadDirection reads the current directional input
// Returns the new direction and true if direction changed
func (is *InputSystem) ReadDirection(currentDir domain.Direction) (domain.Direction, bool) {
    if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
        if !currentDir.IsOpposite(domain.DirUp) {
            return domain.DirUp, true
        }
    }
    if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
        if !currentDir.IsOpposite(domain.DirDown) {
            return domain.DirDown, true
        }
    }
    if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
        if !currentDir.IsOpposite(domain.DirLeft) {
            return domain.DirLeft, true
        }
    }
    if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
        if !currentDir.IsOpposite(domain.DirRight) {
            return domain.DirRight, true
        }
    }
    
    return currentDir, false
}

// IsRestartPressed checks if restart key is pressed
func (is *InputSystem) IsRestartPressed() bool {
    return ebiten.IsKeyPressed(ebiten.KeyEnter) || 
           ebiten.IsKeyPressed(ebiten.KeySpace)
}

// IsPausePressed checks if pause key is pressed
func (is *InputSystem) IsPausePressed() bool {
    return ebiten.IsKeyPressed(ebiten.KeyP)
}
```

### File: `systems/collision.go`

```go
package systems

import (
    "yourproject/domain"
    "yourproject/game"
)

// CollisionSystem handles all collision detection
type CollisionSystem struct {
    config game.Config
}

// NewCollisionSystem creates a collision system with config
func NewCollisionSystem(config game.Config) *CollisionSystem {
    return &CollisionSystem{config: config}
}

// CheckWallCollision checks if point hits a wall
func (cs *CollisionSystem) CheckWallCollision(p domain.Point) bool {
    return !cs.config.IsInBounds(p.X, p.Y)
}

// CheckSelfCollision checks if point collides with any segment
func (cs *CollisionSystem) CheckSelfCollision(p domain.Point, segments []domain.Point) bool {
    for _, segment := range segments {
        if p.Equals(segment) {
            return true
        }
    }
    return false
}

// CheckFoodCollision checks if snake head is on food
func (cs *CollisionSystem) CheckFoodCollision(head, food domain.Point) bool {
    return head.Equals(food)
}

// CheckAnyCollision checks all collision types
// Returns true if ANY collision occurred
func (cs *CollisionSystem) CheckAnyCollision(point domain.Point, snake *domain.Snake) bool {
    if cs.CheckWallCollision(point) {
        return true
    }
    if cs.CheckSelfCollision(point, snake.AllSegments()) {
        return true
    }
    return false
}
```

### File: `systems/food.go`

```go
package systems

import (
    "errors"
    "math/rand/v2"
    "yourproject/domain"
    "yourproject/game"
)

// FoodSystem handles food spawning
type FoodSystem struct {
    config game.Config
}

// NewFoodSystem creates a new food system
func NewFoodSystem(config game.Config) *FoodSystem {
    return &FoodSystem{config: config}
}

// SpawnFood spawns food at a random valid location
// Returns error if unable to find valid spot after max attempts
func (fs *FoodSystem) SpawnFood(snake *domain.Snake) (domain.Point, error) {
    maxAttempts := 100
    
    for i := 0; i < maxAttempts; i++ {
        candidate := domain.Point{
            X: rand.IntN(fs.config.GridWidth()),
            Y: rand.IntN(fs.config.GridHeight()),
        }
        
        // Check if this point is not on the snake
        if !snake.Contains(candidate) {
            return candidate, nil
        }
    }
    
    // If snake is too long and fills the board
    return domain.Point{}, errors.New("unable to spawn food: board full")
}
```

### File: `systems/movement.go`

```go
package systems

import (
    "yourproject/domain"
)

// MovementSystem handles snake movement
type MovementSystem struct {
    collisionSystem *CollisionSystem
}

// NewMovementSystem creates a movement system
func NewMovementSystem(collisionSystem *CollisionSystem) *MovementSystem {
    return &MovementSystem{
        collisionSystem: collisionSystem,
    }
}

// MoveResult represents the result of a move attempt
type MoveResult struct {
    Success      bool
    AteFood      bool
    GameOver     bool
    GameOverType string // "wall" or "self"
}

// TryMove attempts to move the snake
func (ms *MovementSystem) TryMove(snake *domain.Snake, food domain.Point) MoveResult {
    // Calculate where head will be
    nextHead := snake.Head().Add(snake.Direction().ToPoint())
    
    // Check wall collision
    if ms.collisionSystem.CheckWallCollision(nextHead) {
        return MoveResult{
            Success:      false,
            GameOver:     true,
            GameOverType: "wall",
        }
    }
    
    // Check self collision
    if ms.collisionSystem.CheckSelfCollision(nextHead, snake.AllSegments()) {
        return MoveResult{
            Success:      false,
            GameOver:     true,
            GameOverType: "self",
        }
    }
    
    // Check food collision
    ateFood := ms.collisionSystem.CheckFoodCollision(nextHead, food)
    
    // Move the snake (grow if ate food)
    snake.Move(ateFood)
    
    return MoveResult{
        Success: true,
        AteFood: ateFood,
        GameOver: false,
    }
}
```

**Why create systems?**
- ✅ Single responsibility per system
- ✅ Easy to test (can mock dependencies)
- ✅ Reusable across different game modes
- ✅ Clear interfaces

---

## 📝 Step 5: Create Renderer

### File: `rendering/renderer.go`

```go
package rendering

import (
    "fmt"
    "image/color"
    
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    
    "yourproject/domain"
    "yourproject/game"
)

// Renderer handles all drawing operations
type Renderer struct {
    config     game.Config
    fontSource *text.GoTextFaceSource
}

// NewRenderer creates a new renderer with dependencies
func NewRenderer(config game.Config, fontSource *text.GoTextFaceSource) *Renderer {
    return &Renderer{
        config:     config,
        fontSource: fontSource,
    }
}

// DrawSnake renders the snake
func (r *Renderer) DrawSnake(screen *ebiten.Image, snake *domain.Snake) {
    for _, segment := range snake.AllSegments() {
        r.drawGridCell(screen, segment, r.config.SnakeColor)
    }
}

// DrawFood renders the food
func (r *Renderer) DrawFood(screen *ebiten.Image, food domain.Point) {
    r.drawGridCell(screen, food, r.config.FoodColor)
}

// DrawScore renders the current score
func (r *Renderer) DrawScore(screen *ebiten.Image, score int) {
    face := &text.GoTextFace{
        Source: r.fontSource,
        Size:   24,
    }
    
    scoreText := fmt.Sprintf("Score: %d", score)
    op := &text.DrawOptions{}
    op.GeoM.Translate(10, 10)
    op.ColorScale.ScaleWithColor(color.White)
    
    text.Draw(screen, scoreText, face, op)
}

// DrawGameOver renders the game over screen
func (r *Renderer) DrawGameOver(screen *ebiten.Image, score int, reason string) {
    // Semi-transparent overlay
    r.drawOverlay(screen)
    
    // "Game Over" text
    r.drawCenteredText(screen, "Game Over!", 48, r.config.ScreenHeight/2-60)
    
    // Reason
    reasonText := fmt.Sprintf("Cause: %s", reason)
    r.drawCenteredText(screen, reasonText, 24, r.config.ScreenHeight/2-10)
    
    // Score
    scoreText := fmt.Sprintf("Final Score: %d", score)
    r.drawCenteredText(screen, scoreText, 32, r.config.ScreenHeight/2+30)
    
    // Instructions
    r.drawCenteredText(screen, "Press ENTER to restart", 20, r.config.ScreenHeight/2+80)
}

// drawGridCell draws a single grid cell
func (r *Renderer) drawGridCell(screen *ebiten.Image, p domain.Point, clr color.Color) {
    vector.FillRect(
        screen,
        float32(p.X*r.config.GridSize),
        float32(p.Y*r.config.GridSize),
        float32(r.config.GridSize),
        float32(r.config.GridSize),
        clr,
        true,
    )
}

// drawOverlay draws a semi-transparent overlay
func (r *Renderer) drawOverlay(screen *ebiten.Image) {
    vector.FillRect(
        screen,
        0, 0,
        float32(r.config.ScreenWidth),
        float32(r.config.ScreenHeight),
        color.RGBA{0, 0, 0, 180},
        true,
    )
}

// drawCenteredText draws text centered horizontally
func (r *Renderer) drawCenteredText(screen *ebiten.Image, txt string, size float64, y float64) {
    face := &text.GoTextFace{
        Source: r.fontSource,
        Size:   size,
    }
    
    w, _ := text.Measure(txt, face, face.Size)
    
    op := &text.DrawOptions{}
    op.GeoM.Translate(float64(r.config.ScreenWidth)/2-w/2, y)
    op.ColorScale.ScaleWithColor(color.White)
    
    text.Draw(screen, txt, face, op)
}
```

**Benefits:**
- ✅ All rendering logic in one place
- ✅ No rendering code in game logic
- ✅ Easy to swap renderers (e.g., for testing)
- ✅ Reusable drawing utilities

---

## 📝 Step 6: Game State Management

### File: `game/state.go`

```go
package game

// State represents the current game state
type State int

const (
    StateMenu State = iota
    StatePlaying
    StatePaused
    StateGameOver
)

// String returns the string representation of the state
func (s State) String() string {
    switch s {
    case StateMenu:
        return "Menu"
    case StatePlaying:
        return "Playing"
    case StatePaused:
        return "Paused"
    case StateGameOver:
        return "Game Over"
    default:
        return "Unknown"
    }
}
```

---

## 📝 Step 7: Orchestrate Everything in Game

### File: `game/game.go`

```go
package game

import (
    "time"
    
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    
    "yourproject/domain"
    "yourproject/rendering"
    "yourproject/systems"
)

// Game orchestrates all systems and manages game state
type Game struct {
    // Configuration
    config Config
    
    // Game state
    state        State
    snake        *domain.Snake
    food         domain.Point
    score        int
    lastUpdate   time.Time
    gameOverType string
    
    // Systems (injected dependencies)
    inputSystem     *systems.InputSystem
    movementSystem  *systems.MovementSystem
    collisionSystem *systems.CollisionSystem
    foodSystem      *systems.FoodSystem
    renderer        *rendering.Renderer
}

// NewGame creates a new game with all dependencies
func NewGame(config Config, fontSource *text.GoTextFaceSource) *Game {
    // Create systems
    collisionSystem := systems.NewCollisionSystem(config)
    movementSystem := systems.NewMovementSystem(collisionSystem)
    foodSystem := systems.NewFoodSystem(config)
    inputSystem := systems.NewInputSystem()
    renderer := rendering.NewRenderer(config, fontSource)
    
    // Create game
    g := &Game{
        config:          config,
        state:           StatePlaying,
        collisionSystem: collisionSystem,
        movementSystem:  movementSystem,
        foodSystem:      foodSystem,
        inputSystem:     inputSystem,
        renderer:        renderer,
        lastUpdate:      time.Now(),
    }
    
    // Initialize game state
    g.reset()
    
    return g
}

// Update is called every frame
func (g *Game) Update() error {
    switch g.state {
    case StatePlaying:
        return g.updatePlaying()
    case StateGameOver:
        return g.updateGameOver()
    case StatePaused:
        return g.updatePaused()
    default:
        return nil
    }
}

// updatePlaying handles the playing state
func (g *Game) updatePlaying() error {
    // Handle input
    if newDir, changed := g.inputSystem.ReadDirection(g.snake.Direction()); changed {
        g.snake.SetDirection(newDir)
    }
    
    // Check for pause
    if g.inputSystem.IsPausePressed() {
        g.state = StatePaused
        return nil
    }
    
    // Time-based update
    if time.Since(g.lastUpdate) < g.config.GameSpeed {
        return nil
    }
    g.lastUpdate = time.Now()
    
    // Move snake
    result := g.movementSystem.TryMove(g.snake, g.food)
    
    if result.GameOver {
        g.state = StateGameOver
        g.gameOverType = result.GameOverType
        return nil
    }
    
    if result.AteFood {
        g.score += 10
        newFood, err := g.foodSystem.SpawnFood(g.snake)
        if err != nil {
            // Board is full - player wins!
            g.state = StateGameOver
            g.gameOverType = "victory"
        } else {
            g.food = newFood
        }
    }
    
    return nil
}

// updateGameOver handles the game over state
func (g *Game) updateGameOver() error {
    if g.inputSystem.IsRestartPressed() {
        g.reset()
        g.state = StatePlaying
    }
    return nil
}

// updatePaused handles the paused state
func (g *Game) updatePaused() error {
    if g.inputSystem.IsPausePressed() {
        g.state = StatePlaying
    }
    return nil
}

// Draw renders the current game state
func (g *Game) Draw(screen *ebiten.Image) {
    // Always draw game elements
    g.renderer.DrawSnake(screen, g.snake)
    g.renderer.DrawFood(screen, g.food)
    g.renderer.DrawScore(screen, g.score)
    
    // Draw state-specific UI
    if g.state == StateGameOver {
        g.renderer.DrawGameOver(screen, g.score, g.gameOverType)
    }
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return g.config.ScreenWidth, g.config.ScreenHeight
}

// reset resets the game to initial state
func (g *Game) reset() {
    // Create new snake in center
    centerX := g.config.GridWidth() / 2
    centerY := g.config.GridHeight() / 2
    start := domain.Point{X: centerX, Y: centerY}
    
    g.snake = domain.NewSnake(start, g.config.InitialSnakeLength, domain.DirRight)
    g.score = 0
    g.lastUpdate = time.Now()
    g.gameOverType = ""
    
    // Spawn initial food
    food, _ := g.foodSystem.SpawnFood(g.snake)
    g.food = food
}
```

**Benefits:**
- ✅ Clean orchestration of systems
- ✅ Clear state machine
- ✅ All dependencies injected
- ✅ Easy to test (mock systems)

---

## 📝 Step 8: Clean Main Function

### File: `main.go`

```go
package main

import (
    "bytes"
    "log"
    
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    
    "yourproject/game"
)

func main() {
    // Load font
    fontSource, err := text.NewGoTextFaceSource(
        bytes.NewReader(fonts.MPlus1pRegular_ttf),
    )
    if err != nil {
        log.Fatalf("Failed to load font: %v", err)
    }
    
    // Create configuration
    config := game.DefaultConfig()
    
    // Create game with dependencies
    g := game.NewGame(config, fontSource)
    
    // Setup window
    ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
    ebiten.SetWindowTitle(config.WindowTitle)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    
    // Run game
    if err := ebiten.RunGame(g); err != nil {
        log.Fatalf("Game error: %v", err)
    }
}
```

**Benefits:**
- ✅ Minimal main function
- ✅ Clear dependency setup
- ✅ Easy to understand flow
- ✅ Error handling

---

## 🧪 Step 9: Add Tests

### File: `tests/collision_test.go`

```go
package tests

import (
    "testing"
    "yourproject/domain"
    "yourproject/game"
    "yourproject/systems"
)

func TestCollisionSystem_WallCollision(t *testing.T) {
    config := game.DefaultConfig()
    cs := systems.NewCollisionSystem(config)
    
    tests := []struct {
        name     string
        point    domain.Point
        expected bool
    }{
        {"Inside bounds", domain.Point{X: 5, Y: 5}, false},
        {"At origin", domain.Point{X: 0, Y: 0}, false},
        {"Negative X", domain.Point{X: -1, Y: 5}, true},
        {"Negative Y", domain.Point{X: 5, Y: -1}, true},
        {"Beyond width", domain.Point{X: 100, Y: 5}, true},
        {"Beyond height", domain.Point{X: 5, Y: 100}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := cs.CheckWallCollision(tt.point)
            if result != tt.expected {
                t.Errorf("Expected %v, got %v for point %v", tt.expected, result, tt.point)
            }
        })
    }
}

func TestCollisionSystem_SelfCollision(t *testing.T) {
    config := game.DefaultConfig()
    cs := systems.NewCollisionSystem(config)
    
    segments := []domain.Point{
        {X: 5, Y: 5},
        {X: 5, Y: 6},
        {X: 5, Y: 7},
    }
    
    tests := []struct {
        name     string
        point    domain.Point
        expected bool
    }{
        {"Not colliding", domain.Point{X: 10, Y: 10}, false},
        {"Colliding with first", domain.Point{X: 5, Y: 5}, true},
        {"Colliding with middle", domain.Point{X: 5, Y: 6}, true},
        {"Colliding with last", domain.Point{X: 5, Y: 7}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := cs.CheckSelfCollision(tt.point, segments)
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### File: `tests