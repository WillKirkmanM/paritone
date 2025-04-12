package pathfinding

import (
	"math"
	"time"
)

type PathfindingOptions struct {
	AllowBreaking         bool
	AllowPlacing          bool
	AvoidWater            bool
	MinimiseHeight        bool
	JumpPointOptimisation bool
	MaxIterations         int
	HeuristicWeight       float64
}

type PathfindingResult struct {
	Path            []Point
	NodesExplored   int
	ComputationTime time.Duration
	BlocksBroken    []Point
	BlocksPlaced    []Point
	WaterCrossed    int
	VerticalChange  int
	TotalCost       float64
	MaxMemoryUsed   int
	Iterations      int
	OptimalityRatio float64
}

type World interface {
	IsWalkable(p Point) bool
	CanBreak(p Point) bool
	GetBlockType(p Point) string
	GetMovementCost(from, to Point) float64
}

func FindPathWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	var result PathfindingResult

	if options.AllowBreaking {

		path, nodesExplored, blocksBroken, waterCrossed, verticalChange, totalCost :=
			findPathWithBreaking(start, goal, world, options)

		result = PathfindingResult{
			Path:            path,
			NodesExplored:   nodesExplored,
			ComputationTime: time.Since(startTime),
			BlocksBroken:    blocksBroken,
			WaterCrossed:    waterCrossed,
			VerticalChange:  verticalChange,
			TotalCost:       totalCost,
		}
	} else if options.AllowPlacing {

		path, nodesExplored, blocksPlaced, waterCrossed, verticalChange, totalCost :=
			findPathWithPlacing(start, goal, world, options)

		result = PathfindingResult{
			Path:            path,
			NodesExplored:   nodesExplored,
			ComputationTime: time.Since(startTime),
			BlocksPlaced:    blocksPlaced,
			WaterCrossed:    waterCrossed,
			VerticalChange:  verticalChange,
			TotalCost:       totalCost,
		}
	} else {

		path, nodesExplored, waterCrossed, verticalChange, totalCost :=
			findPathStandard(start, goal, world, options)

		result = PathfindingResult{
			Path:            path,
			NodesExplored:   nodesExplored,
			ComputationTime: time.Since(startTime),
			WaterCrossed:    waterCrossed,
			VerticalChange:  verticalChange,
			TotalCost:       totalCost,
		}
	}

	return result
}

func ManhattanDistance(a, b Point) float64 {
	return float64(abs(a.X-b.X) + abs(a.Y-b.Y) + abs(a.Z-b.Z))
}

func EuclideanDistance(a, b Point) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	dz := float64(a.Z - b.Z)
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func ChebyshevDistance(a, b Point) float64 {
	return float64(max(abs(a.X-b.X), max(abs(a.Y-b.Y), abs(a.Z-b.Z))))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func findPathWithPlacing(start, goal Point, world World, options PathfindingOptions) (
	[]Point, int, []Point, int, int, float64) {

	path, nodesExplored, waterCrossed, verticalChange, totalCost :=
		findPathStandard(start, goal, world, options)

	return path, nodesExplored, []Point{}, waterCrossed, verticalChange, totalCost
}

func findPathStandard(start, goal Point, world World, options PathfindingOptions) (
	[]Point, int, int, int, float64) {

	path := FindPath(start, goal, world)

	nodesExplored := 0
	waterCrossed := 0
	verticalChange := 0
	totalCost := 0.0

	if path != nil {
		lastY := start.Y
		for _, p := range path {
			verticalChange += abs(p.Y - lastY)
			lastY = p.Y

			if world.GetBlockType(p) == "water" {
				waterCrossed++
			}

			totalCost += 1.0
		}
	}

	return path, nodesExplored, waterCrossed, verticalChange, totalCost
}
