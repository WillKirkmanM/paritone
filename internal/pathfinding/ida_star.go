package pathfinding

import (
	"math"
	"time"
)

func FindPathIDA(start, goal Point, world World) []Point {
	bound := ManhattanDistance(start, goal)

	for {
		visited := make(map[Point]bool)

		t := idaSearchDFS(start, 0, bound, goal, world, nil, visited, make([]Point, 0))

		if path, ok := t.([]Point); ok {
			return path
		}

		if newBound, ok := t.(float64); ok && newBound == math.Inf(1) {
			return nil
		}

		bound = t.(float64)
	}
}

func idaSearchDFS(current Point, g float64, bound float64, goal Point, world World, parent *Point, visited map[Point]bool, path []Point) interface{} {
	f := g + ManhattanDistance(current, goal)

	if f > bound {
		return f
	}

	if current.X == goal.X && current.Y == goal.Y && current.Z == goal.Z {
		return append(path, current)
	}

	visited[current] = true

	path = append(path, current)

	minBound := math.Inf(1)

	neighbors := GetWalkableNeighbors(current, world)

	for _, neighbor := range neighbors {

		if parent != nil && neighbor.X == parent.X && neighbor.Y == parent.Y && neighbor.Z == parent.Z {
			continue
		}
		if visited[neighbor] {
			continue
		}

		cost := world.GetMovementCost(current, neighbor)

		t := idaSearchDFS(neighbor, g+cost, bound, goal, world, &current, visited, path)

		if _, ok := t.([]Point); ok {
			return t
		}

		if newBound, ok := t.(float64); ok && newBound < minBound {
			minBound = newBound
		}
	}

	visited[current] = false

	return minBound
}

func FindPathIDAWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	maxIterations := options.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 1000
	}

	nodesExplored := 0
	iterations := 0

	bound := Heuristic(start, goal, options)

	var blocksBroken []Point
	var blocksPlaced []Point
	breakPoints := make(map[Point]bool)
	placePoints := make(map[Point]bool)

	var finalPath []Point

	for iterations < maxIterations {
		iterations++

		visited := make(map[Point]bool)
		pathStack := []Point{start}
		gScores := make(map[Point]float64)
		gScores[start] = 0
		parents := make(map[Point]Point)

		result, newBound, explored, manipulationPoints :=
			idaSearchWithOptions(start, 0, bound, goal, world, options, visited,
				pathStack, gScores, parents, breakPoints, placePoints)

		nodesExplored += explored

		for p := range manipulationPoints.breaks {
			breakPoints[p] = true
		}
		for p := range manipulationPoints.places {
			placePoints[p] = true
		}

		if result {

			finalPath = reconstructPath(goal, parents)
			break
		}

		if newBound == math.Inf(1) || iterations >= maxIterations {
			break
		}

		bound = newBound
	}

	if finalPath == nil {
		return PathfindingResult{
			Path:            nil,
			NodesExplored:   nodesExplored,
			ComputationTime: time.Since(startTime),
			Iterations:      iterations,
		}
	}

	waterCrossed := 0
	verticalChange := 0
	lastY := start.Y
	totalCost := 0.0

	for i, pos := range finalPath {
		if i > 0 {

			vertChange := abs(pos.Y - lastY)
			verticalChange += vertChange
			lastY = pos.Y

			if world.GetBlockType(pos) == "water" {
				waterCrossed++
			}

			prev := finalPath[i-1]
			totalCost += world.GetMovementCost(prev, pos)

			if breakPoints[pos] {
				blocksBroken = append(blocksBroken, pos)
			}
			if placePoints[pos] {
				blocksPlaced = append(blocksPlaced, pos)
			}
		}
	}

	return PathfindingResult{
		Path:            finalPath,
		NodesExplored:   nodesExplored,
		ComputationTime: time.Since(startTime),
		BlocksBroken:    blocksBroken,
		BlocksPlaced:    blocksPlaced,
		WaterCrossed:    waterCrossed,
		VerticalChange:  verticalChange,
		TotalCost:       totalCost,
		Iterations:      iterations,
	}
}

type manipulationPoints struct {
	breaks map[Point]bool
	places map[Point]bool
}

func idaSearchWithOptions(
	current Point,
	g float64,
	bound float64,
	goal Point,
	world World,
	options PathfindingOptions,
	visited map[Point]bool,
	pathStack []Point,
	gScores map[Point]float64,
	parents map[Point]Point,
	breakPoints, placePoints map[Point]bool,
) (bool, float64, int, manipulationPoints) {
	f := g + Heuristic(current, goal, options)

	if f > bound {
		return false, f, 1, manipulationPoints{
			breaks: make(map[Point]bool),
			places: make(map[Point]bool),
		}
	}

	if current.X == goal.X && current.Y == goal.Y && current.Z == goal.Z {
		return true, bound, 1, manipulationPoints{
			breaks: make(map[Point]bool),
			places: make(map[Point]bool),
		}
	}

	visited[current] = true

	var neighbors []Point
	localBreakPoints := make(map[Point]bool)
	localPlacePoints := make(map[Point]bool)

	if options.AllowBreaking {

		for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
			neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

			if world.IsWalkable(neighbor) {
				neighbors = append(neighbors, neighbor)
			} else if world.CanBreak(neighbor) {
				neighbors = append(neighbors, neighbor)
				localBreakPoints[neighbor] = true
			}
		}
	} else if options.AllowPlacing {

		for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {

			neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}
			if world.IsWalkable(neighbor) {
				neighbors = append(neighbors, neighbor)
			}

			placingNeighbor := Point{current.X + dir.X*2, current.Y, current.Z + dir.Z*2}
			if !world.IsWalkable(placingNeighbor) {

				below := Point{placingNeighbor.X, placingNeighbor.Y - 1, placingNeighbor.Z}
				if world.GetBlockType(below) != "air" {
					neighbors = append(neighbors, placingNeighbor)
					localPlacePoints[placingNeighbor] = true
				}
			}
		}
	} else {

		for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
			neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

			if world.IsWalkable(neighbor) {

				if !options.AvoidWater || world.GetBlockType(neighbor) != "water" {
					neighbors = append(neighbors, neighbor)
				}
			}
		}
	}

	minBound := math.Inf(1)
	totalExplored := 1

	sortNeighborsByHeuristic(neighbors, goal, options)

	allManipulationPoints := manipulationPoints{
		breaks: make(map[Point]bool),
		places: make(map[Point]bool),
	}

	for _, neighbor := range neighbors {

		if visited[neighbor] {
			continue
		}

		cost := world.GetMovementCost(current, neighbor)

		if localBreakPoints[neighbor] {
			cost += 5.0
		} else if localPlacePoints[neighbor] {
			cost += 3.0
		}

		if options.AvoidWater && world.GetBlockType(neighbor) == "water" {
			cost += 10.0
		}

		if options.MinimiseHeight && neighbor.Y != current.Y {
			cost += 2.0 * float64(abs(neighbor.Y-current.Y))
		}

		newG := g + cost

		parents[neighbor] = current
		gScores[neighbor] = newG

		found, newBound, explored, manipPoints := idaSearchWithOptions(
			neighbor, newG, bound, goal, world, options, visited,
			append(pathStack, neighbor), gScores, parents,
			breakPoints, placePoints,
		)

		totalExplored += explored

		for p := range manipPoints.breaks {
			allManipulationPoints.breaks[p] = true
		}
		for p := range manipPoints.places {
			allManipulationPoints.places[p] = true
		}

		if found {

			for p := range localBreakPoints {
				allManipulationPoints.breaks[p] = true
			}
			for p := range localPlacePoints {
				allManipulationPoints.places[p] = true
			}
			return true, bound, totalExplored, allManipulationPoints
		}

		if newBound < minBound {
			minBound = newBound
		}
	}

	for p := range localBreakPoints {
		allManipulationPoints.breaks[p] = true
	}
	for p := range localPlacePoints {
		allManipulationPoints.places[p] = true
	}

	visited[current] = false

	return false, minBound, totalExplored, allManipulationPoints
}

func sortNeighborsByHeuristic(neighbors []Point, goal Point, options PathfindingOptions) {
	for i := 0; i < len(neighbors); i++ {
		for j := i + 1; j < len(neighbors); j++ {
			h1 := Heuristic(neighbors[i], goal, options)
			h2 := Heuristic(neighbors[j], goal, options)

			if h1 > h2 {
				neighbors[i], neighbors[j] = neighbors[j], neighbors[i]
			}
		}
	}
}

func reconstructPath(goal Point, parents map[Point]Point) []Point {
	path := []Point{goal}
	current := goal

	for {
		parent, exists := parents[current]
		if !exists {
			break
		}

		path = append([]Point{parent}, path...)
		current = parent
	}

	return path
}
