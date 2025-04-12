package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/WillKirkmanM/paritone/internal/pathfinding"
	"github.com/WillKirkmanM/paritone/internal/world"
)

type PathRequest struct {
	StartX        int     `json:"startX"`
	StartY        int     `json:"startY"`
	StartZ        int     `json:"startZ"`
	EndX          int     `json:"endX"`
	EndY          int     `json:"endY"`
	EndZ          int     `json:"endZ"`
	Algorithm     string  `json:"algorithm"`
	AllowBreaking bool    `json:"allowBreaking"`
	AllowPlacing  bool    `json:"allowPlacing"`
	AvoidWater    bool    `json:"avoidWater"`
	MinVertical   bool    `json:"minimiseVertical"`
}

type PathResponse struct {
	Path            []pathfinding.Point `json:"path"`
	Error           string              `json:"error,omitempty"`
	ComputationTime int64               `json:"computationTime"`
	NodesExplored   int                 `json:"nodesExplored"`
	BlocksTraversed int                 `json:"blocksTraversed"`
	BlocksBroken    []pathfinding.Point `json:"blocksBroken"`
	BlocksPlaced    []pathfinding.Point `json:"blocksPlaced"`
	WaterCrossed    int                 `json:"waterCrossed"`
	VerticalChange  int                 `json:"verticalChange"`
	EstimatedTime   float64             `json:"estimatedTime"`
	TotalCost       float64             `json:"totalCost"`
}

func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler(w, r)
	}
}

func findPathHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Received path request: %+v\n", req)

	gameWorld := world.NewWorld()

	setupWorld(gameWorld, req)

	start := pathfinding.Point{X: req.StartX, Y: req.StartY, Z: req.StartZ}
	goal := pathfinding.Point{X: req.EndX, Y: req.EndY, Z: req.EndZ}

	gameWorld.SetBlock(start, world.Block{
		Type:      "air",
		Walkable:  true,
		Breakable: false,
		MoveCost:  1.0,
	})

	gameWorld.SetBlock(goal, world.Block{
		Type:      "air",
		Walkable:  true,
		Breakable: false,
		MoveCost:  1.0,
	})

	options := pathfinding.PathfindingOptions{
		AllowBreaking:  req.AllowBreaking,
		AllowPlacing:   req.AllowPlacing,
		AvoidWater:     req.AvoidWater,
		MinimiseHeight: req.MinVertical,
	}

	var result pathfinding.PathfindingResult

	fmt.Printf("Finding path from %v to %v using %s with options %+v\n", start, goal, req.Algorithm, options)

	switch req.Algorithm {
	case "bfs":
		result = pathfinding.FindPathBFSWithOptions(start, goal, gameWorld, options)
	case "dijkstra":
		result = pathfinding.FindPathDijkstraWithOptions(start, goal, gameWorld, options)
	default:
		result = pathfinding.FindPathWithOptions(start, goal, gameWorld, options)
	}

	response := PathResponse{
		Path:            result.Path,
		ComputationTime: result.ComputationTime.Milliseconds(),
		NodesExplored:   result.NodesExplored,
		BlocksTraversed: len(result.Path),
		BlocksBroken:    result.BlocksBroken,
		BlocksPlaced:    result.BlocksPlaced,
		WaterCrossed:    result.WaterCrossed,
		VerticalChange:  result.VerticalChange,
		EstimatedTime:   estimateTimeToTraverse(result),
		TotalCost:       result.TotalCost,
	}

	if len(result.Path) == 0 {
		response.Error = "No path found"
		fmt.Println("No path found")
	} else {
		fmt.Printf("Path found with %d steps\n", len(result.Path))
		printPathStats(result)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func compareAlgorithmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Received algorithm comparison request for %+v\n", req)

	gameWorld := world.NewWorld()
	setupWorld(gameWorld, req)

	start := pathfinding.Point{X: req.StartX, Y: req.StartY, Z: req.StartZ}
	goal := pathfinding.Point{X: req.EndX, Y: req.EndY, Z: req.EndZ}

	gameWorld.SetBlock(start, world.Block{
		Type:      "air",
		Walkable:  true,
		Breakable: false,
		MoveCost:  1.0,
	})

	gameWorld.SetBlock(goal, world.Block{
		Type:      "air",
		Walkable:  true,
		Breakable: false,
		MoveCost:  1.0,
	})

	options := pathfinding.PathfindingOptions{
		AllowBreaking:  req.AllowBreaking,
		AllowPlacing:   req.AllowPlacing,
		AvoidWater:     req.AvoidWater,
		MinimiseHeight: req.MinVertical,
	}

	astarResult := pathfinding.FindPathWithOptions(start, goal, gameWorld, options)
	dijkstraResult := pathfinding.FindPathDijkstraWithOptions(start, goal, gameWorld, options)
	bfsResult := pathfinding.FindPathBFSWithOptions(start, goal, gameWorld, options)

	type AlgorithmComparison struct {
		Algorithm       string                `json:"algorithm"`
		Path            []pathfinding.Point   `json:"path"`
		ComputationTime int64                 `json:"computationTime"`
		NodesExplored   int                   `json:"nodesExplored"`
		PathLength      int                   `json:"pathLength"`
		TotalCost       float64               `json:"totalCost"`
	}

	response := struct {
		AStar    AlgorithmComparison `json:"astar"`
		Dijkstra AlgorithmComparison `json:"dijkstra"`
		BFS      AlgorithmComparison `json:"bfs"`
	}{
		AStar: AlgorithmComparison{
			Algorithm:       "astar",
			Path:            astarResult.Path,
			ComputationTime: astarResult.ComputationTime.Milliseconds(),
			NodesExplored:   astarResult.NodesExplored,
			PathLength:      len(astarResult.Path),
			TotalCost:       astarResult.TotalCost,
		},
		Dijkstra: AlgorithmComparison{
			Algorithm:       "dijkstra",
			Path:            dijkstraResult.Path,
			ComputationTime: dijkstraResult.ComputationTime.Milliseconds(),
			NodesExplored:   dijkstraResult.NodesExplored,
			PathLength:      len(dijkstraResult.Path),
			TotalCost:       dijkstraResult.TotalCost,
		},
		BFS: AlgorithmComparison{
			Algorithm:       "bfs",
			Path:            bfsResult.Path,
			ComputationTime: bfsResult.ComputationTime.Milliseconds(),
			NodesExplored:   bfsResult.NodesExplored,
			PathLength:      len(bfsResult.Path),
			TotalCost:       bfsResult.TotalCost,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func setupWorld(gameWorld *world.World, req PathRequest) {
	minX, maxX := -20, 20
	minY, maxY := 0, 10
	minZ, maxZ := -20, 20

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: y, Z: z}, world.Block{
					Type:      "air",
					Walkable:  y > 0,
					Breakable: false,
					MoveCost:  1.0,
				})
			}
		}
	}

	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			gameWorld.SetBlock(pathfinding.Point{X: x, Y: 0, Z: z}, world.Block{
				Type:      "grass",
				Walkable:  false,
				Breakable: true,
				MoveCost:  1.0,
			})
		}
	}

	isMultiLevel := req.StartY != req.EndY

	if isMultiLevel {
		maxHeight := max(req.StartY, req.EndY)

		for level := 1; level <= maxHeight; level++ {
			startX := minX + (level-1)*8
			startZ := minZ + (level-1)*8

			for x := startX; x <= maxX; x++ {
				for z := startZ; z <= maxZ; z++ {
					for y := 1; y < level; y++ {
						gameWorld.SetBlock(pathfinding.Point{X: x, Y: y, Z: z}, world.Block{
							Type:      "stone",
							Walkable:  false,
							Breakable: true,
							MoveCost:  5.0,
						})
					}

					gameWorld.SetBlock(pathfinding.Point{X: x, Y: level, Z: z}, world.Block{
						Type:      "air",
						Walkable:  true,
						Breakable: false,
						MoveCost:  1.0,
					})
				}
			}

			if level > 1 {
				prevLevelX := startX - 8
				prevLevelZ := startZ - 8

				for i := 0; i < 8; i++ {
					rampX := prevLevelX + i
					rampZ := prevLevelZ + i

					gameWorld.SetBlock(pathfinding.Point{X: rampX, Y: level - 1, Z: rampZ}, world.Block{
						Type:      "wood",
						Walkable:  true,
						Breakable: false,
						MoveCost:  1.2,
					})

					for y := 1; y < level-1; y++ {
						gameWorld.SetBlock(pathfinding.Point{X: rampX, Y: y, Z: rampZ}, world.Block{
							Type:      "stone",
							Walkable:  false,
							Breakable: true,
							MoveCost:  5.0,
						})
					}
				}
			}
		}
	} else {

		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: 1, Z: z}, world.Block{
					Type:      "air",
					Walkable:  true,
					Breakable: false,
					MoveCost:  1.0,
				})
			}
		}

		for x := -5; x <= 5; x++ {
			for z := -5; z <= 5; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: 1, Z: z}, world.Block{
					Type:      "stone",
					Walkable:  false,
					Breakable: true,
					MoveCost:  5.0,
				})
			}
		}

		for x := 10; x <= 15; x++ {
			for z := 10; z <= 15; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: 1, Z: z}, world.Block{
					Type:      "water",
					Walkable:  true,
					Breakable: false,
					MoveCost:  3.0,
				})
			}
		}

		for x := -15; x <= -10; x++ {
			for z := 10; z <= 15; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: 1, Z: z}, world.Block{
					Type:      "sand",
					Walkable:  true,
					Breakable: true,
					MoveCost:  1.5,
				})
			}
		}

		for x := -15; x <= -10; x++ {
			for z := -15; z <= -10; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: 1, Z: z}, world.Block{
					Type:      "lava",
					Walkable:  false,
					Breakable: true,
					MoveCost:  10.0,
				})
			}
		}

		for x := 10; x <= 15; x++ {
			for z := -15; z <= -10; z++ {
				gameWorld.SetBlock(pathfinding.Point{X: x, Y: 1, Z: z}, world.Block{
					Type:      "ice",
					Walkable:  true,
					Breakable: true,
					MoveCost:  0.7,
				})
			}
		}
	}
}

func estimateTimeToTraverse(result pathfinding.PathfindingResult) float64 {
	timePerBlock := 0.25

	breakingTime := float64(len(result.BlocksBroken)) * 1.0

	placingTime := float64(len(result.BlocksPlaced)) * 0.5

	waterTime := float64(result.WaterCrossed) * 0.5

	verticalTime := float64(result.VerticalChange) * 0.2

	return float64(len(result.Path)) * timePerBlock + breakingTime + placingTime + waterTime + verticalTime
}

func printPathStats(result pathfinding.PathfindingResult) {
	fmt.Printf("Path length: %d blocks\n", len(result.Path))
	fmt.Printf("Computation time: %v\n", result.ComputationTime)
	fmt.Printf("Nodes explored: %d\n", result.NodesExplored)
	fmt.Printf("Blocks broken: %d\n", len(result.BlocksBroken))
	fmt.Printf("Blocks placed: %d\n", len(result.BlocksPlaced))
	fmt.Printf("Water blocks crossed: %d\n", result.WaterCrossed)
	fmt.Printf("Vertical change: %d blocks\n", result.VerticalChange)
	fmt.Printf("Estimated traversal time: %.2f seconds\n", estimateTimeToTraverse(result))
	fmt.Printf("Total path cost: %.2f\n", result.TotalCost)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println("Paritone Backend Starting...")

	http.HandleFunc("/api/find-path", enableCORS(findPathHandler))
	http.HandleFunc("/api/compare-algorithms", enableCORS(compareAlgorithmsHandler))

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	frontendPath := filepath.Join(workDir, "frontend")
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		frontendPath = filepath.Join(workDir, "..", "..", "frontend")
		if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
			log.Fatal("Frontend directory not found at:", frontendPath)
		}
	}

	fmt.Println("Serving frontend files from:", frontendPath)

	fs := http.FileServer(http.Dir(frontendPath))

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
			return
		}
		
		http.StripPrefix("/", fs).ServeHTTP(w, r)
	}))

	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}