package pathfinding

import (
	"container/heap"
	"time"
)

func FindPathJPS(start, goal Point, world World) []Point {
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

		successors := identifySuccessors(current.Position, goal, world, current.Parent)

		for _, successor := range successors {

			tentativeGScore := gScore[current.Position] + Heuristic(current.Position, successor, PathfindingOptions{})

			if val, exists := gScore[successor]; !exists || tentativeGScore < val {
				gScore[successor] = tentativeGScore
				fScore := tentativeGScore + Heuristic(successor, goal, PathfindingOptions{})

				successorNode := &Node{
					Position: successor,
					GScore:   tentativeGScore,
					FScore:   fScore,
					Parent:   current,
				}

				cameFrom[successor] = current
				heap.Push(openSet, successorNode)
			}
		}
	}

	return nil
}

func FindPathJPSWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	if options.AllowBreaking || options.AllowPlacing || options.AvoidWater {

		return FindPathWithOptions(start, goal, world, options)
	}

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

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)
		nodesExplored++

		if current.Position.X == goal.X && current.Position.Y == goal.Y && current.Position.Z == goal.Z {

			path := []Point{}
			waterCrossed := 0
			totalVerticalChange := 0
			lastY := start.Y

			for node := current; node != nil; node = node.Parent {
				path = append([]Point{node.Position}, path...)

				if node.Parent != nil {

					interpolatedPoints := interpolatePath(node.Parent.Position, node.Position, world)
					path = append(path[:len(path)-1], interpolatedPoints...)
				}
			}

			for i, p := range path {
				if i > 0 {
					vertChange := abs(p.Y - lastY)
					totalVerticalChange += vertChange
					lastY = p.Y

					if world.GetBlockType(p) == "water" {
						waterCrossed++
					}
				}
			}

			return PathfindingResult{
				Path:            path,
				NodesExplored:   nodesExplored,
				ComputationTime: time.Since(startTime),
				WaterCrossed:    waterCrossed,
				VerticalChange:  totalVerticalChange,
				TotalCost:       current.GScore,
			}
		}

		successors := identifySuccessors(current.Position, goal, world, current.Parent)

		for _, successor := range successors {

			tentativeGScore := gScore[current.Position] +
				float64(manhattanDistance(current.Position, successor))

			if val, exists := gScore[successor]; !exists || tentativeGScore < val {
				gScore[successor] = tentativeGScore
				fScore := tentativeGScore + Heuristic(successor, goal, options)

				successorNode := &Node{
					Position: successor,
					GScore:   tentativeGScore,
					FScore:   fScore,
					Parent:   current,
				}

				cameFrom[successor] = current
				heap.Push(openSet, successorNode)
			}
		}
	}

	return PathfindingResult{
		Path:            nil,
		NodesExplored:   nodesExplored,
		ComputationTime: time.Since(startTime),
	}
}

func identifySuccessors(current, goal Point, world World, parent *Node) []Point {
	successors := []Point{}

	neighbors := getPrunedNeighbors(current, parent, world)

	for _, neighbor := range neighbors {

		dx := sign(neighbor.X - current.X)
		dy := sign(neighbor.Y - current.Y)
		dz := sign(neighbor.Z - current.Z)

		jp := jump(current, dx, dy, dz, goal, world)

		if jp.X != -1 {
			successors = append(successors, jp)
		}
	}

	return successors
}

func getPrunedNeighbors(current Point, parent *Node, world World) []Point {
	neighbors := []Point{}

	if parent == nil {
		for _, dir := range []Point{
			{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
			{1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1},
		} {
			neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}
			if world.IsWalkable(neighbor) {
				neighbors = append(neighbors, neighbor)
			}
		}
		return neighbors
	}

	parentPos := parent.Position

	dx := sign(current.X - parentPos.X)
	dy := sign(current.Y - parentPos.Y)
	dz := sign(current.Z - parentPos.Z)

	if dx != 0 && dz != 0 {

		diag := Point{current.X + dx, current.Y, current.Z + dz}
		if world.IsWalkable(diag) {
			neighbors = append(neighbors, diag)
		}

		if !world.IsWalkable(Point{current.X - dx, current.Y, current.Z}) {
			horizontal := Point{current.X - dx, current.Y, current.Z + dz}
			if world.IsWalkable(horizontal) {
				neighbors = append(neighbors, horizontal)
			}
		}

		if !world.IsWalkable(Point{current.X, current.Y, current.Z - dz}) {
			vertical := Point{current.X + dx, current.Y, current.Z - dz}
			if world.IsWalkable(vertical) {
				neighbors = append(neighbors, vertical)
			}
		}

		if dy != 0 {
			upDown := Point{current.X, current.Y + dy, current.Z}
			if world.IsWalkable(upDown) {
				neighbors = append(neighbors, upDown)
			}
		}
	} else if dx != 0 {

		horz := Point{current.X + dx, current.Y, current.Z}
		if world.IsWalkable(horz) {
			neighbors = append(neighbors, horz)
		}

		if !world.IsWalkable(Point{current.X, current.Y + 1, current.Z}) {
			upNeighbor := Point{current.X + dx, current.Y + 1, current.Z}
			if world.IsWalkable(upNeighbor) {
				neighbors = append(neighbors, upNeighbor)
			}
		}

		if !world.IsWalkable(Point{current.X, current.Y - 1, current.Z}) {
			downNeighbor := Point{current.X + dx, current.Y - 1, current.Z}
			if world.IsWalkable(downNeighbor) {
				neighbors = append(neighbors, downNeighbor)
			}
		}

		if dy != 0 {
			upDown := Point{current.X, current.Y + dy, current.Z}
			if world.IsWalkable(upDown) {
				neighbors = append(neighbors, upDown)
			}
		}
	} else if dz != 0 {

		zDir := Point{current.X, current.Y, current.Z + dz}
		if world.IsWalkable(zDir) {
			neighbors = append(neighbors, zDir)
		}

		if !world.IsWalkable(Point{current.X, current.Y + 1, current.Z}) {
			upNeighbor := Point{current.X, current.Y + 1, current.Z + dz}
			if world.IsWalkable(upNeighbor) {
				neighbors = append(neighbors, upNeighbor)
			}
		}

		if !world.IsWalkable(Point{current.X, current.Y - 1, current.Z}) {
			downNeighbor := Point{current.X, current.Y - 1, current.Z + dz}
			if world.IsWalkable(downNeighbor) {
				neighbors = append(neighbors, downNeighbor)
			}
		}

		if dy != 0 {
			upDown := Point{current.X, current.Y + dy, current.Z}
			if world.IsWalkable(upDown) {
				neighbors = append(neighbors, upDown)
			}
		}
	} else if dy != 0 {

		vert := Point{current.X, current.Y + dy, current.Z}
		if world.IsWalkable(vert) {
			neighbors = append(neighbors, vert)
		}

		if !world.IsWalkable(Point{current.X + 1, current.Y, current.Z}) {
			xNeighbor := Point{current.X + 1, current.Y + dy, current.Z}
			if world.IsWalkable(xNeighbor) {
				neighbors = append(neighbors, xNeighbor)
			}
		}

		if !world.IsWalkable(Point{current.X - 1, current.Y, current.Z}) {
			xNeighbor := Point{current.X - 1, current.Y + dy, current.Z}
			if world.IsWalkable(xNeighbor) {
				neighbors = append(neighbors, xNeighbor)
			}
		}

		if !world.IsWalkable(Point{current.X, current.Y, current.Z + 1}) {
			zNeighbor := Point{current.X, current.Y + dy, current.Z + 1}
			if world.IsWalkable(zNeighbor) {
				neighbors = append(neighbors, zNeighbor)
			}
		}

		if !world.IsWalkable(Point{current.X, current.Y, current.Z - 1}) {
			zNeighbor := Point{current.X, current.Y + dy, current.Z - 1}
			if world.IsWalkable(zNeighbor) {
				neighbors = append(neighbors, zNeighbor)
			}
		}
	}

	return neighbors
}

func jump(current Point, dx, dy, dz int, goal Point, world World) Point {
	next := Point{current.X + dx, current.Y + dy, current.Z + dz}

	if !world.IsWalkable(next) {
		return Point{-1, -1, -1}
	}

	if next.X == goal.X && next.Y == goal.Y && next.Z == goal.Z {
		return next
	}

	if dx != 0 && dz != 0 {

		if world.IsWalkable(Point{next.X - dx, next.Y, next.Z + dz}) &&
			!world.IsWalkable(Point{next.X - dx, next.Y, next.Z}) {
			return next
		}
		if world.IsWalkable(Point{next.X + dx, next.Y, next.Z - dz}) &&
			!world.IsWalkable(Point{next.X, next.Y, next.Z - dz}) {
			return next
		}

		hJump := jump(next, dx, 0, 0, goal, world)
		vJump := jump(next, 0, 0, dz, goal, world)
		if hJump.X != -1 || vJump.X != -1 {
			return next
		}
	} else if dx != 0 {

		if world.IsWalkable(Point{next.X, next.Y + 1, next.Z}) &&
			!world.IsWalkable(Point{current.X, current.Y + 1, current.Z}) {
			return next
		}
		if world.IsWalkable(Point{next.X, next.Y - 1, next.Z}) &&
			!world.IsWalkable(Point{current.X, current.Y - 1, current.Z}) {
			return next
		}
	} else if dz != 0 {

		if world.IsWalkable(Point{next.X, next.Y + 1, next.Z}) &&
			!world.IsWalkable(Point{current.X, current.Y + 1, current.Z}) {
			return next
		}
		if world.IsWalkable(Point{next.X, next.Y - 1, next.Z}) &&
			!world.IsWalkable(Point{current.X, current.Y - 1, current.Z}) {
			return next
		}
	} else if dy != 0 {

		if world.IsWalkable(Point{next.X + 1, next.Y, next.Z}) &&
			!world.IsWalkable(Point{current.X + 1, current.Y, current.Z}) {
			return next
		}
		if world.IsWalkable(Point{next.X - 1, next.Y, next.Z}) &&
			!world.IsWalkable(Point{current.X - 1, current.Y, current.Z}) {
			return next
		}

		if world.IsWalkable(Point{next.X, next.Y, next.Z + 1}) &&
			!world.IsWalkable(Point{current.X, current.Y, current.Z + 1}) {
			return next
		}
		if world.IsWalkable(Point{next.X, next.Y, next.Z - 1}) &&
			!world.IsWalkable(Point{current.X, current.Y, current.Z - 1}) {
			return next
		}
	}

	return jump(next, dx, dy, dz, goal, world)
}

func interpolatePath(from, to Point, world World) []Point {
	dx := sign(to.X - from.X)
	dy := sign(to.Y - from.Y)
	dz := sign(to.Z - from.Z)

	steps := max(max(abs(to.X-from.X), abs(to.Y-from.Y)), abs(to.Z-from.Z))

	points := []Point{}

	for i := 1; i <= steps; i++ {

		x := from.X + dx*i
		y := from.Y + dy*i
		z := from.Z + dz*i

		point := Point{x, y, z}

		if !world.IsWalkable(point) {

			altPoint := Point{x, from.Y, z}
			if world.IsWalkable(altPoint) {
				points = append(points, altPoint)
				continue
			}

			altPoint = Point{x, from.Y + 1, z}
			if world.IsWalkable(altPoint) {
				points = append(points, altPoint)
				continue
			}

			altPoint = Point{x, from.Y - 1, z}
			if world.IsWalkable(altPoint) {
				points = append(points, altPoint)
				continue
			}
		} else {
			points = append(points, point)
		}
	}

	return points
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

func manhattanDistance(a, b Point) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y) + abs(a.Z-b.Z)
}
