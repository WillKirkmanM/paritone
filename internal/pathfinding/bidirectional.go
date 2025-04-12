package pathfinding

import (
	"container/list"
	"time"
)

func FindPathBidirectional(start, goal Point, world World) []Point {
	forwardQueue := list.New()
	backwardQueue := list.New()

	forwardQueue.PushBack(start)
	backwardQueue.PushBack(goal)

	forwardVisited := make(map[Point]bool)
	backwardVisited := make(map[Point]bool)
	forwardVisited[start] = true
	backwardVisited[goal] = true

	forwardParent := make(map[Point]Point)
	backwardParent := make(map[Point]Point)

	var meetingPoint Point
	meetFound := false

	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 && !meetFound {

		if !meetFound && forwardQueue.Len() > 0 {
			current := forwardQueue.Remove(forwardQueue.Front()).(Point)

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

				if world.IsWalkable(neighbor) && !forwardVisited[neighbor] {
					forwardQueue.PushBack(neighbor)
					forwardVisited[neighbor] = true
					forwardParent[neighbor] = current

					if backwardVisited[neighbor] {
						meetingPoint = neighbor
						meetFound = true
						break
					}
				}
			}
		}

		if !meetFound && backwardQueue.Len() > 0 {
			current := backwardQueue.Remove(backwardQueue.Front()).(Point)

			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

				if world.IsWalkable(neighbor) && !backwardVisited[neighbor] {
					backwardQueue.PushBack(neighbor)
					backwardVisited[neighbor] = true
					backwardParent[neighbor] = current

					if forwardVisited[neighbor] {
						meetingPoint = neighbor
						meetFound = true
						break
					}
				}
			}
		}
	}

	if meetFound {
		forwardPath := []Point{}
		for p := meetingPoint; p != start; {
			forwardPath = append([]Point{p}, forwardPath...)
			p = forwardParent[p]
		}
		forwardPath = append([]Point{start}, forwardPath...)

		backwardPath := []Point{}
		for p := backwardParent[meetingPoint]; p != goal; {
			backwardPath = append(backwardPath, p)
			p = backwardParent[p]
		}
		backwardPath = append(backwardPath, goal)

		completePath := append(forwardPath, backwardPath...)
		return completePath
	}

	return nil
}

func FindPathBidirectionalWithOptions(start, goal Point, world World, options PathfindingOptions) PathfindingResult {
	startTime := time.Now()

	forwardQueue := list.New()
	backwardQueue := list.New()

	forwardQueue.PushBack(start)
	backwardQueue.PushBack(goal)

	forwardVisited := make(map[Point]bool)
	backwardVisited := make(map[Point]bool)
	forwardVisited[start] = true
	backwardVisited[goal] = true

	forwardParent := make(map[Point]Point)
	backwardParent := make(map[Point]Point)

	nodesExplored := 0
	var blocksBroken []Point
	var blocksPlaced []Point
	breakPoints := make(map[Point]bool)
	placePoints := make(map[Point]bool)

	var meetingPoint Point
	meetFound := false

	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 && !meetFound {

		if !meetFound && forwardQueue.Len() > 0 {
			current := forwardQueue.Remove(forwardQueue.Front()).(Point)
			nodesExplored++

			neighbors := getNeighborsWithOptions(current, world, options, breakPoints, placePoints)

			for _, neighbor := range neighbors {
				if !forwardVisited[neighbor] {
					forwardQueue.PushBack(neighbor)
					forwardVisited[neighbor] = true
					forwardParent[neighbor] = current

					if backwardVisited[neighbor] {
						meetingPoint = neighbor
						meetFound = true
						break
					}
				}
			}
		}

		if !meetFound && backwardQueue.Len() > 0 {
			current := backwardQueue.Remove(backwardQueue.Front()).(Point)
			nodesExplored++

			neighbors := []Point{}
			for _, dir := range []Point{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
				neighbor := Point{current.X + dir.X, current.Y + dir.Y, current.Z + dir.Z}

				if world.IsWalkable(neighbor) {

					if !options.AvoidWater || world.GetBlockType(neighbor) != "water" {
						neighbors = append(neighbors, neighbor)
					}
				}
			}

			for _, neighbor := range neighbors {
				if !backwardVisited[neighbor] {
					backwardQueue.PushBack(neighbor)
					backwardVisited[neighbor] = true
					backwardParent[neighbor] = current

					if forwardVisited[neighbor] {
						meetingPoint = neighbor
						meetFound = true
						break
					}
				}
			}
		}
	}

	if meetFound {

		forwardPath := []Point{}
		for p := meetingPoint; !isEqual(p, start); {
			forwardPath = append([]Point{p}, forwardPath...)

			if breakPoints[p] {
				blocksBroken = append(blocksBroken, p)
			}

			if placePoints[p] {
				blocksPlaced = append(blocksPlaced, p)
			}

			p = forwardParent[p]
		}
		forwardPath = append([]Point{start}, forwardPath...)

		backwardPath := []Point{}
		for p := backwardParent[meetingPoint]; !isEqual(p, goal); {
			backwardPath = append(backwardPath, p)
			p = backwardParent[p]
		}
		backwardPath = append(backwardPath, goal)

		completePath := append(forwardPath, backwardPath...)

		waterCrossed := 0
		totalVerticalChange := 0
		lastY := start.Y

		for _, p := range completePath {
			if world.GetBlockType(p) == "water" {
				waterCrossed++
			}

			vertChange := abs(p.Y - lastY)
			totalVerticalChange += vertChange
			lastY = p.Y
		}

		return PathfindingResult{
			Path:            completePath,
			NodesExplored:   nodesExplored,
			ComputationTime: time.Since(startTime),
			BlocksBroken:    blocksBroken,
			BlocksPlaced:    blocksPlaced,
			WaterCrossed:    waterCrossed,
			VerticalChange:  totalVerticalChange,
			TotalCost:       float64(len(completePath)),
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

func getNeighborsWithOptions(current Point, world World, options PathfindingOptions,
	breakPoints, placePoints map[Point]bool) []Point {

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

	return neighbors
}
