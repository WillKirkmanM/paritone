package pathfinding

import (
	"container/heap"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func (p Point) IsEqual(other Point) bool {
	return p.X == other.X && p.Y == other.Y && p.Z == other.Z
}

type Node struct {
	Position Point
	GScore   float64
	FScore   float64
	Parent   *Node
	index    int
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].FScore < pq[j].FScore
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*Node)
	node.index = n
	*pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	node.index = -1
	*pq = old[0 : n-1]
	return node
}

func Heuristic(a, b Point, options PathfindingOptions) float64 {
	baseDist := float64(abs(a.X-b.X) + abs(a.Z-b.Z))

	verticalDist := float64(abs(a.Y - b.Y))
	if options.MinimiseHeight {
		verticalDist *= 2.0
	}

	return baseDist + verticalDist
}

func GetNeighbors(p Point, world World) []Point {
	directions := []Point{
		{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
	}

	var neighbors []Point
	for _, dir := range directions {
		newPoint := Point{p.X + dir.X, p.Y + dir.Y, p.Z + dir.Z}
		if world.IsWalkable(newPoint) {
			neighbors = append(neighbors, newPoint)
		}
	}
	return neighbors
}

type NeighborInfo struct {
	Point            Point
	RequiresBreaking bool
	RequiresPlacing  bool
}

func GetNeighborsWithBreaking(p Point, world World, options PathfindingOptions) []NeighborInfo {
	directions := []Point{
		{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1},
		{1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1},

		{0, 1, 0}, {0, -1, 0},
	}

	var neighbors []NeighborInfo

	for _, dir := range directions {
		newPoint := Point{p.X + dir.X, p.Y + dir.Y, p.Z + dir.Z}

		if world.IsWalkable(newPoint) {
			neighbors = append(neighbors, NeighborInfo{
				Point:            newPoint,
				RequiresBreaking: false,
				RequiresPlacing:  false,
			})
			continue
		}

		if options.AllowBreaking && world.CanBreak(newPoint) {
			neighbors = append(neighbors, NeighborInfo{
				Point:            newPoint,
				RequiresBreaking: true,
				RequiresPlacing:  false,
			})
		}
	}

	return neighbors
}

func FindPath(start, goal Point, world World) []Point {
	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: start,
		GScore:   0,
		FScore:   Heuristic(start, goal, PathfindingOptions{}),
		Parent:   nil,
	}

	heap.Push(openSet, startNode)

	cameFrom := make(map[Point]*Node)
	gScore := make(map[Point]float64)
	gScore[start] = 0

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)

		if current.Position.IsEqual(goal) {

			path := []Point{}
			for node := current; node != nil; node = node.Parent {
				path = append([]Point{node.Position}, path...)
			}
			return path
		}

		for _, neighbor := range GetNeighbors(current.Position, world) {
			tentativeGScore := gScore[current.Position] + 1

			if val, exists := gScore[neighbor]; !exists || tentativeGScore < val {
				neighborNode := &Node{
					Position: neighbor,
					GScore:   tentativeGScore,
					FScore:   tentativeGScore + Heuristic(neighbor, goal, PathfindingOptions{}),
					Parent:   current,
				}

				gScore[neighbor] = tentativeGScore
				cameFrom[neighbor] = current

				heap.Push(openSet, neighborNode)
			}
		}
	}

	return nil
}

func findPathWithBreaking(start, goal Point, world World, options PathfindingOptions) (
	[]Point, int, []Point, int, int, float64) {

	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: start,
		GScore:   0,
		FScore:   Heuristic(start, goal, options),
		Parent:   nil,
	}

	heap.Push(openSet, startNode)

	cameFrom := make(map[Point]*Node)
	gScore := make(map[Point]float64)
	gScore[start] = 0

	blocksBroken := make([]Point, 0)
	waterCrossed := 0
	nodesExplored := 0

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)
		nodesExplored++

		if current.Position.IsEqual(goal) {

			path := []Point{}
			totalVerticalChange := 0

			for node := current; node != nil; node = node.Parent {
				path = append([]Point{node.Position}, path...)
				if node.Parent != nil {

					vertChange := abs(node.Position.Y - node.Parent.Position.Y)
					totalVerticalChange += vertChange

					if world.GetBlockType(node.Position) == "water" {
						waterCrossed++
					}
				}
			}

			return path, nodesExplored, blocksBroken, waterCrossed, totalVerticalChange, current.GScore
		}

		neighbors := GetNeighborsWithBreaking(current.Position, world, options)

		for _, neighbor := range neighbors {

			moveCost := world.GetMovementCost(current.Position, neighbor.Point)

			if neighbor.RequiresBreaking {
				moveCost += 5.0
				blocksBroken = append(blocksBroken, neighbor.Point)
			}

			if options.AvoidWater && world.GetBlockType(neighbor.Point) == "water" {
				moveCost += 10.0
			}

			if options.MinimiseHeight && neighbor.Point.Y != current.Position.Y {
				moveCost += 2.0 * float64(abs(neighbor.Point.Y-current.Position.Y))
			}

			tentativeGScore := gScore[current.Position] + moveCost

			if val, exists := gScore[neighbor.Point]; !exists || tentativeGScore < val {
				neighborNode := &Node{
					Position: neighbor.Point,
					GScore:   tentativeGScore,
					FScore:   tentativeGScore + Heuristic(neighbor.Point, goal, options),
					Parent:   current,
				}

				gScore[neighbor.Point] = tentativeGScore
				cameFrom[neighbor.Point] = current

				heap.Push(openSet, neighborNode)
			}
		}
	}

	return nil, nodesExplored, blocksBroken, waterCrossed, 0, 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
