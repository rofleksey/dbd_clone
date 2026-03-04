package gamemap

import "math"

// Map dimensions: 80x80 units, Y is up
// Each unit = 1 meter approximately

const (
	MapWidth     = 80.0
	MapHeight    = 80.0
	WallHeight   = 3.0
	FloorY       = 0.0
	SecondFloorY = 3.5
)

type Vec3 struct {
	X, Y, Z float64
}

type AABB struct {
	MinX, MinY, MinZ float64
	MaxX, MaxY, MaxZ float64
}

func (a AABB) Contains(x, y, z float64) bool {
	return x >= a.MinX && x <= a.MaxX &&
		y >= a.MinY && y <= a.MaxY &&
		z >= a.MinZ && z <= a.MaxZ
}

func (a AABB) Intersects(b AABB) bool {
	return a.MinX <= b.MaxX && a.MaxX >= b.MinX &&
		a.MinY <= b.MaxY && a.MaxY >= b.MinY &&
		a.MinZ <= b.MaxZ && a.MaxZ >= b.MinZ
}

func (a AABB) IntersectsXZ(b AABB) bool {
	return a.MinX <= b.MaxX && a.MaxX >= b.MinX &&
		a.MinZ <= b.MaxZ && a.MaxZ >= b.MinZ
}

type Wall struct {
	Bounds AABB
	Height float64
}

type MapData struct {
	Walls          []Wall
	Generators     []GeneratorSpawn
	Pallets        []PalletSpawn
	Hooks          []HookSpawn
	Windows        []WindowSpawn
	ExitGates      []GateSpawn
	TrapSpawns     []Vec3
	SurvivorSpawns []Vec3
	KillerSpawn    Vec3
	// Stairs / ramps for second floor
	Stairs []AABB
}

type GeneratorSpawn struct {
	ID  string
	Pos Vec3
}

type PalletSpawn struct {
	ID   string
	Pos  Vec3
	RotY float64
}

type HookSpawn struct {
	ID  string
	Pos Vec3
}

type WindowSpawn struct {
	ID   string
	Pos  Vec3
	RotY float64
}

type GateSpawn struct {
	ID   string
	Pos  Vec3
	RotY float64
}

// CreateAzarovRealm creates the hardcoded Azarov's Realm map
func CreateAzarovRealm() *MapData {
	m := &MapData{}

	// ==============================
	// MAP BOUNDARY WALLS
	// ==============================
	boundaryThickness := 1.0
	m.addWall(-1, 0, -1, MapWidth+1, WallHeight, 0)                  // South
	m.addWall(-1, 0, MapHeight, MapWidth+1, WallHeight, MapHeight+1) // North
	m.addWall(-1, 0, 0, 0, WallHeight, MapHeight)                    // West
	m.addWall(MapWidth, 0, 0, MapWidth+1, WallHeight, MapHeight)     // East
	_ = boundaryThickness

	// ==============================
	// MAIN BUILDING (center-east, ~20x15, 2 floors)
	// Located at roughly (45,0,30) to (65,0,48)
	// ==============================
	mbX, mbZ := 45.0, 30.0
	mbW, mbD := 20.0, 18.0

	// Main building outer walls
	wallThk := 0.5
	// South wall with gap for door
	m.addWall(mbX, 0, mbZ, mbX+8, WallHeight, mbZ+wallThk)
	m.addWall(mbX+10, 0, mbZ, mbX+mbW, WallHeight, mbZ+wallThk)
	// North wall with gap for door
	m.addWall(mbX, 0, mbZ+mbD-wallThk, mbX+7, WallHeight, mbZ+mbD)
	m.addWall(mbX+9, 0, mbZ+mbD-wallThk, mbX+mbW, WallHeight, mbZ+mbD)
	// West wall
	m.addWall(mbX, 0, mbZ, mbX+wallThk, WallHeight, mbZ+mbD)
	// East wall with window gap
	m.addWall(mbX+mbW-wallThk, 0, mbZ, mbX+mbW, WallHeight, mbZ+6)
	m.addWall(mbX+mbW-wallThk, 0, mbZ+8, mbX+mbW, WallHeight, mbZ+mbD)

	// Internal wall dividing rooms (ground floor)
	m.addWall(mbX+10-wallThk/2, 0, mbZ, mbX+10+wallThk/2, WallHeight, mbZ+7)
	m.addWall(mbX+10-wallThk/2, 0, mbZ+9, mbX+10+wallThk/2, WallHeight, mbZ+mbD)

	// Second floor (full floor slab for collision)
	m.addWall(mbX, SecondFloorY-0.2, mbZ, mbX+mbW, SecondFloorY, mbZ+mbD)

	// Second floor walls (same outline but at second floor height)
	m.addWall(mbX, SecondFloorY, mbZ, mbX+mbW, SecondFloorY+WallHeight, mbZ+wallThk)
	m.addWall(mbX, SecondFloorY, mbZ+mbD-wallThk, mbX+6, SecondFloorY+WallHeight, mbZ+mbD)
	m.addWall(mbX+8, SecondFloorY, mbZ+mbD-wallThk, mbX+mbW, SecondFloorY+WallHeight, mbZ+mbD)
	m.addWall(mbX, SecondFloorY, mbZ, mbX+wallThk, SecondFloorY+WallHeight, mbZ+mbD)
	m.addWall(mbX+mbW-wallThk, SecondFloorY, mbZ, mbX+mbW, SecondFloorY+WallHeight, mbZ+mbD)

	// Stairs (ramp area going from ground to second floor) - in the west room
	m.Stairs = append(m.Stairs, AABB{
		MinX: mbX + 1, MinY: 0, MinZ: mbZ + 12,
		MaxX: mbX + 4, MaxY: SecondFloorY, MaxZ: mbZ + mbD - 1,
	})

	// ==============================
	// KILLER SHACK (northwest area, ~6x6)
	// Located at roughly (8,0,60) to (14,0,66)
	// ==============================
	skX, skZ := 8.0, 60.0
	skW, skD := 6.0, 6.0

	// Shack walls
	m.addWall(skX, 0, skZ, skX+skW, WallHeight, skZ+wallThk)             // South
	m.addWall(skX, 0, skZ+skD-wallThk, skX+3, WallHeight, skZ+skD)       // North left (door gap right)
	m.addWall(skX+4.5, 0, skZ+skD-wallThk, skX+skW, WallHeight, skZ+skD) // North right
	m.addWall(skX, 0, skZ, skX+wallThk, WallHeight, skZ+skD)             // West
	// East wall with window
	m.addWall(skX+skW-wallThk, 0, skZ, skX+skW, WallHeight, skZ+2)
	m.addWall(skX+skW-wallThk, 0, skZ+3.5, skX+skW, WallHeight, skZ+skD)

	// ==============================
	// HILL (south-center, a raised platform)
	// Located at roughly (35,0,10) to (45,0,18)
	// ==============================
	hillX, hillZ := 35.0, 10.0
	// Hill is a raised platform with slopes (simplified as a block)
	m.addWall(hillX, 0, hillZ, hillX+10, 2.0, hillZ+8)

	// ==============================
	// DEBRIS WALLS / LOOP WALLS (scattered around map)
	// These create the chase loops
	// ==============================

	// Loop 1 - L-shaped wall (southwest)
	m.addWall(12, 0, 15, 12.5, WallHeight, 22)
	m.addWall(12, 0, 15, 17, WallHeight, 15.5)

	// Loop 2 - T-wall (south-center-east)
	m.addWall(55, 0, 12, 55.5, WallHeight, 19)
	m.addWall(53, 0, 15, 58, WallHeight, 15.5)

	// Loop 3 - Straight wall (northwest)
	m.addWall(20, 0, 55, 20.5, WallHeight, 62)

	// Loop 4 - L-shape (northeast)
	m.addWall(60, 0, 58, 60.5, WallHeight, 65)
	m.addWall(60, 0, 58, 66, WallHeight, 58.5)

	// Loop 5 - Corner walls (center-west)
	m.addWall(18, 0, 35, 18.5, WallHeight, 42)
	m.addWall(18, 0, 38, 24, WallHeight, 38.5)

	// Loop 6 - Car/debris (center)
	m.addWall(32, 0, 38, 38, WallHeight*0.6, 41) // shorter wall like a car

	// Loop 7 - Wall near main building
	m.addWall(40, 0, 25, 40.5, WallHeight, 30)

	// Loop 8 - Wall near shack
	m.addWall(16, 0, 55, 16.5, WallHeight, 59)

	// Some extra debris for visual interest
	m.addWall(70, 0, 20, 72, WallHeight*0.5, 22)
	m.addWall(5, 0, 40, 7, WallHeight*0.5, 42)
	m.addWall(50, 0, 55, 52, WallHeight*0.5, 57)

	// ==============================
	// GENERATORS (7 total)
	// ==============================
	m.Generators = []GeneratorSpawn{
		{ID: "gen_0", Pos: Vec3{14, 0, 18}},            // Near loop 1
		{ID: "gen_1", Pos: Vec3{56, 0, 16}},            // Near loop 2
		{ID: "gen_2", Pos: Vec3{50, 0, 35}},            // Near main building entrance
		{ID: "gen_3", Pos: Vec3{55, SecondFloorY, 40}}, // Main building 2nd floor
		{ID: "gen_4", Pos: Vec3{10, 0, 63}},            // In shack
		{ID: "gen_5", Pos: Vec3{22, 0, 58}},            // Near shack area
		{ID: "gen_6", Pos: Vec3{65, 0, 62}},            // Northeast corner
	}

	// ==============================
	// PALLETS (15 total)
	// ==============================
	m.Pallets = []PalletSpawn{
		// Loop pallets
		{ID: "pallet_0", Pos: Vec3{14.5, 0, 18}, RotY: 0},           // Loop 1
		{ID: "pallet_1", Pos: Vec3{55.5, 0, 16}, RotY: math.Pi / 2}, // Loop 2
		{ID: "pallet_2", Pos: Vec3{20.5, 0, 58}, RotY: 0},           // Loop 3
		{ID: "pallet_3", Pos: Vec3{63, 0, 61}, RotY: math.Pi / 2},   // Loop 4
		{ID: "pallet_4", Pos: Vec3{21, 0, 39}, RotY: math.Pi / 2},   // Loop 5
		{ID: "pallet_5", Pos: Vec3{35, 0, 40}, RotY: 0},             // Loop 6
		{ID: "pallet_6", Pos: Vec3{40.5, 0, 27}, RotY: 0},           // Loop 7
		// Shack pallet
		{ID: "pallet_7", Pos: Vec3{11, 0, 65.5}, RotY: math.Pi / 2}, // Shack doorway
		// Main building pallets
		{ID: "pallet_8", Pos: Vec3{49, 0, 30}, RotY: math.Pi / 2},   // Main building south door
		{ID: "pallet_9", Pos: Vec3{52, 0, 47.5}, RotY: math.Pi / 2}, // Main building north door
		// Extra pallets
		{ID: "pallet_10", Pos: Vec3{30, 0, 25}, RotY: 0},
		{ID: "pallet_11", Pos: Vec3{70, 0, 40}, RotY: math.Pi / 2},
		{ID: "pallet_12", Pos: Vec3{10, 0, 45}, RotY: 0},
		{ID: "pallet_13", Pos: Vec3{45, 0, 60}, RotY: math.Pi / 2},
		{ID: "pallet_14", Pos: Vec3{25, 0, 10}, RotY: 0},
	}

	// ==============================
	// HOOKS (10 total, spread across map)
	// ==============================
	m.Hooks = []HookSpawn{
		{ID: "hook_0", Pos: Vec3{10, 0, 25}},
		{ID: "hook_1", Pos: Vec3{30, 0, 15}},
		{ID: "hook_2", Pos: Vec3{50, 0, 20}},
		{ID: "hook_3", Pos: Vec3{70, 0, 30}},
		{ID: "hook_4", Pos: Vec3{60, 0, 45}},
		{ID: "hook_5", Pos: Vec3{40, 0, 50}},
		{ID: "hook_6", Pos: Vec3{20, 0, 50}},
		{ID: "hook_7", Pos: Vec3{15, 0, 70}},
		{ID: "hook_8", Pos: Vec3{55, 0, 65}},
		{ID: "hook_9", Pos: Vec3{35, 0, 35}},
	}

	// ==============================
	// WINDOWS (8 total)
	// ==============================
	m.Windows = []WindowSpawn{
		// Main building windows
		{ID: "window_0", Pos: Vec3{65, 0, 37}, RotY: math.Pi / 2},            // East wall window
		{ID: "window_1", Pos: Vec3{55, 0, 39}, RotY: 0},                      // Internal wall window (ground)
		{ID: "window_2", Pos: Vec3{52, SecondFloorY, 48}, RotY: math.Pi / 2}, // 2nd floor north
		// Shack window
		{ID: "window_3", Pos: Vec3{14, 0, 62.5}, RotY: math.Pi / 2}, // Shack east wall
		// Map windows (freestanding in loops)
		{ID: "window_4", Pos: Vec3{12.5, 0, 19}, RotY: math.Pi / 2}, // Loop 1
		{ID: "window_5", Pos: Vec3{55.5, 0, 14}, RotY: 0},           // Loop 2
		{ID: "window_6", Pos: Vec3{60.5, 0, 62}, RotY: 0},           // Loop 4
		{ID: "window_7", Pos: Vec3{18.5, 0, 40}, RotY: 0},           // Loop 5
	}

	// ==============================
	// EXIT GATES (2, on opposite sides)
	// ==============================
	m.ExitGates = []GateSpawn{
		{ID: "gate_0", Pos: Vec3{0, 0, 40}, RotY: math.Pi / 2},   // West side
		{ID: "gate_1", Pos: Vec3{80, 0, 40}, RotY: -math.Pi / 2}, // East side
	}

	// ==============================
	// TRAP SPAWNS (8 locations)
	// ==============================
	m.TrapSpawns = []Vec3{
		{15, 0, 20},
		{50, 0, 18},
		{60, 0, 40},
		{35, 0, 45},
		{12, 0, 55},
		{25, 0, 65},
		{65, 0, 60},
		{40, 0, 15},
	}

	// ==============================
	// SPAWN POINTS
	// ==============================
	m.SurvivorSpawns = []Vec3{
		{5, 0, 10},
		{75, 0, 10},
		{75, 0, 70},
		{5, 0, 70},
		{40, 0, 5},
	}
	m.KillerSpawn = Vec3{40, 0, 40}

	return m
}

func (m *MapData) addWall(minX, minY, minZ, maxX, maxY, maxZ float64) {
	m.Walls = append(m.Walls, Wall{
		Bounds: AABB{MinX: minX, MinY: minY, MinZ: minZ, MaxX: maxX, MaxY: maxY, MaxZ: maxZ},
		Height: maxY - minY,
	})
}

// CheckCollision checks if a player AABB collides with any wall
func (m *MapData) CheckCollision(playerBounds AABB) bool {
	for _, wall := range m.Walls {
		if wall.Bounds.Intersects(playerBounds) {
			return true
		}
	}
	return false
}

// CheckCollisionXZ checks collision ignoring Y (for ground movement)
func (m *MapData) CheckCollisionXZ(playerBounds AABB) bool {
	for _, wall := range m.Walls {
		if wall.Bounds.IntersectsXZ(playerBounds) {
			// Also check Y overlap
			if playerBounds.MinY < wall.Bounds.MaxY && playerBounds.MaxY > wall.Bounds.MinY {
				return true
			}
		}
	}
	return false
}

// GetHeightAt returns the floor height at a given XZ position
func (m *MapData) GetHeightAt(x, z float64) float64 {
	// Check stairs
	for _, stair := range m.Stairs {
		if x >= stair.MinX && x <= stair.MaxX && z >= stair.MinZ && z <= stair.MaxZ {
			// Linear interpolation along Z for stair height
			t := (z - stair.MinZ) / (stair.MaxZ - stair.MinZ)
			return t * stair.MaxY
		}
	}

	// Check if on second floor (check if there's a floor slab beneath)
	// Second floor is at SecondFloorY
	return FloorY
}

// ResolveCollision pushes a player out of walls
func (m *MapData) ResolveCollision(x, z, radius float64, y float64) (float64, float64) {
	playerBounds := AABB{
		MinX: x - radius, MinY: y, MinZ: z - radius,
		MaxX: x + radius, MaxY: y + 1.8, MaxZ: z + radius,
	}

	for _, wall := range m.Walls {
		if !wall.Bounds.Intersects(playerBounds) {
			continue
		}

		// Push out along the axis of least penetration
		overlapX1 := playerBounds.MaxX - wall.Bounds.MinX
		overlapX2 := wall.Bounds.MaxX - playerBounds.MinX
		overlapZ1 := playerBounds.MaxZ - wall.Bounds.MinZ
		overlapZ2 := wall.Bounds.MaxZ - playerBounds.MinZ

		minOverlap := overlapX1
		pushX, pushZ := -overlapX1, 0.0

		if overlapX2 < minOverlap {
			minOverlap = overlapX2
			pushX, pushZ = overlapX2, 0.0
		}
		if overlapZ1 < minOverlap {
			minOverlap = overlapZ1
			pushX, pushZ = 0.0, -overlapZ1
		}
		if overlapZ2 < minOverlap {
			pushX, pushZ = 0.0, overlapZ2
		}

		x += pushX
		z += pushZ

		// Update player bounds
		playerBounds = AABB{
			MinX: x - radius, MinY: y, MinZ: z - radius,
			MaxX: x + radius, MaxY: y + 1.8, MaxZ: z + radius,
		}
	}

	// Clamp to map boundaries
	x = math.Max(radius, math.Min(MapWidth-radius, x))
	z = math.Max(radius, math.Min(MapHeight-radius, z))

	return x, z
}

// IsNear checks if position is within distance of a point
func IsNear(x1, z1, x2, z2, dist float64) bool {
	dx := x1 - x2
	dz := z1 - z2
	return dx*dx+dz*dz <= dist*dist
}

// Distance2D returns the XZ distance between two points
func Distance2D(x1, z1, x2, z2 float64) float64 {
	dx := x1 - x2
	dz := z1 - z2
	return math.Sqrt(dx*dx + dz*dz)
}
