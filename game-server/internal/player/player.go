package player

import (
	"sync"
	"time"
)

const (
	HealthDead    = 0
	HealthDying   = 1
	HealthInjured = 2
	HealthHealthy = 3

	RoleSurvivor = "survivor"
	RoleKiller   = "killer"

	MoveIdle      = "idle"
	MoveWalking   = "walking"
	MoveRunning   = "running"
	MoveCrouching = "crouching"

	ActionNone           = "none"
	ActionRepairing      = "repairing"
	ActionHealing        = "healing"
	ActionUnhooking      = "unhooking"
	ActionOpeningGate    = "opening_gate"
	ActionCarrying       = "carrying"
	ActionBeingCarried   = "being_carried"
	ActionHooked         = "hooked"
	ActionTrapped        = "trapped"
	ActionStunned        = "stunned"
	ActionAttacking      = "attacking"
	ActionPlacingTrap    = "placing_trap"
	ActionPickingUpTrap  = "picking_up_trap"
	ActionBreakingPallet = "breaking_pallet"

	// Movement speeds (units per second)
	SurvivorWalkSpeed   = 2.26
	SurvivorRunSpeed    = 4.0
	SurvivorCrouchSpeed = 1.13
	KillerBaseSpeed     = 4.6

	// Interaction distances
	InteractDistance = 2.0
	AttackRange      = 2.5
	AttackWidth      = 1.5

	// Timings (seconds)
	RepairTime         = 80.0 // Full gen solo repair time
	HealTime           = 16.0 // Heal another survivor
	UnhookTime         = 1.5
	GateOpenTime       = 20.0
	AttackCooldown     = 3.0 // Successful hit cooldown
	AttackMissCooldown = 1.5
	PalletStunTime     = 2.5
	TrapEscapeTime     = 0.5 // Each attempt
	TrapEscapeChance   = 0.25
	PlaceTrapTime      = 2.0
	PickupTrapTime     = 1.0
	BreakPalletTime    = 2.5
	PickupSurvivorTime = 0.0 // Instant
	HookTime           = 1.5

	// Hook timings
	HookStage1Time   = 60.0 // Seconds before stage 2
	HookStage2Time   = 60.0 // Seconds before sacrifice (stage 3 = death)
	SelfUnhookChance = 0.10

	// Gen regression
	GenRegressionRate = 0.005 // Progress lost per second when regressing
	KickGenPause      = 1.5

	// Scratch marks
	ScratchMarkInterval = 0.3
	ScratchMarkDuration = 7.0

	// Blood trails
	BloodTrailInterval = 1.0
	BloodTrailDuration = 10.0

	// Dying state
	DyingBleedoutTime = 240.0 // 4 minutes to bleed out
	DyingRecoveryTime = 32.0  // Time for full recovery (need someone to pick up at 95%)

	PlayerRadius = 0.4
)

type Player struct {
	mu sync.RWMutex

	UserID   int
	Username string
	Role     string // survivor, killer

	// Position
	PosX, PosY, PosZ float64
	RotY             float64

	// State
	Health         int // 0-3
	MoveState      string
	ActionState    string
	ActionProgress float64
	ActionTarget   string // ID of object being interacted with

	// Carrying (killer only)
	CarryingPlayerID int

	// Hook state
	HookedOnID   string
	HookStage    int // 0=not hooked, 1, 2, 3=sacrificed
	HookTimer    float64
	HookAttempts int

	// Trap state
	TrappedInID  string
	TrapAttempts int

	// Killer specific
	TrapCount           int
	AttackCooldownTimer float64
	StunTimer           float64

	// Survivor specific
	DyingTimer       float64
	RecoveryProgress float64

	// Status
	IsAlive    bool
	HasEscaped bool

	// Stats
	Kills    int
	GensDone int

	// Network
	Ping         int
	LastPingTime time.Time

	// Scratch marks / blood
	LastScratchTime float64
	LastBloodTime   float64
}

func NewPlayer(userID int, username string, role string) *Player {
	p := &Player{
		UserID:      userID,
		Username:    username,
		Role:        role,
		Health:      HealthHealthy,
		MoveState:   MoveIdle,
		ActionState: ActionNone,
		IsAlive:     true,
		HasEscaped:  false,
	}

	if role == RoleKiller {
		p.TrapCount = 1
	}

	return p
}

func (p *Player) GetSpeed() float64 {
	if p.Role == RoleKiller {
		if p.CarryingPlayerID > 0 {
			return KillerBaseSpeed * 0.92 // Slightly slower when carrying
		}
		return KillerBaseSpeed
	}

	switch p.MoveState {
	case MoveRunning:
		if p.Health == HealthInjured {
			return SurvivorRunSpeed * 0.95
		}
		return SurvivorRunSpeed
	case MoveCrouching:
		return SurvivorCrouchSpeed
	case MoveWalking:
		return SurvivorWalkSpeed
	default:
		return 0
	}
}

func (p *Player) CanAct() bool {
	return p.IsAlive && !p.HasEscaped &&
		p.ActionState != ActionBeingCarried &&
		p.ActionState != ActionHooked &&
		p.ActionState != ActionStunned &&
		p.Health > HealthDying
}

func (p *Player) CanMove() bool {
	return p.CanAct() &&
		p.ActionState != ActionTrapped &&
		p.ActionState != ActionRepairing &&
		p.ActionState != ActionHealing &&
		p.ActionState != ActionUnhooking &&
		p.ActionState != ActionOpeningGate &&
		p.ActionState != ActionPlacingTrap &&
		p.ActionState != ActionPickingUpTrap &&
		p.ActionState != ActionBreakingPallet &&
		p.ActionState != ActionAttacking
}

func (p *Player) IsImmobile() bool {
	return p.ActionState == ActionBeingCarried ||
		p.ActionState == ActionHooked ||
		p.ActionState == ActionTrapped ||
		p.Health <= HealthDying ||
		p.HasEscaped
}
