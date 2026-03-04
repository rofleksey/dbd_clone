package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"dbd-game/internal/gamemap"
	"dbd-game/internal/objects"
	"dbd-game/internal/player"
	"dbd-game/internal/protocol"
)

const (
	TickRate                   = 30
	TickDuration               = time.Second / TickRate
	GameTimeout                = 15 * 60 // 15 minutes in seconds
	GensRequired               = 5
	TotalGens                  = 7
	ProgressReportInterval     = 5 * time.Second
	HookSurvivorHeightOffset   = 1.35 // World Y offset from hook base so survivor appears on the hook
)

type Game struct {
	mu sync.RWMutex

	ID        int
	MasterURL string
	KillerID  int

	Players    map[int]*player.Player
	Map        *gamemap.MapData
	Generators []*objects.Generator
	Pallets    []*objects.Pallet
	Hooks      []*objects.Hook
	Traps      []*objects.Trap
	Gates      []*objects.ExitGate
	Windows    []*objects.Window

	ScratchMarks []protocol.ScratchMark
	BloodTrails  []protocol.BloodTrail

	Tick          int64
	ElapsedTime   float64
	GensCompleted int
	GatesPowered  bool
	GameOver      bool
	Result        string

	// Broadcast channel
	broadcast chan []byte
	// Player connections
	connections map[int]chan []byte

	startTime          time.Time
	lastProgressReport time.Time
}

func NewGame(id int, masterURL string, killerID int) *Game {
	mapData := gamemap.CreateAzarovRealm()

	g := &Game{
		ID:          id,
		MasterURL:   masterURL,
		KillerID:    killerID,
		Players:     make(map[int]*player.Player),
		Map:         mapData,
		broadcast:   make(chan []byte, 256),
		connections: make(map[int]chan []byte),
	}

	// Initialize generators
	for _, gs := range mapData.Generators {
		g.Generators = append(g.Generators, objects.NewGenerator(gs.ID, gs.Pos))
	}

	// Initialize pallets
	for _, ps := range mapData.Pallets {
		g.Pallets = append(g.Pallets, objects.NewPallet(ps.ID, ps.Pos, ps.RotY))
	}

	// Initialize hooks
	for _, hs := range mapData.Hooks {
		g.Hooks = append(g.Hooks, objects.NewHook(hs.ID, hs.Pos))
	}

	// Initialize traps (not placed yet, in world at spawn positions)
	for i, ts := range mapData.TrapSpawns {
		trap := objects.NewTrap(fmt.Sprintf("trap_%d", i), ts)
		trap.Placed = true // Traps start placed at spawn positions for pickup
		g.Traps = append(g.Traps, trap)
	}

	// Initialize exit gates
	for _, gs := range mapData.ExitGates {
		g.Gates = append(g.Gates, objects.NewExitGate(gs.ID, gs.Pos, gs.RotY))
	}

	// Initialize windows
	for _, ws := range mapData.Windows {
		g.Windows = append(g.Windows, objects.NewWindow(ws.ID, ws.Pos, ws.RotY))
	}

	return g
}

func (g *Game) AddPlayer(userID int, username string, role string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	p := player.NewPlayer(userID, username, role)

	// Set spawn position
	if role == player.RoleKiller {
		p.PosX = g.Map.KillerSpawn.X
		p.PosY = g.Map.KillerSpawn.Y
		p.PosZ = g.Map.KillerSpawn.Z
	} else {
		spawnIdx := len(g.Players) % len(g.Map.SurvivorSpawns)
		spawn := g.Map.SurvivorSpawns[spawnIdx]
		p.PosX = spawn.X
		p.PosY = spawn.Y
		p.PosZ = spawn.Z
	}

	g.Players[userID] = p

	// Create send channel
	g.connections[userID] = make(chan []byte, 256)
}

func (g *Game) RemovePlayer(userID int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.Players, userID)
	if ch, ok := g.connections[userID]; ok {
		close(ch)
		delete(g.connections, userID)
	}

	// If anyone disconnects, end the game
	if !g.GameOver {
		g.GameOver = true
		g.Result = "disconnected"
	}
}

func (g *Game) GetSendChan(userID int) chan []byte {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.connections[userID]
}

func (g *Game) HandleInput(userID int, msg protocol.ClientMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()

	p, ok := g.Players[userID]
	if !ok || g.GameOver {
		return
	}

	switch msg.Type {
	case "move":
		g.handleMove(p, msg)
	case "interact":
		g.handleInteract(p, msg)
	case "stop_interact":
		g.handleStopInteract(p)
	case "ping":
		p.Ping = int(time.Since(p.LastPingTime).Milliseconds())
		p.LastPingTime = time.Now()
	}
}

func (g *Game) handleMove(p *player.Player, msg protocol.ClientMessage) {
	if !p.CanMove() {
		return
	}

	p.MoveState = msg.State
	p.RotY = msg.RotY

	// Validate and apply movement
	speed := p.GetSpeed()
	dt := 1.0 / float64(TickRate)

	// Calculate desired position from client input
	dx := msg.PosX - p.PosX
	dz := msg.PosZ - p.PosZ
	dist := math.Sqrt(dx*dx + dz*dz)

	maxDist := speed * dt * 1.5 // Allow some tolerance
	if dist > maxDist && dist > 0.01 {
		// Clamp movement to max allowed
		scale := maxDist / dist
		dx *= scale
		dz *= scale
	}

	newX := p.PosX + dx
	newZ := p.PosZ + dz

	// Get height at position
	newY := g.Map.GetHeightAt(newX, newZ)

	// Resolve collision
	newX, newZ = g.Map.ResolveCollision(newX, newZ, player.PlayerRadius, newY)

	// Check dropped pallet collision
	for _, pal := range g.Pallets {
		if pal.Dropped && !pal.Broken {
			bounds := pal.GetCollisionBounds()
			pBounds := gamemap.AABB{
				MinX: newX - player.PlayerRadius, MinY: newY,
				MinZ: newZ - player.PlayerRadius,
				MaxX: newX + player.PlayerRadius, MaxY: newY + 1.8,
				MaxZ: newZ + player.PlayerRadius,
			}
			if bounds.Intersects(pBounds) {
				// Don't move into pallet
				newX = p.PosX
				newZ = p.PosZ
				break
			}
		}
	}

	p.PosX = newX
	p.PosY = newY
	p.PosZ = newZ

	// Check trap collision (survivors only)
	if p.Role == player.RoleSurvivor && p.ActionState != player.ActionTrapped {
		for _, trap := range g.Traps {
			if trap.Placed && !trap.Triggered {
				if gamemap.IsNear(p.PosX, p.PosZ, trap.Pos.X, trap.Pos.Z, 0.5) {
					// Stepped in trap!
					trap.Triggered = true
					p.ActionState = player.ActionTrapped
					p.TrappedInID = trap.ID
					p.TrapAttempts = 0
					p.MoveState = player.MoveIdle
				}
			}
		}
	}
}

func (g *Game) handleInteract(p *player.Player, msg protocol.ClientMessage) {
	switch msg.Action {
	case "repair":
		g.handleRepair(p, msg.Target)
	case "heal":
		g.handleHeal(p, msg.Target)
	case "unhook":
		g.handleUnhook(p, msg.Target)
	case "open_gate":
		g.handleOpenGate(p, msg.Target)
	case "drop_pallet":
		g.handleDropPallet(p, msg.Target)
	case "vault":
		g.handleVault(p, msg.Target)
	case "attack":
		g.handleAttack(p)
	case "pickup":
		g.handlePickupSurvivor(p)
	case "hook":
		g.handleHookSurvivor(p, msg.Target)
	case "place_trap":
		g.handlePlaceTrap(p)
	case "pickup_trap":
		g.handlePickupTrap(p, msg.Target)
	case "break_pallet":
		g.handleBreakPallet(p, msg.Target)
	case "kick_gen":
		g.handleKickGen(p, msg.Target)
	case "self_unhook":
		g.handleSelfUnhook(p)
	case "escape_trap":
		g.handleEscapeTrap(p)
	case "drop_survivor":
		g.handleDropSurvivor(p)
	}
}

func (g *Game) handleStopInteract(p *player.Player) {
	if p.ActionState == player.ActionRepairing {
		// Stop repairing
		for _, gen := range g.Generators {
			delete(gen.RepairerIDs, p.UserID)
			if len(gen.RepairerIDs) == 0 {
				gen.BeingRepaired = false
			}
		}
	}

	if p.ActionState == player.ActionRepairing ||
		p.ActionState == player.ActionHealing ||
		p.ActionState == player.ActionOpeningGate ||
		p.ActionState == player.ActionPlacingTrap ||
		p.ActionState == player.ActionPickingUpTrap {
		p.ActionState = player.ActionNone
		p.ActionProgress = 0
		p.ActionTarget = ""
	}
}

func (g *Game) handleRepair(p *player.Player, targetID string) {
	if p.Role != player.RoleSurvivor || !p.CanAct() {
		return
	}

	gen := g.findGenerator(targetID)
	if gen == nil || gen.Done {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, gen.Pos.X, gen.Pos.Z, player.InteractDistance) {
		return
	}

	p.ActionState = player.ActionRepairing
	p.ActionTarget = targetID
	p.MoveState = player.MoveIdle
	p.RotY = math.Atan2(gen.Pos.X-p.PosX, gen.Pos.Z-p.PosZ)
	gen.RepairerIDs[p.UserID] = true
	gen.BeingRepaired = true
	gen.Regressing = false
}

func (g *Game) handleHeal(p *player.Player, targetIDStr string) {
	if p.Role != player.RoleSurvivor || !p.CanAct() {
		return
	}

	// Find target player
	targetID := 0
	fmt.Sscanf(targetIDStr, "%d", &targetID)
	target, ok := g.Players[targetID]
	if !ok || target.Health != player.HealthInjured || target.UserID == p.UserID {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, target.PosX, target.PosZ, player.InteractDistance) {
		return
	}

	p.ActionState = player.ActionHealing
	p.ActionTarget = targetIDStr
	p.ActionProgress = 0
	p.MoveState = player.MoveIdle
}

func (g *Game) handleUnhook(p *player.Player, targetHookID string) {
	if p.Role != player.RoleSurvivor || !p.CanAct() {
		return
	}

	hook := g.findHook(targetHookID)
	if hook == nil || !hook.Occupied {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, hook.Pos.X, hook.Pos.Z, player.InteractDistance) {
		return
	}

	p.ActionState = player.ActionUnhooking
	p.ActionTarget = targetHookID
	p.ActionProgress = 0
	p.MoveState = player.MoveIdle
}

func (g *Game) handleSelfUnhook(p *player.Player) {
	// Survivors cannot self-unhook
	return
}

func (g *Game) handleEscapeTrap(p *player.Player) {
	// Survivors cannot self-untrap
	return
}

func (g *Game) handleOpenGate(p *player.Player, targetID string) {
	if p.Role != player.RoleSurvivor || !p.CanAct() || !g.GatesPowered {
		return
	}

	gate := g.findGate(targetID)
	if gate == nil || !gate.Powered || gate.Open {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, gate.Pos.X, gate.Pos.Z, player.InteractDistance+0.6) {
		return
	}

	p.ActionState = player.ActionOpeningGate
	p.ActionTarget = targetID
	p.MoveState = player.MoveIdle
}

func (g *Game) handleDropPallet(p *player.Player, targetID string) {
	if p.Role != player.RoleSurvivor || !p.CanAct() {
		return
	}

	pallet := g.findPallet(targetID)
	if pallet == nil || pallet.Dropped || pallet.Broken {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, pallet.Pos.X, pallet.Pos.Z, player.InteractDistance) {
		return
	}

	pallet.Dropped = true

	// Check if killer is in the pallet zone - stun!
	for _, other := range g.Players {
		if other.Role == player.RoleKiller {
			if gamemap.IsNear(other.PosX, other.PosZ, pallet.Pos.X, pallet.Pos.Z, 1.5) {
				other.ActionState = player.ActionStunned
				other.StunTimer = player.PalletStunTime

				// If carrying, drop the survivor
				if other.CarryingPlayerID > 0 {
					g.dropCarriedSurvivor(other)
				}
			}
		}
	}
}

func (g *Game) handleVault(p *player.Player, targetID string) {
	if !p.CanAct() {
		return
	}

	window := g.findWindow(targetID)
	if window == nil {
		// Check if it's a dropped pallet vault
		pallet := g.findPallet(targetID)
		if pallet != nil && pallet.Dropped && !pallet.Broken {
			if gamemap.IsNear(p.PosX, p.PosZ, pallet.Pos.X, pallet.Pos.Z, player.InteractDistance) {
				// Teleport to other side of pallet
				g.vaultOver(p, pallet.Pos.X, pallet.Pos.Z, pallet.RotY)
			}
		}
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, window.Pos.X, window.Pos.Z, player.InteractDistance) {
		return
	}

	g.vaultOver(p, window.Pos.X, window.Pos.Z, window.RotY)
}

func (g *Game) vaultOver(p *player.Player, vaultX, vaultZ, rotY float64) {
	// Move player to other side
	offset := 2.0
	dx := math.Sin(rotY) * offset
	dz := math.Cos(rotY) * offset

	// Determine which side the player is on
	playerDx := p.PosX - vaultX
	playerDz := p.PosZ - vaultZ
	dot := playerDx*math.Sin(rotY) + playerDz*math.Cos(rotY)

	if dot > 0 {
		p.PosX = vaultX - dx
		p.PosZ = vaultZ - dz
	} else {
		p.PosX = vaultX + dx
		p.PosZ = vaultZ + dz
	}
}

func (g *Game) handleAttack(p *player.Player) {
	if p.Role != player.RoleKiller || p.AttackCooldownTimer > 0 || p.ActionState == player.ActionStunned {
		return
	}
	if p.CarryingPlayerID > 0 {
		return
	}

	p.ActionState = player.ActionAttacking
	p.ActionProgress = 0

	// Check for hits
	hit := false
	attackX := p.PosX + math.Sin(p.RotY)*player.AttackRange/2
	attackZ := p.PosZ + math.Cos(p.RotY)*player.AttackRange/2

	for _, target := range g.Players {
		if target.Role == player.RoleSurvivor && target.IsAlive && !target.HasEscaped &&
			target.ActionState != player.ActionBeingCarried &&
			target.Health > player.HealthDying {

			if gamemap.IsNear(attackX, attackZ, target.PosX, target.PosZ, player.AttackWidth) {
				// Hit!
				target.Health--
				hit = true

				if target.Health == player.HealthDying {
					target.MoveState = player.MoveIdle
					target.ActionState = player.ActionDying
					target.DyingTimer = player.DyingBleedoutTime
				}

				break // Only hit one survivor per attack
			}
		}
	}

	if hit {
		p.AttackCooldownTimer = player.AttackCooldown
	} else {
		p.AttackCooldownTimer = player.AttackMissCooldown
	}
}

func (g *Game) handlePickupSurvivor(p *player.Player) {
	if p.Role != player.RoleKiller || p.CarryingPlayerID > 0 {
		return
	}

	// Find nearest dying survivor (not hooked)
	for _, target := range g.Players {
		if target.Role == player.RoleSurvivor && target.Health == player.HealthDying &&
			target.IsAlive && target.ActionState != player.ActionBeingCarried &&
			target.ActionState != player.ActionHooked {

			if gamemap.IsNear(p.PosX, p.PosZ, target.PosX, target.PosZ, player.InteractDistance) {
				p.CarryingPlayerID = target.UserID
				p.ActionState = player.ActionCarrying
				target.ActionState = player.ActionBeingCarried
				break
			}
		}
	}
}

func (g *Game) handleDropSurvivor(p *player.Player) {
	if p.Role != player.RoleKiller || p.CarryingPlayerID == 0 {
		return
	}
	g.dropCarriedSurvivor(p)
}

func (g *Game) dropCarriedSurvivor(killer *player.Player) {
	if carried, ok := g.Players[killer.CarryingPlayerID]; ok {
		carried.ActionState = player.ActionNone
		carried.PosX = killer.PosX
		carried.PosZ = killer.PosZ
		carried.PosY = killer.PosY
	}
	killer.CarryingPlayerID = 0
	if killer.ActionState == player.ActionCarrying {
		killer.ActionState = player.ActionNone
	}
}

func (g *Game) handleHookSurvivor(p *player.Player, hookID string) {
	if p.Role != player.RoleKiller || p.CarryingPlayerID == 0 {
		return
	}

	hook := g.findHook(hookID)
	if hook == nil || hook.Occupied {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, hook.Pos.X, hook.Pos.Z, player.InteractDistance) {
		return
	}

	carried, ok := g.Players[p.CarryingPlayerID]
	if !ok {
		return
	}

	// Hook the survivor
	hook.Occupied = true
	hook.PlayerID = carried.UserID

	carried.ActionState = player.ActionHooked
	carried.HookedOnID = hookID
	carried.HookStage++
	if carried.HookStage >= 3 {
		// Sacrificed
		carried.IsAlive = false
		carried.Health = player.HealthDead
		hook.Occupied = false
		hook.PlayerID = 0
		p.Kills++
	} else {
		carried.HookTimer = player.HookStage1Time
		carried.PosX = hook.Pos.X
		carried.PosY = hook.Pos.Y + HookSurvivorHeightOffset
		carried.PosZ = hook.Pos.Z
	}

	p.CarryingPlayerID = 0
	p.ActionState = player.ActionNone
}

func (g *Game) handlePlaceTrap(p *player.Player) {
	if p.Role != player.RoleKiller || p.TrapCount <= 0 {
		return
	}
	if p.ActionState == player.ActionPlacingTrap {
		return
	}

	p.ActionState = player.ActionPlacingTrap
	p.ActionProgress = 0
	p.MoveState = player.MoveIdle
}

func (g *Game) handlePickupTrap(p *player.Player, trapID string) {
	if p.Role != player.RoleKiller {
		return
	}

	trap := g.findTrap(trapID)
	if trap == nil || !trap.Placed {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, trap.Pos.X, trap.Pos.Z, player.InteractDistance) {
		return
	}

	trap.Placed = false
	p.TrapCount++
}

func (g *Game) handleBreakPallet(p *player.Player, palletID string) {
	if p.Role != player.RoleKiller {
		return
	}

	pallet := g.findPallet(palletID)
	if pallet == nil || !pallet.Dropped || pallet.Broken {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, pallet.Pos.X, pallet.Pos.Z, player.InteractDistance) {
		return
	}

	p.ActionState = player.ActionBreakingPallet
	p.ActionTarget = palletID
	p.ActionProgress = 0
	p.MoveState = player.MoveIdle
}

func (g *Game) handleKickGen(p *player.Player, genID string) {
	if p.Role != player.RoleKiller {
		return
	}

	gen := g.findGenerator(genID)
	if gen == nil || gen.Done || gen.Progress <= 0 {
		return
	}

	if !gamemap.IsNear(p.PosX, p.PosZ, gen.Pos.X, gen.Pos.Z, player.InteractDistance) {
		return
	}

	gen.Regressing = true
	gen.BeingRepaired = false
}

// unhookPlayer frees a survivor from a hook
func (g *Game) unhookPlayer(p *player.Player) {
	hook := g.findHook(p.HookedOnID)
	if hook != nil {
		hook.Occupied = false
		hook.PlayerID = 0
	}

	p.ActionState = player.ActionNone
	p.HookedOnID = ""
	p.HookTimer = 0
	p.Health = player.HealthInjured
}

// Lookup helpers
func (g *Game) findGenerator(id string) *objects.Generator {
	for _, gen := range g.Generators {
		if gen.ID == id {
			return gen
		}
	}
	return nil
}

func (g *Game) findPallet(id string) *objects.Pallet {
	for _, p := range g.Pallets {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) findHook(id string) *objects.Hook {
	for _, h := range g.Hooks {
		if h.ID == id {
			return h
		}
	}
	return nil
}

func (g *Game) findTrap(id string) *objects.Trap {
	for _, t := range g.Traps {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (g *Game) findGate(id string) *objects.ExitGate {
	for _, gate := range g.Gates {
		if gate.ID == id {
			return gate
		}
	}
	return nil
}

func (g *Game) findWindow(id string) *objects.Window {
	for _, w := range g.Windows {
		if w.ID == id {
			return w
		}
	}
	return nil
}

// Run starts the main game loop
func (g *Game) Run() {
	g.startTime = time.Now()
	g.lastProgressReport = time.Now()

	ticker := time.NewTicker(TickDuration)
	defer ticker.Stop()

	log.Printf("Game %d started with tick rate %d", g.ID, TickRate)

	for range ticker.C {
		g.mu.Lock()

		if g.GameOver {
			g.mu.Unlock()
			g.endGame()
			return
		}

		g.update()
		state := g.buildState()
		g.mu.Unlock()

		// Broadcast state to all players
		data, _ := json.Marshal(protocol.ServerMessage{
			Type:    "state",
			Payload: state,
		})

		g.mu.RLock()
		for _, ch := range g.connections {
			select {
			case ch <- data:
			default:
				// Drop frame if client is slow
			}
		}
		g.mu.RUnlock()

		// Report progress to master periodically
		if time.Since(g.lastProgressReport) > ProgressReportInterval {
			g.reportProgress()
			g.lastProgressReport = time.Now()
		}
	}
}

func (g *Game) update() {
	dt := 1.0 / float64(TickRate)
	g.Tick++
	g.ElapsedTime += dt

	// Check game timeout
	if g.ElapsedTime >= GameTimeout {
		g.GameOver = true
		g.Result = "timeout"
		return
	}

	// Update player timers
	for _, p := range g.Players {
		// Attack cooldown
		if p.AttackCooldownTimer > 0 {
			p.AttackCooldownTimer -= dt
			if p.AttackCooldownTimer <= 0 {
				p.AttackCooldownTimer = 0
				if p.ActionState == player.ActionAttacking {
					p.ActionState = player.ActionNone
				}
			}
		}

		// Stun timer
		if p.StunTimer > 0 {
			p.StunTimer -= dt
			if p.StunTimer <= 0 {
				p.StunTimer = 0
				p.ActionState = player.ActionNone
			}
		}

		// Hook timer
		if p.ActionState == player.ActionHooked {
			p.HookTimer -= dt
			if p.HookTimer <= 0 {
				p.HookStage++
				if p.HookStage >= 3 {
					// Sacrificed
					p.IsAlive = false
					p.Health = player.HealthDead
					p.ActionState = player.ActionNone
					hook := g.findHook(p.HookedOnID)
					if hook != nil {
						hook.Occupied = false
						hook.PlayerID = 0
					}
					// Credit kill to killer
					for _, k := range g.Players {
						if k.Role == player.RoleKiller {
							k.Kills++
							break
						}
					}
				} else {
					p.HookTimer = player.HookStage2Time
				}
			}
		}

		// Dying timer (bleedout)
		if p.Health == player.HealthDying && p.ActionState != player.ActionBeingCarried && p.ActionState != player.ActionHooked {
			p.ActionState = player.ActionDying
			p.DyingTimer -= dt
			if p.DyingTimer <= 0 {
				p.IsAlive = false
				p.Health = player.HealthDead
			}
		}

		// Action progress updates
		switch p.ActionState {
		case player.ActionRepairing:
			gen := g.findGenerator(p.ActionTarget)
			if gen != nil && !gen.Done {
				// Multiple repairers speed up (but with penalty)
				repairerCount := len(gen.RepairerIDs)
				speedMult := 1.0
				if repairerCount == 2 {
					speedMult = 1.7
				} else if repairerCount >= 3 {
					speedMult = 2.2
				}
				gen.Progress += (dt / player.RepairTime) * speedMult / float64(repairerCount)
				if gen.Progress >= 1.0 {
					gen.Progress = 1.0
					gen.Done = true
					gen.BeingRepaired = false
					gen.RepairerIDs = make(map[int]bool)
					g.GensCompleted++

					// Credit gen completion to repairers
					for uid := range gen.RepairerIDs {
						if pp, ok := g.Players[uid]; ok {
							pp.GensDone++
						}
					}
					p.GensDone++

					// Check if gates should be powered
					if g.GensCompleted >= GensRequired {
						g.GatesPowered = true
						for _, gate := range g.Gates {
							gate.Powered = true
						}
					}

					p.ActionState = player.ActionNone
					p.ActionTarget = ""
				}
				p.ActionProgress = gen.Progress
			}

		case player.ActionHealing:
			p.ActionProgress += dt / player.HealTime
			if p.ActionProgress >= 1.0 {
				// Find target and heal
				targetID := 0
				fmt.Sscanf(p.ActionTarget, "%d", &targetID)
				if target, ok := g.Players[targetID]; ok && target.Health == player.HealthInjured {
					target.Health = player.HealthHealthy
				}
				p.ActionState = player.ActionNone
				p.ActionProgress = 0
				p.ActionTarget = ""
			}

		case player.ActionUnhooking:
			p.ActionProgress += dt / player.UnhookTime
			if p.ActionProgress >= 1.0 {
				hook := g.findHook(p.ActionTarget)
				if hook != nil && hook.Occupied {
					if hooked, ok := g.Players[hook.PlayerID]; ok {
						g.unhookPlayer(hooked)
					}
				}
				p.ActionState = player.ActionNone
				p.ActionProgress = 0
				p.ActionTarget = ""
			}

		case player.ActionOpeningGate:
			gate := g.findGate(p.ActionTarget)
			if gate != nil && !gate.Open {
				gate.Progress += dt / player.GateOpenTime
				p.ActionProgress = gate.Progress
				if gate.Progress >= 1.0 {
					gate.Progress = 1.0
					gate.Open = true
					p.ActionState = player.ActionNone
					p.ActionProgress = 0
					p.ActionTarget = ""
				}
			}

		case player.ActionPlacingTrap:
			p.ActionProgress += dt / player.PlaceTrapTime
			if p.ActionProgress >= 1.0 {
				// Place trap at current position
				trap := &objects.Trap{
					ID:     fmt.Sprintf("trap_placed_%d_%d", p.UserID, g.Tick),
					Pos:    gamemap.Vec3{X: p.PosX + math.Sin(p.RotY)*1.5, Y: p.PosY, Z: p.PosZ + math.Cos(p.RotY)*1.5},
					Placed: true,
				}
				g.Traps = append(g.Traps, trap)
				p.TrapCount--
				p.ActionState = player.ActionNone
				p.ActionProgress = 0
			}

		case player.ActionBreakingPallet:
			p.ActionProgress += dt / player.BreakPalletTime
			if p.ActionProgress >= 1.0 {
				pallet := g.findPallet(p.ActionTarget)
				if pallet != nil {
					pallet.Broken = true
				}
				p.ActionState = player.ActionNone
				p.ActionProgress = 0
				p.ActionTarget = ""
			}
		}

		// Update carried survivor position
		if p.Role == player.RoleKiller && p.CarryingPlayerID > 0 {
			if carried, ok := g.Players[p.CarryingPlayerID]; ok {
				carried.PosX = p.PosX
				carried.PosY = p.PosY + 1.0
				carried.PosZ = p.PosZ
			}
		}

		// Scratch marks (running survivors)
		if p.Role == player.RoleSurvivor && p.MoveState == player.MoveRunning && p.IsAlive {
			if g.ElapsedTime-p.LastScratchTime > player.ScratchMarkInterval {
				g.ScratchMarks = append(g.ScratchMarks, protocol.ScratchMark{
					PosX: p.PosX + (rand.Float64()-0.5)*2,
					PosZ: p.PosZ + (rand.Float64()-0.5)*2,
					Age:  0,
				})
				p.LastScratchTime = g.ElapsedTime
			}
		}

		// Blood trails (injured survivors)
		if p.Role == player.RoleSurvivor && p.Health == player.HealthInjured && p.IsAlive {
			if g.ElapsedTime-p.LastBloodTime > player.BloodTrailInterval {
				g.BloodTrails = append(g.BloodTrails, protocol.BloodTrail{
					PosX: p.PosX,
					PosZ: p.PosZ,
					Age:  0,
				})
				p.LastBloodTime = g.ElapsedTime
			}
		}
	}

	// Generator regression
	for _, gen := range g.Generators {
		if gen.Regressing && !gen.Done && gen.Progress > 0 {
			gen.Progress -= player.GenRegressionRate * dt
			if gen.Progress < 0 {
				gen.Progress = 0
				gen.Regressing = false
			}
		}
	}

	// Age and clean scratch marks
	var activeScratch []protocol.ScratchMark
	for _, sm := range g.ScratchMarks {
		sm.Age += dt
		if sm.Age < player.ScratchMarkDuration {
			activeScratch = append(activeScratch, sm)
		}
	}
	g.ScratchMarks = activeScratch

	// Age and clean blood trails
	var activeBlood []protocol.BloodTrail
	for _, bt := range g.BloodTrails {
		bt.Age += dt
		if bt.Age < player.BloodTrailDuration {
			activeBlood = append(activeBlood, bt)
		}
	}
	g.BloodTrails = activeBlood

	// Check exit gate escapes
	for _, gate := range g.Gates {
		if gate.Open {
			for _, p := range g.Players {
				if p.Role == player.RoleSurvivor && p.IsAlive && !p.HasEscaped &&
					p.ActionState != player.ActionBeingCarried {
					// Check if survivor is past the gate
					if gamemap.IsNear(p.PosX, p.PosZ, gate.Pos.X, gate.Pos.Z, 3.0) {
						// Check if they're on the outside edge
						if (gate.ID == "gate_0" && p.PosX <= 1.0) ||
							(gate.ID == "gate_1" && p.PosX >= gamemap.MapWidth-1.0) {
							p.HasEscaped = true
							p.ActionState = player.ActionNone
						}
					}
				}
			}
		}
	}

	// Check win conditions
	g.checkWinConditions()
}

func (g *Game) checkWinConditions() {
	aliveSurvivors := 0
	escapedSurvivors := 0
	totalSurvivors := 0

	for _, p := range g.Players {
		if p.Role == player.RoleSurvivor {
			totalSurvivors++
			if p.IsAlive && !p.HasEscaped {
				aliveSurvivors++
			}
			if p.HasEscaped {
				escapedSurvivors++
			}
		}
	}

	// All survivors dead or escaped
	if aliveSurvivors == 0 {
		g.GameOver = true
		if escapedSurvivors > totalSurvivors/2 {
			g.Result = "survivor_win"
		} else {
			g.Result = "killer_win"
		}
	}
}

func (g *Game) buildState() protocol.GameState {
	state := protocol.GameState{
		Tick:          g.Tick,
		TimeRemaining: GameTimeout - g.ElapsedTime,
		GensCompleted: g.GensCompleted,
		GensRequired:  GensRequired,
		GatesPowered:  g.GatesPowered,
	}

	// Players
	for _, p := range g.Players {
		ps := protocol.PlayerState{
			UserID:         p.UserID,
			Username:       p.Username,
			Role:           p.Role,
			PosX:           p.PosX,
			PosY:           p.PosY,
			PosZ:           p.PosZ,
			RotY:           p.RotY,
			Health:         p.Health,
			MoveState:      p.MoveState,
			ActionState:    p.ActionState,
			ActionTarget:   p.ActionTarget,
			ActionProgress: p.ActionProgress,
			CarryingID:     p.CarryingPlayerID,
			HookedOn:       p.HookedOnID,
			HookStage:      p.HookStage,
			TrappedIn:      p.TrappedInID,
			Ping:           p.Ping,
			IsAlive:        p.IsAlive,
			HasEscaped:     p.HasEscaped,
			TrapCount:      p.TrapCount,
		}
		state.Players = append(state.Players, ps)
	}

	// Generators
	for _, gen := range g.Generators {
		state.Generators = append(state.Generators, protocol.GenState{
			ID:            gen.ID,
			PosX:          gen.Pos.X,
			PosY:          gen.Pos.Y,
			PosZ:          gen.Pos.Z,
			Progress:      gen.Progress,
			Done:          gen.Done,
			Regressing:    gen.Regressing,
			BeingRepaired: gen.BeingRepaired,
		})
	}

	// Pallets
	for _, pal := range g.Pallets {
		state.Pallets = append(state.Pallets, protocol.PalletState{
			ID:      pal.ID,
			PosX:    pal.Pos.X,
			PosY:    pal.Pos.Y,
			PosZ:    pal.Pos.Z,
			RotY:    pal.RotY,
			Dropped: pal.Dropped,
			Broken:  pal.Broken,
		})
	}

	// Traps (visibility based on role will be handled client-side)
	for _, trap := range g.Traps {
		state.Traps = append(state.Traps, protocol.TrapState{
			ID:        trap.ID,
			PosX:      trap.Pos.X,
			PosY:      trap.Pos.Y,
			PosZ:      trap.Pos.Z,
			Placed:    trap.Placed,
			Triggered: trap.Triggered,
			Visible:   true, // Simplified: always visible
		})
	}

	// Gates
	for _, gate := range g.Gates {
		state.Gates = append(state.Gates, protocol.GateState{
			ID:       gate.ID,
			PosX:     gate.Pos.X,
			PosY:     gate.Pos.Y,
			PosZ:     gate.Pos.Z,
			RotY:     gate.RotY,
			Progress: gate.Progress,
			Open:     gate.Open,
			Powered:  gate.Powered,
		})
	}

	// Hooks
	for _, hook := range g.Hooks {
		state.Hooks = append(state.Hooks, protocol.HookState{
			ID:       hook.ID,
			PosX:     hook.Pos.X,
			PosY:     hook.Pos.Y,
			PosZ:     hook.Pos.Z,
			Occupied: hook.Occupied,
			PlayerID: hook.PlayerID,
		})
	}

	// Windows
	for _, w := range g.Windows {
		state.Windows = append(state.Windows, protocol.WindowState{
			ID:   w.ID,
			PosX: w.Pos.X,
			PosY: w.Pos.Y,
			PosZ: w.Pos.Z,
			RotY: w.RotY,
		})
	}

	// Scratch marks and blood trails
	state.ScratchMarks = g.ScratchMarks
	state.BloodTrails = g.BloodTrails

	return state
}

func (g *Game) endGame() {
	log.Printf("Game %d ending with result: %s", g.ID, g.Result)

	// Build report
	report := protocol.GameReport{
		GameID: g.ID,
		Result: g.Result,
	}

	g.mu.RLock()
	for _, p := range g.Players {
		report.Players = append(report.Players, protocol.PlayerReport{
			UserID:   p.UserID,
			Role:     p.Role,
			Survived: p.HasEscaped || (p.IsAlive && p.Role == player.RoleSurvivor),
			Kills:    p.Kills,
			GensDone: p.GensDone,
		})
	}

	// Notify all players
	endMsg, _ := json.Marshal(protocol.ServerMessage{
		Type: "game_end",
		Payload: map[string]interface{}{
			"result":  g.Result,
			"players": report.Players,
		},
	})

	for _, ch := range g.connections {
		select {
		case ch <- endMsg:
		default:
		}
	}
	g.mu.RUnlock()

	// Report to master server
	g.sendReport(report)
}

func (g *Game) sendReport(report protocol.GameReport) {
	data, _ := json.Marshal(report)
	resp, err := http.Post(g.MasterURL+"/api/internal/game-report", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to send game report: %v", err)
		return
	}
	resp.Body.Close()
}

func (g *Game) reportProgress() {
	g.mu.RLock()
	defer g.mu.RUnlock()

	survivorsAlive := 0
	var playerNames []string
	killerName := ""

	for _, p := range g.Players {
		playerNames = append(playerNames, p.Username)
		if p.Role == player.RoleSurvivor && p.IsAlive {
			survivorsAlive++
		}
		if p.Role == player.RoleKiller {
			killerName = p.Username
		}
	}

	progress := protocol.GameProgress{
		GameID:         g.ID,
		GensCompleted:  g.GensCompleted,
		SurvivorsAlive: survivorsAlive,
		ElapsedSeconds: g.ElapsedTime,
		PlayerNames:    playerNames,
		KillerName:     killerName,
	}

	data, _ := json.Marshal(progress)
	resp, err := http.Post(g.MasterURL+"/api/internal/game-progress", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to send progress: %v", err)
		return
	}
	resp.Body.Close()
}

func (g *Game) IsGameOver() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.GameOver
}

func (g *Game) PlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Players)
}
