package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GridSize    int   = 1024
	NbParticles int   = 2500
	Fps         int32 = 120
)

var (
	initRadius          float64    = 15.0
	radius              float64    = initRadius
	backgroundColor     color.RGBA = rl.Black
	movingParticleColor color.RGBA = rl.Gray
	particleColor       color.RGBA = rl.White
	showUI              bool       = true
	step                float32    = 1
)

type Cell struct {
	Row    int
	Col    int
	Filled bool
}

type Grid struct {
	Rows  int
	Cols  int
	Cells [][]Cell
}

type Particle struct {
	Row int
	Col int
}

type Particles struct {
	Particles []*Particle
}

var (
	grid      Grid
	syncM     sync.Mutex
	particles Particles
	src       = rand.NewSource(time.Now().UnixNano())
	r         = rand.New(src)
)

func reset() {
	radius = initRadius
	initialCells := make([][]Cell, GridSize)
	for i := int(0); i < GridSize; i++ {
		initialCells[i] = make([]Cell, GridSize)
		for j := int(0); j < GridSize; j++ {
			initialCells[i][j] = Cell{Row: i, Col: j}
		}
	}
	initialCells[GridSize/2][GridSize/2].Filled = true

	grid = Grid{Rows: GridSize, Cols: GridSize, Cells: initialCells}

	particles.Particles = make([]*Particle, NbParticles)
	for i := range particles.Particles {
		particles.Particles[i] = newParticle()
		particles.Particles[i].Respawn()
	}
}

func newParticle() *Particle {
	return new(Particle)
}

func (p *Particle) Respawn() {
	t := rand.Float64() * (2.0 * math.Pi)
	x := int(radius * math.Cos(t))
	y := int(radius * math.Sin(t))
	p.Row = x + (GridSize / 2)
	p.Col = y + (GridSize / 2)
}

func main() {
	reset()
	rl.InitWindow(int32(GridSize), int32(GridSize), "DLA")
	defer rl.CloseWindow()

	rl.SetTargetFPS(Fps)

	for !rl.WindowShouldClose() {
		rl.SetTargetFPS(Fps)
		rl.BeginDrawing()
		rl.ClearBackground(backgroundColor)
		grid.Draw()
		particles.Update()
		if showUI {
			particles.Draw()
			step = gui.Slider(rl.Rectangle{10, 10, 165, 20}, "",
				fmt.Sprintf("STEP: %2.0f", step), step, 1, 50)

			if gui.Button(rl.Rectangle{10, 35, 100, 25}, "Reset") {
				reset()
			}
			rl.DrawText("H: Hide the UI / S: Take a screenshot", 10, int32(GridSize)-20, 15, particleColor)
		}
		if rl.IsKeyPressed(rl.KeyH) {
			showUI = !showUI
		}
		if rl.IsKeyPressed(rl.KeyS) {
			t := time.Now().Unix()
			rl.TakeScreenshot(fmt.Sprintf("%v.png", t))

			if err := os.Rename(fmt.Sprintf("%v.png", t), fmt.Sprintf("snapshots/%v.png", t)); err != nil {
				log.Fatal(err)
			}
		}
		rl.EndDrawing()
	}
}

func (g *Grid) CellAt(row int, col int) *Cell {
	if row >= GridSize || col >= GridSize {
		return nil
	}
	if row < 0 || col < 0 {
		return nil
	}
	return &g.Cells[row][col]
}

func (g *Grid) WalkCells(callback func(grid *Grid, cell *Cell)) {
	for i, row := range g.Cells {
		for j := range row {
			callback(g, &g.Cells[i][j])
		}
	}
}

func (g *Grid) Draw() {
	g.WalkCells(func(grid *Grid, cell *Cell) {
		if cell.Filled {
			rl.DrawPixel(int32(cell.Col), int32(cell.Row), particleColor)
		}
	})
}

func (p *Particles) Update() {
	for i := range p.Particles {
		p.Particles[i].Move()
		if p.Particles[i].HasNeighbours() {
			if c := grid.CellAt(p.Particles[i].Row, p.Particles[i].Col); c != nil {
				grid.CellAt(p.Particles[i].Row, p.Particles[i].Col).Filled = true
				if r := p.Particles[i].getRadius(); r > (0.95 * radius) {
					syncM.Lock()
					radius = radius * 1.05
					syncM.Unlock()
				}
			}
			p.Particles[i].Respawn()
		}

		if r := p.Particles[i].getRadius(); r > (1.05 * radius) {
			p.Particles[i].Respawn()
		}
	}
}

func (p *Particles) Draw() {
	for i := range p.Particles {
		rl.DrawPixel(int32(p.Particles[i].Col), int32(p.Particles[i].Row), movingParticleColor)
	}
}

// func (p *Particle) Move() {
// 	x := p.Row - (GridSize / 2)
// 	y := p.Col - (GridSize / 2)
// 	p.Row += y
// 	p.Col -= x
// }

func (p *Particle) Move() {
	p.Row += int(math.Pow(float64(-1), float64(rand.Intn(2)))) * rand.Intn(int(step)+1)
	p.Col += int(math.Pow(float64(-1), float64(rand.Intn(2)))) * rand.Intn(int(step)+1)
	// p.Row += int(math.Pow(float64(-1), float64(rand.Intn(2))))
	// p.Col += int(math.Pow(float64(-1), float64(rand.Intn(2))))
}

func (p *Particle) HasNeighbours() bool {
	for x := -1; x < 2; x++ {
		for y := -1; y < 2; y++ {
			if c := grid.CellAt(p.Row+x, p.Col+y); c != nil && c.Filled {
				return true
			}
		}
	}
	return false
}

func (p *Particle) getRadius() float64 {
	return math.Sqrt(math.Pow(float64(p.Row-GridSize/2), 2.0) + math.Pow(float64(p.Col-GridSize/2), 2.0))
}
