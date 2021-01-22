package chip8

import (
	"fmt"
	"io/ioutil"
	"math/rand"
)

var fontSet = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, //0
	0x20, 0x60, 0x20, 0x20, 0x70, //1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
	0x90, 0x90, 0xF0, 0x10, 0x10, //4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
	0xF0, 0x10, 0x20, 0x40, 0x40, //7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
	0xF0, 0x90, 0xF0, 0x90, 0x90, //A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
	0xF0, 0x80, 0x80, 0x80, 0xF0, //C
	0xE0, 0x90, 0x90, 0x90, 0xE0, //D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
	0xF0, 0x80, 0xF0, 0x80, 0x80, //F
}

type CHIP8 struct {
	//Opcode
	opcode uint16

	//Memory
	memory [4096]uint8
	stack  [16]uint16

	//V, I, PC, and SP registers
	v  [16]uint8
	i  uint16
	pc uint16
	sp uint8

	//Display
	display    [64 * 32]uint8
	drawScreen bool

	//Timers
	delayTimer uint8
	soundTimer uint8

	//Keys
	keys [16]uint8
}

func (c *CHIP8) Init() {
	c.pc = 0x200 // Execution starts at 0x200
	c.opcode = 0
	c.i = 0
	c.sp = 0

	c.drawScreen = true

	// Clear display
	for i := 0; i < 64*32; i++ {
		c.display[i] = 0
	}
	// Clear stack
	for i := 0; i < 16; i++ {
		c.stack[i] = 0
	}
	// Clear registers V0-VF
	for i := 0; i < 16; i++ {
		c.v[i] = 0
	}
	// Clear memory
	for i := 0; i < 4096; i++ {
		c.memory[i] = 0
	}
	// Load fontset
	for i := 0; i < 80; i++ {
		c.memory[i] = fontSet[i]
	}
}

func (c *CHIP8) Cycle() {
	//Fetch (2bytes)
	c.opcode = uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])

	//Decode hierarchically (bit-by-bit)
	switch c.opcode & 0xF000 {
	case 0x0000:
		switch c.opcode & 0x000F {
		case 0x0000: //00E0 - CLS
			for i := 0; i < len(c.display); i++ {
				c.display[i] = 0x0
			}
			c.drawScreen = true
			c.pc += 2

		case 0x000E: //00EE - RET
			c.sp -= 1
			c.pc = c.stack[c.sp]
			c.pc += 2

		// Opcode ignored --> 0nnn - SYS addr

		default:
			fmt.Printf("Unknown opcode: 0x%X\n", c.opcode)
		}
	case 0x1000: //1nnn - JP addr
		c.pc = c.opcode & 0x0FFF
	case 0x2000: //2nnn - CALL addr
		c.stack[c.sp] = c.pc
		c.sp += 1
		c.pc = c.opcode & 0x0FFF
	case 0x3000: //3xkk - SE Vx, byte
		if uint16(c.v[(c.opcode&0x0F00)>>8]) == c.opcode&0x00FF {
			c.pc += 4
		} else {
			c.pc += 2
		}
	case 0x4000: //4xkk - SNE Vx, byte
		if uint16(c.v[(c.opcode&0x0F00)>>8]) != c.opcode&0x00FF {
			c.pc += 4
		} else {
			c.pc += 2
		}
	case 0x5000: //5xy0 - SE Vx, Vy
		if c.v[(c.opcode&0x0F00)>>8] == c.v[(c.opcode&0x00F0)>>4] {
			c.pc += 4
		} else {
			c.pc += 2
		}
	case 0x6000: //6xkk - LD Vx, byte
		c.v[(c.opcode&0x0F00)>>8] = uint8(c.opcode & 0x00FF)
		c.pc += 2
	case 0x7000: //7xkk - ADD Vx, byte
		c.v[(c.opcode&0x0F00)>>8] += uint8(c.opcode & 0x00FF)
		c.pc += 2

	case 0x8000:
		switch c.opcode & 0x000F {
		case 0x0000: //8xy0 - LD Vx, Vy
			c.v[(c.opcode&0x0F00)>>8] = c.v[(c.opcode&0x00F0)>>4]
			c.pc += 2
		case 0x0001: //8xy1 - OR Vx, Vy
			c.v[(c.opcode&0x0F00)>>8] |= c.v[(c.opcode&0x00F0)>>4]
			c.pc += 2
		case 0x0002: //8xy2 - AND Vx, Vy
			c.v[(c.opcode&0x0F00)>>8] &= c.v[(c.opcode&0x00F0)>>4]
			c.pc += 2
		case 0x0003: //8xy3 - XOR Vx, Vy
			c.v[(c.opcode&0x0F00)>>8] ^= c.v[(c.opcode&0x00F0)>>4]
			c.pc += 2
		case 0x0004: //8xy4 - ADD Vx, Vy
			if (c.v[(c.opcode&0x0F00)>>8] + c.v[(c.opcode&0x00F0)>>4]) < c.v[(c.opcode&0x0F00)>>8] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[(c.opcode&0x0F00)>>8] += c.v[(c.opcode&0x00F0)>>4]
			c.pc += 2
		case 0x0005: //8xy5 - SUB Vx, Vy
			if c.v[(c.opcode&0x0F00)>>8] > c.v[(c.opcode&0x00F0)>>4] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[(c.opcode&0x0F00)>>8] -= c.v[(c.opcode&0x00F0)>>4]
			c.pc += 2
		case 0x0006: //8xy6 - SHR Vx {, Vy}
			c.v[(c.opcode&0x0F00)>>8] >>= 1
			c.pc += 2
		case 0x0007: //8xy7 - SUBN Vx, Vy
			if c.v[(c.opcode&0x0F00)>>8] > c.v[(c.opcode&0x00F0)>>4] {
				c.v[0xF] = 0
			} else {
				c.v[0xF] = 1
			}
			c.v[(c.opcode&0x0F00)>>8] = c.v[(c.opcode&0x00F0)>>4] - c.v[(c.opcode&0x0F00)>>8]
			c.pc += 2
		case 0x000E: //8xyE - SHL Vx {, Vy}
			c.v[(c.opcode&0x0F00)>>8] <<= 1
			c.pc += 2

		default:
			fmt.Printf("Unknown opcode: 0x%X\n", c.opcode)
		}

	case 0x9000: //9xy0 - SNE Vx, Vy
		if c.v[(c.opcode&0x0F00)>>8] != c.v[(c.opcode&0x00F0)>>4] {
			c.pc += 4
		} else {
			c.pc += 2
		}
	case 0xA000: //Annn - LD I, addr
		c.i = c.opcode & 0x0FFF
		c.pc += 2
	case 0xB000: //Bnnn - JP V0, addr
		c.pc = (c.opcode & 0x0FFF) + uint16(c.v[0x0])
		c.pc += 2
	case 0xC000: //Cxkk - RND Vx, byte
		c.v[(c.opcode&0x0F00)>>8] = uint8(rand.Intn(256)) & uint8(c.opcode&0x00FF)
		c.pc += 2
	case 0xD000: //Dxyn - DRW Vx, Vy, nibble
		x := uint16(c.v[(c.opcode&0x0F00)>>8]) //Vx
		y := uint16(c.v[(c.opcode&0x00F0)>>4]) //Vy

		height := c.opcode & 0x00F
		width := uint16(8) //(all sprites are 8-bits wide)

		c.v[0xF] = 0

		for i := uint16(0); i < height; i++ {
			pixel := c.memory[i+c.i]
			for j := uint16(0); j < width; j++ {
				if (pixel & (0x80 >> i)) != 0 {
					if c.display[(x+j+((y+i)*64))] == 1 {
						c.v[0xF] = 1
					}
					c.display[(x + j + ((y + i) * 64))] ^= 1
				}
			}
		}
		c.drawScreen = true
		c.pc += 2
	case 0xE000:
		switch c.opcode & 0x00FF {
		case 0x009E: //Ex9E - SKP Vx
			if c.keys[c.v[(c.opcode&0x0F00)>>8]] == 1 {
				c.pc += 4
			} else {
				c.pc += 2
			}

		case 0x00A1: //ExA1 - SKNP Vx
			if c.keys[c.v[(c.opcode&0x0F00)>>8]] != 1 {
				c.pc += 4
			} else {
				c.pc += 2
			}

		}
	case 0xF000:
		switch c.opcode & 0x00FF {
		case 0x0007: //Fx07 - LD Vx, DT
			c.v[(c.opcode&0x0F00)>>8] = c.delayTimer
			c.pc += 2
		case 0x000A: //Fx0A - LD Vx, K
			pressed := false
			for i := 0; i < len(c.keys); i++ {
				if c.keys[i] != 0 {
					c.v[(c.opcode&0x0F00)>>8] = uint8(i)
					pressed = true
				}
			}
			if !pressed {
				return
			}
			c.pc += 2
		case 0x0015: //Fx15 - LD DT, Vx
			c.delayTimer = c.v[(c.opcode&0x0F00)>>8]
			c.pc += 2
		case 0x0018: //Fx18 - LD ST, Vx
			c.soundTimer = c.v[(c.opcode&0x0F00)>>8]
			c.pc += 2
		case 0x001E: //Fx1E - ADD I, Vx
			c.i += uint16(c.v[(c.opcode&0x0F00)>>8])
			c.pc += 2
		case 0x0029: //Fx29 - LD F, Vx
			c.i = uint16(c.v[(c.opcode&0x0F00)>>8]) * 0x5
			c.pc += 2
		case 0x0033: //Fx33 - LD B, Vx
			c.memory[c.i] = c.v[(c.opcode&0x0F00)>>8] / 100
			c.memory[c.i+1] = (c.v[(c.opcode&0x0F00)>>8] / 10) % 10
			c.memory[c.i+2] = (c.v[(c.opcode&0x0F00)>>8] % 100) / 10
			c.pc += 2
		case 0x0055: //Fx55 - LD [I], Vx
			for i := 0; i < int((c.opcode&0x0F00)>>8)+1; i++ {
				c.memory[uint16(i)+c.i] = c.v[i]
			}
			c.i = ((c.opcode & 0x0F00) >> 8) + 1
			c.pc += 2
		case 0x0065: //Fx65 - LD Vx, [I]
			for i := 0; i < int((c.opcode&0x0F00)>>8)+1; i++ {
				c.v[i] = c.memory[c.i+uint16(i)]
			}
			c.i = ((c.opcode & 0x0F00) >> 8) + 1
			c.pc += 2
		default:
			fmt.Printf("Unknown opcode: 0x%X\n", c.opcode)
		}
	default:
		fmt.Printf("Unknown opcode: 0x%X\n", c.opcode)
	}

	// Update timers
	if c.delayTimer > 0 {
		c.delayTimer -= 1
	}

	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			fmt.Printf("BEEP!\n")
		}
		c.soundTimer -= 1
	}
}

func (c *CHIP8) Buffer() [64 * 32]uint8 {
	return c.display
}

func (c *CHIP8) Draw() bool {
	draw := c.drawScreen
	c.drawScreen = false
	return draw
}

func (c *CHIP8) Key(key uint8, pressed bool) {
	if pressed {
		c.keys[key] = 1
	} else { //released
		c.keys[key] = 0
	}
}

func (c *CHIP8) LoadROM(fileName string) error {
	buffer, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	if len(c.memory)-512 < len(buffer) { // 0x200 == 512 in decimal (starting address of memory)
		return fmt.Errorf("Program size bigger than memory")
	}

	for i := 0; i < len(buffer); i++ {
		c.memory[i+512] = buffer[i]
	}

	return nil
}

func (c *CHIP8) Print() {
	fmt.Printf("%x\n", c.pc)
}
