<p align="center">
  <img src="https://avatars.githubusercontent.com/u/138057124?s=200&v=4" width="150" />
</p>
<h1 align="center">Paritone</h1>

<p align="center">Paritone is a pathfinding tool that implements and compares multiple traversal algorithms including A*, Dijkstra, BFS, Bellman-Ford, Greedy Best-First Search, IDA*, and Jump Point Search in a 3D voxel environment.</p>

<h4 align="center">
  <a href="https://github.com/WillKirkmanM/paritone/releases">Releases</a> •
  <a href="#features">Features</a> •
  <a href="#algorithms">Algorithms</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a>
</h4>


## Features

| Feature | Description |
|---------|-------------|
| **Multiple Algorithms** | Choose from 7 different pathfinding algorithms |
| **3D Environment** | Visualise paths in a fully 3D voxel world |
| **Real-time Comparison** | Compare algorithm performance side by side |
| **Advanced Options** | Configure block breaking, placing, water avoidance, and height minimisation |
| **Algorithm Statistics** | View detailed metrics on path length, computation time, and resource usage |
| **Interactive Scenarios** | Test algorithms in different pre-configured environments |
| **Educational Information** | Learn about each algorithm's characteristics and use cases |

## Algorithms

Paritone implements the following pathfinding algorithms, each with different characteristics:

| Algorithm | Optimality | Speed | Memory Usage | Special Properties |
|-----------|------------|-------|--------------|-------------------|
| **A*** | ✓ Optimal | Fast | Medium | Balanced performance with heuristic guidance |
| **Dijkstra** | ✓ Optimal | Medium | High | Works with any edge weights, no heuristic |
| **BFS** | Optimal for uniform costs | Fast | High | Simple implementation, uniform step cost |
| **Greedy Best-First** | Not optimal | Very Fast | Medium | Uses only heuristic, ignores path cost |
| **Jump Point Search** | ✓ Optimal for grids | Very Fast | Low | Optimised A* for uniform grid maps |
| **IDA*** | ✓ Optimal | Varies | Very Low | Memory-efficient A* with iterative deepening |
| **Bellman-Ford** | ✓ Optimal | Slow | Medium | Can handle negative edge weights |

## Installation

### Method 1: Download Release

1. Download the latest release from the [Releases page](https://github.com/WillKirkmanM/paritone/releases)
2. Extract the zip file
3. Run the executable

### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/WillKirkmanM/paritone
cd paritone

# Build the application
go build -o paritone cmd/paritone/main.go

# Run the application
./paritone
```
