package world

import (
	"github.com/WillKirkmanM/paritone/internal/pathfinding"
)

type Block struct {
	Type      string
	Walkable  bool
	Breakable bool
	MoveCost  float64
}

type World struct {
	Blocks map[pathfinding.Point]Block
}

func NewWorld() *World {
	return &World{
		Blocks: make(map[pathfinding.Point]Block),
	}
}

func (w *World) SetBlock(p pathfinding.Point, block Block) {
	w.Blocks[p] = block
}

func (w *World) GetBlock(p pathfinding.Point) (Block, bool) {
	block, exists := w.Blocks[p]
	return block, exists
}

func (w *World) IsWalkable(p pathfinding.Point) bool {
	block, exists := w.Blocks[p]
	if !exists {
		return false
	}
	return block.Walkable
}

func (w *World) CanBreak(p pathfinding.Point) bool {
	block, exists := w.Blocks[p]
	if !exists {
		return false
	}
	return block.Breakable
}

func (w *World) GetBlockType(p pathfinding.Point) string {
	block, exists := w.Blocks[p]
	if !exists {
		return "unknown"
	}
	return block.Type
}

func (w *World) GetMovementCost(from, to pathfinding.Point) float64 {

	baseCost := 1.0

	if from.X != to.X && from.Z != to.Z {
		baseCost = 1.414
	}

	if to.Y > from.Y {
		baseCost += 1.0 * float64(to.Y-from.Y)
	} else if to.Y < from.Y {
		baseCost += 0.2 * float64(from.Y-to.Y)
	}

	toBlock, exists := w.Blocks[to]
	if exists && toBlock.MoveCost > 0 {
		baseCost *= toBlock.MoveCost
	}

	return baseCost
}
