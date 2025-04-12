package pathfinding

import (
	"container/list"
	"time"
)

func FindPathBFS(start, goal Point, world World) []Point {
	queue := list.New()
	queue.PushBack(start)

	visited := make(map[Point]bool)
	visited[start] = true

	cameFrom := make(map[Point]Point)

	for queue.Len() > 0 {

		current := queue.Remove(queue.Front()).(Point)

		if current.X == goal.X && current.Y == goal.Y && current.Z == goal.Z {

			path := []Point{}
			for current != start {
				path = append([]Point{current}, path...)
				current = cameFrom[current]
			}
			path = append([]Point{start}, path...)
			return path
		}

		for _, neighbor := range GetNeighbors(current, world) {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue.PushBack(neighbor)
				cameFrom[neighbor] = current
			}
		}
	}

	return nil
}

func FindPathBFSWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	nodesExplored := 0
	var blocksBroken []Point
	var blocksPlaced []Point
	waterCrossed := 0

	queue := list.New()
	queue.PushBack(start)

	visited := make(map[Point]bool)
	visited[start] = true

	cameFrom := make(map[Point]Point)
	breakPoints := make(map[Point]bool)
	placePoints := make(map[Point]bool)

	for queue.Len() > 0 {

		current := queue.Remove(queue.Front()).(Point)
		nodesExplored++

		if current.X == goal.X && current.Y == goal.Y && current.Z == goal.Z {

			path := []Point{}
			totalVerticalChange := 0
			lastY := start.Y

			for p := current; !isEqual(p, start); p = cameFrom[p] {
				path = append([]Point{p}, path...)

				vertChange := abs(p.Y - lastY)
				totalVerticalChange += vertChange
				lastY = p.Y

				if world.GetBlockType(p) == "water" {
					waterCrossed++
				}

				if breakPoints[p] {
					blocksBroken = append(blocksBroken, p)
				}

				if placePoints[p] {
					blocksPlaced = append(blocksPlaced, p)
				}
			}
			path = append([]Point{start}, path...)

			return PathfindingResult{
				Path:            path,
				NodesExplored:   nodesExplored,
				ComputationTime: time.Since(startTime),
				BlocksBroken:    blocksBroken,
				BlocksPlaced:    blocksPlaced,
				WaterCrossed:    waterCrossed,
				VerticalChange:  totalVerticalChange,
				TotalCost:       float64(len(path)),
			}
		}

		var neighbors []Point

		if options.AllowBreaking {

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

				if world.IsWalkable(neighbor) {
					neighbors = append(neighbors, neighbor)
				} else if world.CanBreak(neighbor) {
					neighbors = append(neighbors, neighbor)
					breakPoints[neighbor] = true
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
						placePoints[placingNeighbor] = true
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

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue.PushBack(neighbor)
				cameFrom[neighbor] = current
			}
		}
	}

	return PathfindingResult{
		Path:            nil,
		NodesExplored:   nodesExplored,
		ComputationTime: time.Since(startTime),
	}
}

func isEqual(a, b Point) bool {
	return a.X == b.X && a.Y == b.Y && a.Z == b.Z
}
