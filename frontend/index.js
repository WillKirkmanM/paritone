import * as THREE from "https://cdn.skypack.dev/three@0.132.2";
import { OrbitControls } from "https://cdn.skypack.dev/three@0.132.2/examples/jsm/controls/OrbitControls.js";

const scene = new THREE.Scene();
scene.background = new THREE.Color(0x87ceeb);

const camera = new THREE.PerspectiveCamera(
  75,
  window.innerWidth / window.innerHeight,
  0.1,
  1000
);
camera.position.set(30, 30, 30);
camera.lookAt(0, 0, 0);

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById("container").appendChild(renderer.domElement);

const controls = new OrbitControls(camera, renderer.domElement);
controls.enableDamping = true;
controls.dampingFactor = 0.05;

const grid = new THREE.GridHelper(100, 100);
scene.add(grid);

const world = {
  blocks: new Map(),
  setBlock: function (x, y, z, type) {
    const key = `${x},${y},${z}`;
    this.blocks.set(key, { type, mesh: null });
    this.updateBlockMesh(x, y, z);
  },
  getBlock: function (x, y, z) {
    const key = `${x},${y},${z}`;
    return this.blocks.get(key);
  },
  updateBlockMesh: function (x, y, z) {
    const key = `${x},${y},${z}`;
    const block = this.blocks.get(key);

    if (block.mesh) {
      scene.remove(block.mesh);
    }

    if (block.type === "air") return;

    let geometry = new THREE.BoxGeometry(1, 1, 1);
    let material;

    switch (block.type) {
      case "stone":
        material = new THREE.MeshLambertMaterial({ color: 0x888888 });
        break;
      case "grass":
        material = new THREE.MeshLambertMaterial({ color: 0x00aa00 });
        break;
      case "water":
        material = new THREE.MeshLambertMaterial({
          color: 0x0000ff,
          transparent: true,
          opacity: 0.7,
        });
        break;
      case "sand":
        material = new THREE.MeshLambertMaterial({ color: 0xf2d16b });
        break;
      case "wood":
        material = new THREE.MeshLambertMaterial({ color: 0x8b4513 });
        break;
      case "lava":
        material = new THREE.MeshLambertMaterial({
          color: 0xff4500,
          emissive: 0xff0000,
          emissiveIntensity: 0.5,
        });
        break;
      case "ice":
        material = new THREE.MeshLambertMaterial({
          color: 0xadd8e6,
          transparent: true,
          opacity: 0.8,
        });
        break;
      case "path":
        material = new THREE.MeshLambertMaterial({ color: 0xff0000 });
        break;
      case "start":
        material = new THREE.MeshLambertMaterial({ color: 0x00ff00 });
        break;
      case "goal":
        material = new THREE.MeshLambertMaterial({ color: 0x0000ff });
        break;
      case "break":
        material = new THREE.MeshLambertMaterial({
          color: 0xff4500,
          wireframe: true,
          opacity: 0.8,
          transparent: true,
        });
        break;
      case "place":
        material = new THREE.MeshLambertMaterial({
          color: 0x4caf50,
          wireframe: true,
          opacity: 0.6,
          transparent: true,
        });
        break;
      default:
        material = new THREE.MeshLambertMaterial({ color: 0xffffff });
    }

    const mesh = new THREE.Mesh(geometry, material);
    mesh.position.set(x, y, z);
    scene.add(mesh);
    block.mesh = mesh;
  },
  clearWorld: function () {
    for (const [key, block] of this.blocks.entries()) {
      if (block.mesh) {
        scene.remove(block.mesh);
      }
    }
    this.blocks.clear();
  },
};

const worldScenarios = {
  simple: {
    name: "Simple Obstacles",
    description: "Basic flat terrain with some stone obstacles",
    start: { x: -15, y: 1, z: -15 },
    goal: { x: 15, y: 1, z: 15 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "grass");
        }
      }

      for (let x = -5; x <= 5; x++) {
        for (let z = -5; z <= 5; z++) {
          world.setBlock(x, 1, z, "stone");
        }
      }

      for (let x = 10; x <= 15; x++) {
        for (let z = 10; z <= 15; z++) {
          world.setBlock(x, 1, z, "water");
        }
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
  maze: {
    name: "Complex Maze",
    description: "Navigate through a complex maze with narrow passages",
    start: { x: -18, y: 1, z: -18 },
    goal: { x: 18, y: 1, z: 18 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "grass");
        }
      }

      const mazePattern = [
        "XXXXXXXXXXXXXXXXXXXXXXXXX",
        "XS        X             X",
        "XXXXXXXX  X  XXXXXXXXXXX",
        "X         X  X           ",
        "X  XXXXXXXX  X  XXXXXXXXX",
        "X  X         X  X       X",
        "X  X  XXXXXXXX  X  XXX  X",
        "X  X  X         X  X X  X",
        "X  X  X  XXXXXXXX  X X  X",
        "X  X  X  X         X X  X",
        "X  X  X  X  XXXXXXXX X  X",
        "X  X  X  X  X        X  X",
        "X  X  X  X  X  XXXXXX   X",
        "X  X  X  X  X  X     XXXX",
        "X  X  X  X  X  X  X     X",
        "X  X  X  X  X  X  XXXX  X",
        "X  X  X  X  X  X     X  X",
        "X  X  X  X  X  XXXXX X  X",
        "X  X  X  X  X        X  X",
        "X  X  X  X  XXXXXXXXXX  X",
        "X  X  X  X              X",
        "X  X  X  XXXXXXXXXXXXXX X",
        "X  X  X                 X",
        "X  X  XXXXXXXXXXXXXXXXXXX",
        "X  X                    G",
        "XXXXXXXXXXXXXXXXXXXXXXXXXXE",
      ];

      for (let z = 0; z < mazePattern.length; z++) {
        for (let x = 0; x < mazePattern[z].length; x++) {
          const xPos = x - 12;
          const zPos = z - 12;

          if (mazePattern[z][x] === "X") {
            world.setBlock(xPos, 1, zPos, "stone");
          }
        }
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
  multilevel: {
    name: "Multi-Level Terrain",
    description: "Navigate across different elevations",
    start: { x: -18, y: 1, z: -18 },
    goal: { x: 18, y: 5, z: 18 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "grass");
        }
      }

      for (let x = -20; x < -5; x++) {
        for (let z = -20; z < -5; z++) {
          world.setBlock(x, 1, z, "air");
        }
      }

      for (let x = -5; x <= 20; x++) {
        for (let z = -20; z < -5; z++) {
          world.setBlock(x, 1, z, "stone");
          world.setBlock(x, 2, z, "air");
        }
      }

      for (let x = -20; x < -5; x++) {
        for (let z = -5; z <= 20; z++) {
          world.setBlock(x, 1, z, "stone");
          world.setBlock(x, 2, z, "stone");
          world.setBlock(x, 3, z, "air");
        }
      }

      for (let x = -5; x < 5; x++) {
        for (let z = -5; z < 5; z++) {
          world.setBlock(x, 1, z, "stone");
          world.setBlock(x, 2, z, "stone");
          world.setBlock(x, 3, z, "stone");
          world.setBlock(x, 4, z, "air");
        }
      }

      for (let x = 5; x <= 20; x++) {
        for (let z = 5; z <= 20; z++) {
          world.setBlock(x, 1, z, "stone");
          world.setBlock(x, 2, z, "stone");
          world.setBlock(x, 3, z, "stone");
          world.setBlock(x, 4, z, "stone");
          world.setBlock(x, 5, z, "air");
        }
      }

      for (let i = 0; i < 5; i++) {
        world.setBlock(-5 - i, 1, -15, "wood");
      }

      for (let i = 0; i < 5; i++) {
        world.setBlock(-10, 2, -5 - i, "wood");
      }

      for (let i = 0; i < 5; i++) {
        world.setBlock(-5 - i, 3, 0, "wood");
      }

      for (let i = 0; i < 5; i++) {
        world.setBlock(0, 4, 5 + i, "wood");
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
  mixedMaterials: {
    name: "Mixed Materials Challenge",
    description:
      "Navigate through different materials with varying traversal costs",
    start: { x: -18, y: 1, z: 0 },
    goal: { x: 18, y: 1, z: 0 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "grass");
        }
      }

      for (let x = -20; x <= 20; x++) {
        for (let z = -5; z <= 5; z++) {
          world.setBlock(x, 1, z, "air");
        }
      }

      for (let x = -15; x < -5; x++) {
        for (let z = -3; z <= 3; z++) {
          world.setBlock(x, 1, z, "sand");
        }
      }

      for (let x = -5; x < 5; x++) {
        for (let z = -4; z <= 4; z++) {
          world.setBlock(x, 1, z, "water");
        }
      }

      for (let x = 5; x < 15; x++) {
        for (let z = -3; z <= 3; z++) {
          world.setBlock(x, 1, z, "ice");
        }
      }

      for (let z = -5; z <= 5; z++) {
        world.setBlock(-10, 1, z, "lava");
        world.setBlock(10, 1, z, "lava");
      }

      for (let i = -2; i <= 2; i++) {
        world.setBlock(-10, 1, i, "wood");
        world.setBlock(10, 1, i, "wood");
      }

      for (let x = -5; x <= 5; x++) {
        if (x % 2 === 0) {
          world.setBlock(x, 1, -6, "stone");
          world.setBlock(x, 1, 6, "stone");
        }
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
  islands: {
    name: "Islands Challenge",
    description: "Find your way across disconnected islands",
    start: { x: -18, y: 1, z: -18 },
    goal: { x: 18, y: 1, z: 18 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "water");
        }
      }

      const islands = [
        { x: -18, z: -18, radius: 3 },
        { x: -10, z: -12, radius: 2 },
        { x: -5, z: -5, radius: 2 },
        { x: 0, z: 0, radius: 3 },
        { x: 8, z: 5, radius: 2 },
        { x: 12, z: 12, radius: 2 },
        { x: 18, z: 18, radius: 3 },
      ];

      for (const island of islands) {
        for (let x = -island.radius; x <= island.radius; x++) {
          for (let z = -island.radius; z <= island.radius; z++) {
            if (x * x + z * z <= island.radius * island.radius) {
              world.setBlock(island.x + x, 0, island.z + z, "grass");
              world.setBlock(island.x + x, 1, island.z + z, "air");
            }
          }
        }
      }

      const bridges = [
        { x1: -15, z1: -15, x2: -10, z2: -12 },
        { x1: -8, z1: -10, x2: -5, z2: -5 },
        { x1: -3, z1: -3, x2: 0, z2: 0 },
        { x1: 3, z1: 3, x2: 8, z2: 5 },
        { x1: 10, z1: 7, x2: 12, z2: 12 },
        { x1: 14, z1: 14, x2: 18, z2: 18 },
      ];

      for (const bridge of bridges) {
        const dx = bridge.x2 - bridge.x1;
        const dz = bridge.z2 - bridge.z1;
        const steps = Math.max(Math.abs(dx), Math.abs(dz));

        for (let i = 0; i <= steps; i++) {
          const x = Math.floor(bridge.x1 + (dx * i) / steps);
          const z = Math.floor(bridge.z1 + (dz * i) / steps);
          world.setBlock(x, 0, z, "wood");
          world.setBlock(x, 1, z, "air");
        }
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
  algorithmComparison: {
    name: "Algorithm Comparison",
    description:
      "Scenario designed to showcase the differences between pathfinding algorithms",
    start: { x: -18, y: 1, z: 0 },
    goal: { x: 18, y: 1, z: 0 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "grass");
        }
      }

      for (let x = -18; x <= 18; x++) {
        for (let z = -5; z <= 5; z++) {
          world.setBlock(x, 1, z, "air");
        }
      }

      for (let x = -15; x <= 15; x++) {
        world.setBlock(x, 1, 0, "water");
      }

      for (let x = -15; x <= 15; x++) {
        world.setBlock(x, 1, 3, "ice");
        world.setBlock(x, 1, -3, "ice");
      }

      for (let z = -3; z <= 3; z++) {
        world.setBlock(-15, 1, z, "ice");
        world.setBlock(15, 1, z, "ice");
      }

      for (let i = -10; i <= 10; i += 5) {
        world.setBlock(i, 1, 0, "stone");
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
  algorithmShowcase: {
    name: "Algorithm Showcase",
    description:
      "Complex scenario to showcase differences between pathfinding algorithms",
    start: { x: -18, y: 1, z: 0 },
    goal: { x: 18, y: 1, z: 0 },
    create: function () {
      for (let x = -20; x <= 20; x++) {
        for (let z = -20; z <= 20; z++) {
          world.setBlock(x, 0, z, "grass");
        }
      }

      for (let x = -18; x <= 18; x++) {
        for (let z = -10; z <= 10; z++) {
          world.setBlock(x, 1, z, "air");
        }
      }

      for (let x = -15; x <= 15; x++) {
        world.setBlock(x, 1, 0, "water");
      }

      for (let x = -15; x <= 15; x++) {
        world.setBlock(x, 1, 5, "ice");
        world.setBlock(x, 1, -5, "ice");
      }

      for (let z = -5; z <= 5; z++) {
        world.setBlock(-15, 1, z, "ice");
        world.setBlock(15, 1, z, "ice");
      }

      for (let i = -10; i <= 10; i += 5) {
        world.setBlock(i, 1, 0, "stone");
      }

      for (let x = -12; x <= 12; x += 3) {
        for (let z = 2; z <= 4; z++) {
          world.setBlock(x, 1, z, "stone");
        }
        for (let z = -4; z <= -2; z++) {
          world.setBlock(x + 1, 1, z, "stone");
        }
      }

      for (let x = -8; x <= 8; x++) {
        if (x % 4 != 0) {
          world.setBlock(x, 1, 2, "stone");
          world.setBlock(x, 1, -2, "stone");
        }
      }

      for (let x = -5; x <= 5; x++) {
        for (let z = -1; z <= 1; z++) {
          world.setBlock(x, 2, z, "wood");
          if (x % 2 == 0 && z == 0) {
            world.setBlock(x, 3, z, "wood");
          }
        }
      }

      world.setBlock(this.start.x, this.start.y, this.start.z, "start");
      world.setBlock(this.goal.x, this.goal.y, this.goal.z, "goal");
    },
  },
};

const ambientLight = new THREE.AmbientLight(0xffffff, 0.5);
scene.add(ambientLight);

const directionalLight = new THREE.DirectionalLight(0xffffff, 0.5);
directionalLight.position.set(10, 20, 10);
scene.add(directionalLight);

function animate() {
  requestAnimationFrame(animate);
  controls.update();
  renderer.render(scene, camera);
}

function populateAlgorithmDropdown() {
  const algorithmSelect = document.getElementById("algorithm");

  algorithmSelect.innerHTML = "";

  const algorithms = [
    { id: "astar", name: "A*", description: "Balanced speed and optimality" },
    { id: "dijkstra", name: "Dijkstra", description: "Always optimal, slower" },
    { id: "bfs", name: "BFS", description: "Simple breadth-first search" },
    {
      id: "greedy",
      name: "Greedy Best-First",
      description: "Fast but suboptimal",
    },
    {
      id: "jps",
      name: "Jump Point Search",
      description: "Optimised for grid maps",
    },
    { id: "ida", name: "IDA*", description: "Memory efficient A*" },
    {
      id: "bellmanford",
      name: "Bellman-Ford",
      description: "Handles negative costs",
    },
  ];

  algorithms.forEach((algorithm) => {
    const option = document.createElement("option");
    option.value = algorithm.id;
    option.textContent = `${algorithm.name} - ${algorithm.description}`;
    algorithmSelect.appendChild(option);
  });

  algorithmSelect.value = "astar";

  algorithmSelect.dispatchEvent(new Event("change"));
}

document.getElementById("algorithm").addEventListener("change", function () {
  const algorithm = this.value;
  const advancedOptions = document.getElementById("advancedOptions");
  const heuristicOptions = document.getElementById("heuristicOptions");
  const weightOption = document.getElementById("weightOption");
  const iterationsOption = document.getElementById("iterationsOption");

  if (["astar", "greedy", "jps", "ida"].includes(algorithm)) {
    advancedOptions.style.display = "block";

    heuristicOptions.style.display = "block";

    weightOption.style.display = ["astar", "greedy"].includes(algorithm)
      ? "block"
      : "none";

    iterationsOption.style.display = algorithm === "ida" ? "block" : "none";
  } else {
    advancedOptions.style.display = "none";
  }
});

async function findPath(startX, startY, startZ, endX, endY, endZ) {
  try {
    const algorithm = document.getElementById("algorithm").value;
    const allowBreaking = document.getElementById("allowBreaking").checked;
    const allowPlacing = document.getElementById("allowPlacing").checked;
    const avoidWater = document.getElementById("avoidWater").checked;
    const minimiseVertical =
      document.getElementById("minimiseVertical").checked;

    const heuristicType =
      document.getElementById("heuristicType")?.value || "manhattan";
    const heuristicWeight = parseFloat(
      document.getElementById("heuristicWeight")?.value || 1.0
    );
    const maxIterations = parseInt(
      document.getElementById("maxIterations")?.value || 1000
    );
    const jumpPointOptimisation = algorithm === "jps";

    clearPathVisualisation();

    document.getElementById("stats").style.display = "block";
    document.getElementById("pathInfo").style.display = "block";
    document.getElementById("pathInfo").textContent = "Finding path...";

    updateStats({
      path: [],
      computationTime: 0,
      nodesExplored: 0,
      blocksTraversed: 0,
      blocksBroken: [],
      blocksPlaced: [],
      waterCrossed: 0,
      verticalChange: 0,
      estimatedTime: 0,
      totalCost: 0,
    });

    console.log(
      `Requesting path from (${startX},${startY},${startZ}) to (${endX},${endY},${endZ}) using ${algorithm}`
    );
    console.log(
      `Options: breaking=${allowBreaking}, placing=${allowPlacing}, avoidWater=${avoidWater}, minimiseVertical=${minimiseVertical}`
    );
    console.log(
      `Advanced options: heuristic=${heuristicType}, weight=${heuristicWeight}, maxIterations=${maxIterations}`
    );

    const response = await fetch("/api/find-path", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        startX: startX,
        startY: startY,
        startZ: startZ,
        endX: endX,
        endY: endY,
        endZ: endZ,
        algorithm: algorithm,
        allowBreaking: allowBreaking,
        allowPlacing: allowPlacing,
        avoidWater: avoidWater,
        minimiseVertical: minimiseVertical,
        heuristicType: heuristicType,
        heuristicWeight: heuristicWeight,
        maxIterations: maxIterations,
        jumpPointOptimisation: jumpPointOptimisation,
      }),
    });

    const data = await response.json();
    console.log("Response from backend:", data);

    if (data.error) {
      console.error("Pathfinding error:", data.error);
      document.getElementById("pathInfo").textContent = `Error: ${data.error}`;
      return null;
    }

    if (!data.path || data.path.length === 0) {
      console.error("No path returned from backend");
      document.getElementById("pathInfo").textContent = "No path found";
      return null;
    }

    document.getElementById(
      "pathInfo"
    ).textContent = `Path found! ${data.path.length} blocks`;

    const algorithmNames = {
      astar: "A*",
      dijkstra: "Dijkstra",
      bfs: "BFS",
      greedy: "Greedy Best-First",
      jps: "Jump Point Search",
      ida: "IDA*",
      bellmanford: "Bellman-Ford",
    };
    document.getElementById("stat-algorithm").textContent = `Algorithm: ${
      algorithmNames[algorithm] || algorithm
    }`;

    const optionsText = [];
    if (allowBreaking) optionsText.push("Break blocks");
    if (allowPlacing) optionsText.push("Place blocks");
    if (avoidWater) optionsText.push("Avoid water");
    if (minimiseVertical) optionsText.push("Minimise climbing");

    if (["astar", "greedy", "jps", "ida"].includes(algorithm)) {
      optionsText.push(`Heuristic: ${heuristicType}`);

      if (["astar", "greedy"].includes(algorithm)) {
        optionsText.push(`Weight: ${heuristicWeight}`);
      }

      if (algorithm === "ida") {
        optionsText.push(`Max Iterations: ${maxIterations}`);
      }
    }

    document.getElementById("stat-options").textContent = `Options: ${
      optionsText.join(", ") || "None"
    }`;

    updateStats(data);

    const convertedPath = data.path.map((point) => ({
      x: point.X !== undefined ? point.X : point.x,
      y: point.Y !== undefined ? point.Y : point.y,
      z: point.Z !== undefined ? point.Z : point.z,
    }));

    const blocksToBreak = (data.blocksBroken || []).map((point) => ({
      x: point.X !== undefined ? point.X : point.x,
      y: point.Y !== undefined ? point.Y : point.y,
      z: point.Z !== undefined ? point.Z : point.z,
    }));

    const blocksToPlace = (data.blocksPlaced || []).map((point) => ({
      x: point.X !== undefined ? point.X : point.x,
      y: point.Y !== undefined ? point.Y : point.y,
      z: point.Z !== undefined ? point.Z : point.z,
    }));

    console.log("Path:", convertedPath);
    console.log("Blocks to break:", blocksToBreak);
    console.log("Blocks to place:", blocksToPlace);

    return {
      path: convertedPath,
      blocksToBreak: blocksToBreak,
      blocksToPlace: blocksToPlace,
    };
  } catch (error) {
    console.error("Failed to connect to backend:", error);
    document.getElementById(
      "pathInfo"
    ).textContent = `Connection error: ${error.message}`;

    console.log("Using fallback pathfinding...");
    const start = { x: startX, y: startY, z: startZ };
    const end = { x: endX, y: endY, z: endZ };

    const path = [];
    let current = { ...start };

    while (current.x !== end.x || current.z !== end.z) {
      if (current.x < end.x) current.x++;
      else if (current.x > end.x) current.x--;
      else if (current.z < end.z) current.z++;
      else if (current.z > end.z) current.z--;

      path.push({ ...current });
    }

    return {
      path: path,
      blocksToBreak: [],
    };
  }
}

function updateStats(data) {
  document.getElementById("stat-path-length").textContent = `Length: ${
    data.path ? data.path.length : 0
  } blocks`;
  document.getElementById(
    "stat-computation-time"
  ).textContent = `Computation time: ${data.computationTime || 0} ms`;
  document.getElementById(
    "stat-nodes-explored"
  ).textContent = `Nodes explored: ${data.nodesExplored || 0}`;

  document.getElementById(
    "stat-blocks-traveled"
  ).textContent = `Blocks traveled: ${data.blocksTraversed || 0}`;
  document.getElementById("stat-blocks-broken").textContent = `Blocks broken: ${
    data.blocksBroken ? data.blocksBroken.length : 0
  }`;
  document.getElementById("stat-blocks-placed").textContent = `Blocks placed: ${
    data.blocksPlaced ? data.blocksPlaced.length : 0
  }`;
  document.getElementById(
    "stat-water-crossed"
  ).textContent = `Water blocks crossed: ${data.waterCrossed || 0}`;
  document.getElementById(
    "stat-vertical-distance"
  ).textContent = `Vertical distance: ${data.verticalChange || 0} blocks`;

  document.getElementById(
    "stat-time-estimate"
  ).textContent = `Time estimate: ${(data.estimatedTime || 0).toFixed(
    2
  )} seconds`;
  document.getElementById("stat-total-cost").textContent = `Total cost: ${(
    data.totalCost || 0
  ).toFixed(2)}`;
}

function visualisePath(pathData) {
  if (!pathData) return;

  const { path, blocksToBreak } = pathData;
  console.log("Visualising path with", path.length, "points");

  clearPathVisualisation();

  for (const point of path) {
    world.setBlock(point.x, point.y, point.z, "path");
  }

  for (const point of blocksToBreak) {
    world.setBlock(point.x, point.y, point.z, "break");
  }

  document.getElementById("resetWorld").onclick = function () {
    clearPathVisualisation();

    document.getElementById("stats").style.display = "none";
    document.getElementById("pathInfo").style.display = "none";

    const selectedScenario = document.getElementById("scenarioSelector").value;
    loadScenario(selectedScenario);
  };
}

function clearPathVisualisation() {
  for (const [key, block] of world.blocks.entries()) {
    if (
      block.type === "path" ||
      block.type === "break" ||
      block.type === "place"
    ) {
      const [x, y, z] = key.split(",").map(Number);

      world.setBlock(x, y, z, "air");
    }
  }
}

function createScenarioUI() {
  const scenarioSelector = document.getElementById("scenarioSelector");

  while (scenarioSelector.firstChild) {
    scenarioSelector.removeChild(scenarioSelector.firstChild);
  }

  for (const [key, scenario] of Object.entries(worldScenarios)) {
    const option = document.createElement("option");
    option.value = key;
    option.textContent = scenario.name;
    scenarioSelector.appendChild(option);
  }

  scenarioSelector.addEventListener("change", function () {
    const selectedScenario = this.value;
    loadScenario(selectedScenario);
  });

  scenarioSelector.value = Object.keys(worldScenarios)[0];
  loadScenario(scenarioSelector.value);

  document.getElementById("resetWorld").onclick = function () {
    clearPathVisualisation();

    const selectedScenario = document.getElementById("scenarioSelector").value;
    loadScenario(selectedScenario);

    document.getElementById("stats").style.display = "none";
    document.getElementById("pathInfo").style.display = "none";

    for (const [key, block] of world.blocks.entries()) {
      if (block.type === "break" || block.type === "place") {
        const [x, y, z] = key.split(",").map(Number);
        world.setBlock(x, y, z, "air");
      }
    }

    console.log("World reset and path cleared");
  };
}

function loadScenario(scenarioKey) {
  const scenario = worldScenarios[scenarioKey];
  if (!scenario) return;

  document.getElementById("scenarioDescription").textContent =
    scenario.description;

  world.clearWorld();

  scenario.create();

  const pathButton = document.getElementById("startPathfinding");
  pathButton.onclick = async function () {
    const start = scenario.start;
    const goal = scenario.goal;
    const pathData = await findPath(
      start.x,
      start.y,
      start.z,
      goal.x,
      goal.y,
      goal.z
    );
    if (pathData) {
      visualisePath(pathData);
    }
  };
}

document.getElementById("algorithm").addEventListener("change", function () {
  const algorithm = this.value;
  const advancedOptions = document.getElementById("advancedOptions");
  const weightContainer = document.getElementById("weightContainer");
  const iterationsContainer = document.getElementById("iterationsContainer");

  if (["astar", "greedy", "jps", "ida"].includes(algorithm)) {
    advancedOptions.style.display = "block";

    weightContainer.style.display = ["astar", "greedy"].includes(algorithm)
      ? "block"
      : "none";

    iterationsContainer.style.display = algorithm === "ida" ? "block" : "none";
  } else {
    advancedOptions.style.display = "none";
  }
});

document
  .getElementById("heuristicWeight")
  .addEventListener("input", function () {
    document.getElementById("heuristicWeightValue").textContent = this.value;
  });

document.getElementById("algorithm").addEventListener("change", function () {
  const algorithm = this.value;
  const infoDiv = document.getElementById("algorithmInfo");
  const descriptions = {
    astar: {
      description:
        "A* combines Dijkstra's algorithm with a heuristic to guide search toward the goal. It's efficient and guaranteed to find the optimal path when using an admissible heuristic.",
      timeComplexity: "O(E + V log V) with binary heap",
      spaceComplexity: "O(V) - stores all nodes",
      optimality: "Optimal (with admissible heuristic)",
      useCases:
        "General pathfinding problems with reasonable memory constraints",
    },
    dijkstra: {
      description:
        "Dijkstra's algorithm finds the shortest path by exploring nodes in order of increasing distance from the start. It doesn't use a heuristic.",
      timeComplexity: "O(E + V log V) with binary heap",
      spaceComplexity: "O(V) - stores all nodes",
      optimality: "Always optimal",
      useCases: "When optimality is critical and there are weighted edges",
    },
    bfs: {
      description:
        "Breadth-First Search explores all nodes at the current depth before moving deeper. Fast for unweighted graphs but not optimal for weighted ones.",
      timeComplexity: "O(V + E)",
      spaceComplexity: "O(V) - stores all nodes at current level",
      optimality: "Optimal only for unweighted graphs",
      useCases: "Simple unweighted pathfinding, maze solving",
    },
    greedy: {
      description:
        "Greedy Best-First Search only considers the heuristic distance to goal, ignoring path cost. Very fast but often suboptimal.",
      timeComplexity: "O(E + V log V) with binary heap",
      spaceComplexity: "O(V) - stores nodes",
      optimality: "Not guaranteed to be optimal",
      useCases: "When speed is critical and path quality is secondary",
    },
    jps: {
      description:
        "Jump Point Search optimizes A* for uniform grid maps by skipping symmetric paths, dramatically reducing nodes expanded.",
      timeComplexity:
        "O(E log V) but typically much faster than A* in practice",
      spaceComplexity: "O(V) - but explores fewer nodes than A*",
      optimality: "Optimal for uniform cost grids",
      useCases: "Grid-based games with uniform costs and few obstacles",
    },
    ida: {
      description:
        "Iterative Deepening A* performs depth-first searches with increasing cost limits, using minimal memory.",
      timeComplexity:
        "O(b^d) where b is branching factor and d is solution depth",
      spaceComplexity: "O(d) - linear in path depth",
      optimality: "Optimal (with admissible heuristic)",
      useCases: "Memory-constrained environments where optimality matters",
    },
    bellmanford: {
      description:
        "Bellman-Ford algorithm can handle negative edge weights and detect negative cycles, but is slower than Dijkstra's.",
      timeComplexity: "O(V*E) - polynomial time",
      spaceComplexity: "O(V) - stores distances",
      optimality: "Optimal even with negative weights (if no negative cycles)",
      useCases: "When there are negative edge weights or costs",
    },
  };

  if (descriptions[algorithm]) {
    document.getElementById("algorithmDescription").textContent =
      descriptions[algorithm].description;
    document.getElementById("timeComplexity").textContent =
      descriptions[algorithm].timeComplexity;
    document.getElementById("spaceComplexity").textContent =
      descriptions[algorithm].spaceComplexity;
    document.getElementById("optimality").textContent =
      descriptions[algorithm].optimality;
    document.getElementById("useCases").textContent =
      descriptions[algorithm].useCases;
    infoDiv.style.display = "block";
  } else {
    infoDiv.style.display = "none";
  }
});

document.getElementById("algorithm").dispatchEvent(new Event("change"));

createScenarioUI();
populateAlgorithmDropdown();
animate();
