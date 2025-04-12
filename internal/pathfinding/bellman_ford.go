package pathfinding

import (
	"math"
	"time"
)

func FindPathBellmanFord(start, goal Point, world World) []Point {
	vertices := getWalkableVertices(start, goal, world)

	dist := make(map[Point]float64)
	pred := make(map[Point]Point)

	for _, v := range vertices {
		dist[v] = math.Inf(1)
	}
	dist[start] = 0

	for i := 0; i < len(vertices)-1; i++ {
		for _, u := range vertices {

			if dist[u] == math.Inf(1) {
				continue
			}

			neighbors := GetWalkableNeighbors(u, world)

			for _, v := range neighbors {
				weight := world.GetMovementCost(u, v)

				if dist[u]+weight < dist[v] {
					dist[v] = dist[u] + weight
					pred[v] = u
				}
			}
		}
	}

	for _, u := range vertices {
		neighbors := GetWalkableNeighbors(u, world)

		for _, v := range neighbors {
			weight := world.GetMovementCost(u, v)

			if dist[u]+weight < dist[v] {

				return nil
			}
		}
	}

	if dist[goal] == math.Inf(1) {

		return nil
	}

	path := []Point{goal}
	current := goal

	for current != start {
		current = pred[current]
		path = append([]Point{current}, path...)
	}

	return path
}

func FindPathBellmanFordWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	vertices := getLocalWalkableVertices(start, goal, world, options)

	dist := make(map[Point]float64)
	pred := make(map[Point]Point)

	for _, v := range vertices {
		dist[v] = math.Inf(1)
	}
	dist[start] = 0

	nodesExplored := 0
	var blocksBroken []Point
	var blocksPlaced []Point
	breakPoints := make(map[Point]bool)
	placePoints := make(map[Point]bool)

	maxMemoryUsed := len(vertices)

	for i := 0; i < len(vertices)-1; i++ {
		anyUpdate := false

		for _, u := range vertices {

			if dist[u] == math.Inf(1) {
				continue
			}

			nodesExplored++

			var neighbors []Point

			if options.AllowBreaking {

				for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
					neighbor := Point{u.X + dir.X, u.Y + dir.Y, u.Z + dir.Z}

					if world.IsWalkable(neighbor) {
						neighbors = append(neighbors, neighbor)
					} else if world.CanBreak(neighbor) {
						neighbors = append(neighbors, neighbor)
						breakPoints[neighbor] = true
					}
				}
			} else if options.AllowPlacing {

				for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {

					neighbor := Point{u.X + dir.X, u.Y + dir.Y, u.Z + dir.Z}
					if world.IsWalkable(neighbor) {
						neighbors = append(neighbors, neighbor)
					}

					placingNeighbor := Point{u.X + dir.X*2, u.Y, u.Z + dir.Z*2}
					if !world.IsWalkable(placingNeighbor) {

						below := Point{placingNeighbor.X, placingNeighbor.Y - 1, placingNeighbor.Z}
						if world.GetBlockType(below) != "air" {
							neighbors = append(neighbors, placingNeighbor)
							placePoints[placingNeighbor] = true
						}
					}
				}
			} else {

				neighbors = GetWalkableNeighbors(u, world)
			}

			for _, v := range neighbors {

				if _, exists := dist[v]; !exists {
					continue
				}

				weight := world.GetMovementCost(u, v)

				if breakPoints[v] {
					weight += 5.0
				} else if placePoints[v] {
					weight += 3.0
				}

				if options.AvoidWater && world.GetBlockType(v) == "water" {
					weight += 10.0
				}

				if options.MinimiseHeight && v.Y != u.Y {
					weight += 2.0 * float64(abs(v.Y-u.Y))
				}

				if dist[u]+weight < dist[v] {
					dist[v] = dist[u] + weight
					pred[v] = u
					anyUpdate = true
				}
			}
		}

		if !anyUpdate {
			break
		}
	}

	for _, u := range vertices {
		neighbors := GetWalkableNeighbors(u, world)

		for _, v := range neighbors {
			if _, exists := dist[v]; !exists {
				continue
			}

			weight := world.GetMovementCost(u, v)

			if dist[u]+weight < dist[v] {

				return PathfindingResult{
					Path:            nil,
					NodesExplored:   nodesExplored,
					ComputationTime: time.Since(startTime),
					MaxMemoryUsed:   maxMemoryUsed,
				}
			}
		}
	}

	if dist[goal] == math.Inf(1) {
		return PathfindingResult{
			Path:            nil,
			NodesExplored:   nodesExplored,
			ComputationTime: time.Since(startTime),
			MaxMemoryUsed:   maxMemoryUsed,
		}
	}

	path := []Point{goal}
	current := goal

	for current != start {
		current = pred[current]
		path = append([]Point{current}, path...)
	}

	waterCrossed := 0
	verticalChange := 0
	lastY := start.Y

	for i, pos := range path {
		if i > 0 {

			vertChange := abs(pos.Y - lastY)
			verticalChange += vertChange
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
		VerticalChange:  verticalChange,
		TotalCost:       dist[goal],
		MaxMemoryUsed:   maxMemoryUsed,
	}
}

func getWalkableVertices(start, goal Point, world World) []Point {
	vertices := []Point{start, goal}

	minX := min(start.X, goal.X) - 30
	maxX := max(start.X, goal.X) + 30
	minY := min(start.Y, goal.Y) - 10
	maxY := max(start.Y, goal.Y) + 10
	minZ := min(start.Z, goal.Z) - 30
	maxZ := max(start.Z, goal.Z) + 30

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				p := Point{x, y, z}
				if world.IsWalkable(p) {
					vertices = append(vertices, p)
				}
			}
		}
	}

	return vertices
}

func getLocalWalkableVertices(start, goal Point, world World, options PathfindingOptions) []Point {

	vertices := make(map[Point]bool)
	vertices[start] = true
	vertices[goal] = true

	queue := []Point{start}
	visited := make(map[Point]bool)
	visited[start] = true

	maxExploration := 5000

	for len(queue) > 0 && len(vertices) < maxExploration {
		current := queue[0]
		queue = queue[1:]

		var neighbors []Point

		if options.AllowBreaking {

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

				if world.IsWalkable(neighbor) || world.CanBreak(neighbor) {
					neighbors = append(neighbors, neighbor)
				}
			}
		} else if options.AllowPlacing {

			neighbors = GetWalkableNeighbors(current, world)
		} else {

			neighbors = GetWalkableNeighbors(current, world)
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				vertices[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	result := make([]Point, 0, len(vertices))
	for v := range vertices {
		result = append(result, v)
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
