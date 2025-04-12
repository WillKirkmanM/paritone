package pathfinding

import (
	"container/heap"
	"time"
)

func FindPathGreedy(start, goal Point, world World) []Point {
	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: start,
		GScore:   0,
		FScore:   Heuristic(start, goal, PathfindingOptions{}),
		Parent:   nil,
	}

	heap.Push(openSet, startNode)

	visited := make(map[Point]bool)

	cameFrom := make(map[Point]*Node)

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)

		if visited[current.Position] {
			continue
		}

		visited[current.Position] = true

		if current.Position.X == goal.X && current.Position.Y == goal.Y && current.Position.Z == goal.Z {

			path := []Point{}
			for node := current; node != nil; node = node.Parent {
				path = append([]Point{node.Position}, path...)
			}
			return path
		}

		neighbors := GetWalkableNeighbors(current.Position, world)

		for _, neighbor := range neighbors {
			if !visited[neighbor] {

				hScore := Heuristic(neighbor, goal, PathfindingOptions{})

				neighborNode := &Node{
					Position: neighbor,
					GScore:   current.GScore + 1,
					FScore:   hScore,
					Parent:   current,
				}

				heap.Push(openSet, neighborNode)
				cameFrom[neighbor] = current
			}
		}
	}

	return nil
}

func FindPathGreedyWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: start,
		GScore:   0,
		FScore:   Heuristic(start, goal, options),
		Parent:   nil,
	}

	heap.Push(openSet, startNode)

	visited := make(map[Point]bool)

	cameFrom := make(map[Point]*Node)

	nodesExplored := 0
	var blocksBroken []Point
	var blocksPlaced []Point
	breakPoints := make(map[Point]bool)
	placePoints := make(map[Point]bool)

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)
		nodesExplored++

		if visited[current.Position] {
			continue
		}

		visited[current.Position] = true

		if current.Position.X == goal.X && current.Position.Y == goal.Y && current.Position.Z == goal.Z {

			path := []Point{}
			totalVerticalChange := 0
			waterCrossed := 0
			lastY := start.Y

			for node := current; node != nil; node = node.Parent {
				pos := node.Position
				path = append([]Point{pos}, path...)

				if node.Parent != nil {

					vertChange := abs(pos.Y - lastY)
					totalVerticalChange += vertChange
					lastY = pos.Y

					if world.GetBlockType(pos) == "water" {
						waterCrossed++
					}

					if breakPoints[pos] {
						blocksBroken = append(blocksBroken, pos)
					}
					if placePoints[pos] {
						blocksPlaced = append(blocksPlaced, pos)
					}
				}
			}

			return PathfindingResult{
				Path:            path,
				NodesExplored:   nodesExplored,
				ComputationTime: time.Since(startTime),
				BlocksBroken:    blocksBroken,
				BlocksPlaced:    blocksPlaced,
				WaterCrossed:    waterCrossed,
				VerticalChange:  totalVerticalChange,
				TotalCost:       current.GScore,
			}
		}

		var neighbors []Point

		if options.AllowBreaking {

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.Position.X + dir.X, current.Position.Y + dir.Y, current.Position.Z + dir.Z}

				if world.IsWalkable(neighbor) {
					neighbors = append(neighbors, neighbor)
				} else if world.CanBreak(neighbor) {
					neighbors = append(neighbors, neighbor)
					breakPoints[neighbor] = true
				}
			}
		} else if options.AllowPlacing {

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {

				neighbor := Point{current.Position.X + dir.X, current.Position.Y + dir.Y, current.Position.Z + dir.Z}
				if world.IsWalkable(neighbor) {
					neighbors = append(neighbors, neighbor)
				}

				placingNeighbor := Point{current.Position.X + dir.X*2, current.Position.Y, current.Position.Z + dir.Z*2}
				if !world.IsWalkable(placingNeighbor) {

					below := Point{placingNeighbor.X, placingNeighbor.Y - 1, placingNeighbor.Z}
					if world.GetBlockType(below) != "air" {
						neighbors = append(neighbors, placingNeighbor)
						placePoints[placingNeighbor] = true
					}
				}
			}
		} else {

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.Position.X + dir.X, current.Position.Y + dir.Y, current.Position.Z + dir.Z}

				if world.IsWalkable(neighbor) {

					if !options.AvoidWater || world.GetBlockType(neighbor) != "water" {
						neighbors = append(neighbors, neighbor)
					}
				}
			}
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor] {

				hScore := Heuristic(neighbor, goal, options)

				moveCost := 1.0

				if breakPoints[neighbor] {
					moveCost += 5.0
				}
				if placePoints[neighbor] {
					moveCost += 3.0
				}

				if world.GetBlockType(neighbor) == "water" {
					if options.AvoidWater {
						hScore += 10.0
					}
					moveCost += 2.0
				}

				if options.MinimiseHeight && neighbor.Y != current.Position.Y {
					hScore += 2.0 * float64(abs(neighbor.Y-current.Position.Y))
				}

				gScore := current.GScore + moveCost

				neighborNode := &Node{
					Position: neighbor,
					GScore:   gScore,
					FScore:   hScore,
					Parent:   current,
				}

				heap.Push(openSet, neighborNode)
				cameFrom[neighbor] = current
			}
		}
	}

	return PathfindingResult{
		Path:            nil,
		NodesExplored:   nodesExplored,
		ComputationTime: time.Since(startTime),
		BlocksBroken:    blocksBroken,
		BlocksPlaced:    blocksPlaced,
	}
}
