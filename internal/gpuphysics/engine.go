// Package gpuphysics implements optional GPU velocity updates using Ebitengine's
// portable Kage shaders. It must be used from the Ebitengine game lifecycle.
package gpuphysics

import (
	_ "embed"
	"fmt"
	"image"
	"math"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed velocity.kage
var shaderSource []byte

const (
	textureWidth   = 256
	textureHeight  = 128
	velocityScale  = 131072.0
	velocityOffset = 8388608.0
	positionScale  = 16777216.0
)

type Engine struct {
	shader                             *ebiten.Shader
	state, grid, output                *ebiten.Image
	stateBytes, gridBytes, outputBytes []byte
	radii                              [24]float32
	rules                              [144]float32
	options                            ebiten.DrawRectShaderOptions
}

func New() (*Engine, error) {
	shader, err := ebiten.NewShader(shaderSource)
	if err != nil {
		return nil, fmt.Errorf("GPU physics shader: %w", err)
	}
	engine := &Engine{shader: shader}
	bounds := image.Rect(0, 0, textureWidth, textureHeight)
	options := &ebiten.NewImageOptions{Unmanaged: true}
	engine.state = ebiten.NewImageWithOptions(bounds, options)
	engine.grid = ebiten.NewImageWithOptions(bounds, options)
	engine.output = ebiten.NewImageWithOptions(bounds, options)
	engine.stateBytes = make([]byte, textureWidth*textureHeight*4)
	engine.gridBytes = make([]byte, len(engine.stateBytes))
	engine.outputBytes = make([]byte, len(engine.stateBytes))
	engine.options = ebiten.DrawRectShaderOptions{
		Blend:    ebiten.BlendCopy,
		Images:   [4]*ebiten.Image{engine.state, engine.grid},
		Uniforms: make(map[string]any),
	}
	if err := engine.verify(); err != nil {
		engine.Close()
		return nil, err
	}
	return engine, nil
}

func (engine *Engine) Close() {
	engine.state.Deallocate()
	engine.grid.Deallocate()
	engine.output.Deallocate()
	engine.shader.Deallocate()
}

// ComputeVelocities includes uploads and synchronous readback. CPU positions stay
// authoritative, so edits, rendering, restart, and presets need no GPU mirroring.
func (engine *Engine) ComputeVelocities(input simulation.VelocityInput, output []simulation.Velocity) (err error) {
	// Ebitengine reports some unsupported graphics operations as panics. Discard
	// this entire result and let the caller perform the same step on the CPU.
	defer func() {
		if failure := recover(); failure != nil {
			err = fmt.Errorf("GPU physics unavailable: %v", failure)
		}
	}()
	if len(output) != len(input.Particles) || len(output) > simulation.MaxParticles {
		return fmt.Errorf("invalid GPU physics population")
	}
	for index, particle := range input.Particles {
		offset := index * 20
		put24(engine.stateBytes[offset:], min(16777215, int(math.Round(particle.X/input.Config.WorldWidth*positionScale))))
		put24(engine.stateBytes[offset+4:], min(16777215, int(math.Round(particle.Y/input.Config.WorldHeight*positionScale))))
		put24(engine.stateBytes[offset+8:], int(math.Round(particle.VX*velocityScale+velocityOffset)))
		put24(engine.stateBytes[offset+12:], int(math.Round(particle.VY*velocityScale+velocityOffset)))
		next := input.Next[index] + 1
		engine.stateBytes[offset+16] = byte(next)
		engine.stateBytes[offset+17] = byte(next >> 8)
		engine.stateBytes[offset+18] = byte(particle.TypeIndex)
		engine.stateBytes[offset+19] = 255
		column := min(max(int(particle.X/input.CellWidth), 0), input.Columns-1)
		row := min(max(int(particle.Y/input.CellHeight), 0), input.Rows-1)
		put24(engine.gridBytes[(len(input.Heads)+index)*4:], row*input.Columns+column)
	}
	for index, head := range input.Heads {
		put24(engine.gridBytes[index*4:], head+1)
	}
	clear(engine.radii[:])
	clear(engine.rules[:])
	for index, kind := range input.Config.ParticleTypes {
		engine.radii[index*2] = float32(kind.InnerRadius)
		engine.radii[index*2+1] = float32(kind.OuterRadius)
		for other, strength := range input.Config.Relationships[index] {
			engine.rules[index*12+other] = float32(strength)
		}
	}
	uniforms := engine.options.Uniforms
	uniforms["Count"] = len(output)
	uniforms["World"] = []float32{float32(input.Config.WorldWidth), float32(input.Config.WorldHeight)}
	uniforms["Grid"] = []int32{int32(input.Columns), int32(input.Rows)}
	uniforms["Radii"] = engine.radii[:]
	uniforms["Rules"] = engine.rules[:]
	motion := input.Config.Motion
	uniforms["Motion"] = []float32{float32(motion.VelocityRetention), float32(motion.SpeedLimit), float32(motion.RepulsionStrength)}
	engine.state.WritePixels(engine.stateBytes)
	engine.grid.WritePixels(engine.gridBytes)
	engine.output.DrawRectShader(textureWidth, textureHeight, engine.shader, &engine.options)
	// ReadPixels can leave the destination untouched if the graphics backend
	// already has an error. Never accept velocities from the preceding frame.
	clear(engine.outputBytes)
	engine.output.ReadPixels(engine.outputBytes)
	for index := range output {
		offset := index * 8
		if engine.outputBytes[offset+3] != 255 || engine.outputBytes[offset+7] != 255 {
			return fmt.Errorf("GPU physics returned incomplete results")
		}
		output[index] = simulation.Velocity{
			X: (float64(get24(engine.outputBytes[offset:])) - velocityOffset) / velocityScale,
			Y: (float64(get24(engine.outputBytes[offset+4:])) - velocityOffset) / velocityScale,
		}
		// Packing rounds each component independently. Preserve the exact speed
		// cap after readback, including very small user-selected limits.
		if speed := math.Hypot(output[index].X, output[index].Y); speed > motion.SpeedLimit {
			scale := motion.SpeedLimit / speed
			output[index].X *= scale
			output[index].Y *= scale
		}
	}
	return nil
}

func put24(bytes []byte, value int) {
	bytes[0], bytes[1], bytes[2], bytes[3] = byte(value), byte(value>>8), byte(value>>16), 255
}

func get24(bytes []byte) int { return int(bytes[0]) | int(bytes[1])<<8 | int(bytes[2])<<16 }
