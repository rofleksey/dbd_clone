package objects

import (
	"dbd-game/internal/gamemap"
)

type Generator struct {
	ID            string
	Pos           gamemap.Vec3
	Progress      float64 // 0.0 to 1.0
	Done          bool
	Regressing    bool
	BeingRepaired bool
	RepairerIDs   map[int]bool
}

func NewGenerator(id string, pos gamemap.Vec3) *Generator {
	return &Generator{
		ID:          id,
		Pos:         pos,
		RepairerIDs: make(map[int]bool),
	}
}

type Pallet struct {
	ID      string
	Pos     gamemap.Vec3
	RotY    float64
	Dropped bool
	Broken  bool
}

func NewPallet(id string, pos gamemap.Vec3, rotY float64) *Pallet {
	return &Pallet{
		ID:   id,
		Pos:  pos,
		RotY: rotY,
	}
}

// GetCollisionBounds returns the collision AABB when pallet is dropped
func (p *Pallet) GetCollisionBounds() gamemap.AABB {
	if !p.Dropped || p.Broken {
		return gamemap.AABB{}
	}

	// Pallet is ~2.5m wide, 0.3m thick, 1m high when dropped
	halfW := 1.25
	halfD := 0.15
	if p.RotY != 0 {
		halfW, halfD = halfD, halfW
	}

	return gamemap.AABB{
		MinX: p.Pos.X - halfW,
		MinY: 0,
		MinZ: p.Pos.Z - halfD,
		MaxX: p.Pos.X + halfW,
		MaxY: 1.0,
		MaxZ: p.Pos.Z + halfD,
	}
}

type Hook struct {
	ID         string
	Pos        gamemap.Vec3
	Occupied   bool
	PlayerID   int
	BrokenNext bool // Breaks after this use (not implemented for simplicity)
}

func NewHook(id string, pos gamemap.Vec3) *Hook {
	return &Hook{
		ID:  id,
		Pos: pos,
	}
}

type Trap struct {
	ID        string
	Pos       gamemap.Vec3
	Placed    bool
	Triggered bool
	OwnerID   int // Killer who placed it
}

func NewTrap(id string, pos gamemap.Vec3) *Trap {
	return &Trap{
		ID:     id,
		Pos:    pos,
		Placed: false,
	}
}

type ExitGate struct {
	ID       string
	Pos      gamemap.Vec3
	RotY     float64
	Progress float64
	Open     bool
	Powered  bool
}

func NewExitGate(id string, pos gamemap.Vec3, rotY float64) *ExitGate {
	return &ExitGate{
		ID:   id,
		Pos:  pos,
		RotY: rotY,
	}
}

type Window struct {
	ID   string
	Pos  gamemap.Vec3
	RotY float64
}

func NewWindow(id string, pos gamemap.Vec3, rotY float64) *Window {
	return &Window{
		ID:   id,
		Pos:  pos,
		RotY: rotY,
	}
}
