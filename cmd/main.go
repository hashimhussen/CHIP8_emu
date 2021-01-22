package main

import (
	chip8 "chip8-emu/pkg"

	"github.com/veandco/go-sdl2/sdl"
)

const CHIP_8_WIDTH int32 = 64
const CHIP_8_HEIGHT int32 = 32
const modifier int32 = 5

func main() {
	chip8 := chip8.CHIP8{}
	chip8.Init()
	chip8.LoadROM("chip8-picture.c8")

	// Initialize sdl2
	if sdlErr := sdl.Init(uint32(sdl.INIT_EVERYTHING)); sdlErr != nil {
		panic(sdlErr)
	}
	defer sdl.Quit()

	// Create window, chip8 resolution with given modifier
	window, windowErr := sdl.CreateWindow("Chip 8 - "+"Pong", int32(sdl.WINDOWPOS_UNDEFINED), int32(sdl.WINDOWPOS_UNDEFINED), CHIP_8_WIDTH*modifier, CHIP_8_HEIGHT*modifier, uint32(sdl.WINDOW_SHOWN))
	if windowErr != nil {
		panic(windowErr)
	}
	defer window.Destroy()

	// Create render surface
	canvas, canvasErr := sdl.CreateRenderer(window, -1, 0)
	if canvasErr != nil {
		panic(canvasErr)
	}
	defer canvas.Destroy()

	for {
		// Emulate one cycle
		chip8.Cycle()

		// Draw only if required
		if chip8.Draw() {
			// Clear the screen
			canvas.SetDrawColor(0, 0, 0, 255)
			canvas.Clear()

			// Get the display buffer and render
			vector := chip8.Buffer()
			for i := 0; i < len(vector); i++ {
				// Values of pixel are stored in 1D array of size 64 * 32
				if vector[i] != 0 {
					canvas.SetDrawColor(255, 255, 255, 255)
				} else {
					canvas.SetDrawColor(0, 0, 0, 255)
				}
				canvas.FillRect(&sdl.Rect{
					Y: int32(20) * modifier,
					X: int32(i) * modifier,
					W: modifier,
					H: modifier,
				})
			}

			canvas.Present()

		}
		// Chip8 cpu clock worked at frequency of 60Hz, so set delay to (1000/60)ms
		sdl.Delay(1000 / 60)
	}
}
