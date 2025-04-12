package pathfinding

import (
	"container/heap"
	"math"
	"time"
)

func FindPathThetaStar(start, goal Point, world World) []Point {
	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: start,
		GScore:   0,
		FScore:   Heuristic(start, goal, PathfindingOptions{}),
		Parent:   nil,
	}

	heap.Push(openSet, startNode)

	gScore := make(map[Point]float64)
	gScore[start] = 0

	cameFrom := make(map[Point]*Node)

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)

		if current.Position.X == goal.X && current.Position.Y == goal.Y && current.Position.Z == goal.Z {

			path := []Point{}
			for node := current; node != nil; node = node.Parent {
				path = append([]Point{node.Position}, path...)
			}
			return path
		}

		neighbors := GetWalkableNeighbors(current.Position, world)

		for _, neighbor := range neighbors {

			lineOfSight := false
			parent := current.Parent

			if parent != nil && hasLineOfSight(parent.Position, neighbor, world) {

				directCost := gScore[parent.Position] + euclideanDistance(parent.Position, neighbor)

				if val, exists := gScore[neighbor]; !exists || directCost < val {

					gScore[neighbor] = directCost
					cameFrom[neighbor] = parent

					fScore := directCost + euclideanDistance(neighbor, goal)

					neighborNode := &Node{
						Position: neighbor,
						GScore:   directCost,
						FScore:   fScore,
						Parent:   parent,
					}

					heap.Push(openSet, neighborNode)
					lineOfSight = true
				}
			}

			if !lineOfSight {
				tentativeGScore := gScore[current.Position] + euclideanDistance(current.Position, neighbor)

				if val, exists := gScore[neighbor]; !exists || tentativeGScore < val {
					gScore[neighbor] = tentativeGScore
					cameFrom[neighbor] = current

					fScore := tentativeGScore + euclideanDistance(neighbor, goal)

					neighborNode := &Node{
						Position: neighbor,
						GScore:   tentativeGScore,
						FScore:   fScore,
						Parent:   current,
					}

					heap.Push(openSet, neighborNode)
				}
			}
		}
	}

	return nil
}

func FindPathThetaStarWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
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

	gScore := make(map[Point]float64)
	gScore[start] = 0

	cameFrom := make(map[Point]*Node)

	nodesExplored := 0
	var blocksBroken []Point
	var blocksPlaced []Point
	breakPoints := make(map[Point]bool)
	placePoints := make(map[Point]bool)

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)
		nodesExplored++

		if current.Position.X == goal.X && current.Position.Y == goal.Y && current.Position.Z == goal.Z {
			path := []Point{}
			waterCrossed := 0
			totalVerticalChange := 0

			for node := current; node != nil; node = node.Parent {
				pos := node.Position
				path = append([]Point{pos}, path...)

				if node.Parent != nil {

					directPath := getLineOfSightPoints(node.Parent.Position, pos)

					if len(directPath) > 2 {
						path = append(path[:len(path)-1], directPath...)
					}
				}
			}

			for i, p := range path {
				if i > 0 {
					prevPoint := path[i-1]

					vertChange := abs(p.Y - prevPoint.Y)
					totalVerticalChange += vertChange

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

			for _, dir := range []Point{
				{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
				{1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1},
			} {
				neighbor := Point{current.Position.X + dir.X, current.Position.Y + dir.Y, current.Position.Z + dir.Z}

				if world.IsWalkable(neighbor) {
					neighbors = append(neighbors, neighbor)
				} else if world.CanBreak(neighbor) {
					neighbors = append(neighbors, neighbor)
					breakPoints[neighbor] = true
				}
			}
		} else if options.AllowPlacing {

			for _, dir := range []Point{
				{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
				{1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1},
			} {
				neighbor := Point{current.Position.X + dir.X, current.Position.Y + dir.Y, current.Position.Z + dir.Z}
				if world.IsWalkable(neighbor) {
					neighbors = append(neighbors, neighbor)
				}
			}

			for _, dir := range []Point{
				{2, 0, 0}, {-2, 0, 0}, {0, 0, 2}, {0, 0, -2},
				{2, 0, 2}, {2, 0, -2}, {-2, 0, 2}, {-2, 0, -2},
			} {
				far := Point{current.Position.X + dir.X, current.Position.Y, current.Position.Z + dir.Z}

				if !world.IsWalkable(far) {

					below := Point{far.X, far.Y - 1, far.Z}
					if world.GetBlockType(below) != "air" {
						neighbors = append(neighbors, far)
						placePoints[far] = true
					}
				}
			}
		} else {

			for _, dir := range []Point{
				{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
				{1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1},
				{1, 1, 0}, {-1, 1, 0}, {0, 1, 1}, {0, 1, -1},
				{1, -1, 0}, {-1, -1, 0}, {0, -1, 1}, {0, -1, -1},
			} {
				neighbor := Point{current.Position.X + dir.X, current.Position.Y + dir.Y, current.Position.Z + dir.Z}

				if world.IsWalkable(neighbor) {

					if !options.AvoidWater || world.GetBlockType(neighbor) != "water" {
						neighbors = append(neighbors, neighbor)
					}
				}
			}
		}

		for _, neighbor := range neighbors {

			lineOfSight := false
			parent := current.Parent

			if parent != nil && hasLineOfSight(parent.Position, neighbor, world) {

				moveCost := euclideanDistance(parent.Position, neighbor)

				if breakPoints[neighbor] {
					moveCost += 5.0
				}
				if placePoints[neighbor] {
					moveCost += 3.0
				}

				if options.AvoidWater && world.GetBlockType(neighbor) == "water" {
					moveCost += 10.0
				}

				if options.MinimiseHeight && neighbor.Y != parent.Position.Y {
					moveCost += 2.0 * float64(abs(neighbor.Y-parent.Position.Y))
				}

				directCost := gScore[parent.Position] + moveCost

				if val, exists := gScore[neighbor]; !exists || directCost < val {

					gScore[neighbor] = directCost
					cameFrom[neighbor] = parent

					fScore := directCost + Heuristic(neighbor, goal, options)

					neighborNode := &Node{
						Position: neighbor,
						GScore:   directCost,
						FScore:   fScore,
						Parent:   parent,
					}

					heap.Push(openSet, neighborNode)
					lineOfSight = true
				}
			}

			if !lineOfSight {

				moveCost := euclideanDistance(current.Position, neighbor)

				if breakPoints[neighbor] {
					moveCost += 5.0
				}
				if placePoints[neighbor] {
					moveCost += 3.0
				}

				if options.AvoidWater && world.GetBlockType(neighbor) == "water" {
					moveCost += 10.0
				}

				if options.MinimiseHeight && neighbor.Y != current.Position.Y {
					moveCost += 2.0 * float64(abs(neighbor.Y-current.Position.Y))
				}

				tentativeGScore := gScore[current.Position] + moveCost

				if val, exists := gScore[neighbor]; !exists || tentativeGScore < val {
					gScore[neighbor] = tentativeGScore
					cameFrom[neighbor] = current

					fScore := tentativeGScore + Heuristic(neighbor, goal, options)

					neighborNode := &Node{
						Position: neighbor,
						GScore:   tentativeGScore,
						FScore:   fScore,
						Parent:   current,
					}

					heap.Push(openSet, neighborNode)
				}
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

func hasLineOfSight(from, to Point, world World) bool {

	if !world.IsWalkable(from) || !world.IsWalkable(to) {
		return false
	}

	points := getLineOfSightPoints(from, to)

	for _, p := range points {
		if !world.IsWalkable(p) && !isEqual(p, from) && !isEqual(p, to) {
			return false
		}
	}

	return true
}

func getLineOfSightPoints(from, to Point) []Point {

	points := []Point{}

	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	dz := abs(to.Z - from.Z)

	sx := sign(to.X - from.X)
	sy := sign(to.Y - from.Y)
	sz := sign(to.Z - from.Z)

	var err1, err2 int

	if dx >= dy && dx >= dz {
		err1 = 2*dy - dx
		err2 = 2*dz - dx

		x, y, z := from.X, from.Y, from.Z
		for i := 0; i <= dx; i++ {
			points = append(points, Point{x, y, z})

			if err1 > 0 {
				y += sy
				err1 -= 2 * dx
			}
			if err2 > 0 {
				z += sz
				err2 -= 2 * dx
			}

			err1 += 2 * dy
			err2 += 2 * dz
			x += sx
		}
	} else if dy >= dx && dy >= dz {
		err1 = 2*dx - dy
		err2 = 2*dz - dy

		x, y, z := from.X, from.Y, from.Z
		for i := 0; i <= dy; i++ {
			points = append(points, Point{x, y, z})

			if err1 > 0 {
				x += sx
				err1 -= 2 * dy
			}
			if err2 > 0 {
				z += sz
				err2 -= 2 * dy
			}

			err1 += 2 * dx
			err2 += 2 * dz
			y += sy
		}
	} else {
		err1 = 2*dy - dz
		err2 = 2*dx - dz

		x, y, z := from.X, from.Y, from.Z
		for i := 0; i <= dz; i++ {
			points = append(points, Point{x, y, z})

			if err1 > 0 {
				y += sy
				err1 -= 2 * dz
			}
			if err2 > 0 {
				x += sx
				err2 -= 2 * dz
			}

			err1 += 2 * dy
			err2 += 2 * dx
			z += sz
		}
	}

	return points
}

func euclideanDistance(a, b Point) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	dz := float64(a.Z - b.Z)
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
