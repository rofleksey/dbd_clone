<template>
  <div class="game-view">
    <div ref="gameContainer" class="game-canvas"></div>

    <!-- HUD Overlay -->
    <div class="hud" v-if="gameStore.gameState && !gameStore.gameOver">
      <!-- Top: Generator progress -->
      <div class="hud-top">
        <div class="gen-counter">
          <span class="gen-icon">&#x2699;</span>
          <span class="gen-text">{{ gameStore.gameState.gens_completed }} / {{ gameStore.gameState.gens_required }}</span>
          <span v-if="gameStore.gameState.gates_powered" class="gates-powered">GATES POWERED</span>
        </div>
        <div class="timer">{{ formatTime(gameStore.gameState.time_remaining) }}</div>
      </div>

      <!-- Interaction prompt -->
      <div class="interact-prompt" v-if="interactText">
        {{ interactText }}
      </div>

      <!-- Action progress bar -->
      <div class="action-bar" v-if="myPlayer && myPlayer.action_progress > 0 && myPlayer.action_progress < 1">
        <div class="action-bar-label">{{ getActionLabel(myPlayer.action_state) }}</div>
        <div class="action-bar-track">
          <div class="action-bar-fill" :style="{ width: (myPlayer.action_progress * 100) + '%' }"></div>
        </div>
      </div>

      <!-- Self-unhook / escape trap prompt -->
      <div class="special-prompt" v-if="myPlayer && myPlayer.action_state === 'hooked'">
        <p>HOOKED - Stage {{ myPlayer.hook_stage }}</p>
        <p v-if="myPlayer.hook_stage < 2">Press SPACE to attempt self-unhook (10% chance)</p>
        <p v-else>Struggle! Press SPACE rapidly!</p>
      </div>
      <div class="special-prompt" v-if="myPlayer && myPlayer.action_state === 'trapped'">
        <p>TRAPPED!</p>
        <p>Press SPACE to attempt escape (25% chance)</p>
      </div>

      <!-- Killer HUD extras -->
      <div class="killer-hud" v-if="myPlayer && myPlayer.role === 'killer'">
        <div class="trap-counter">Traps: {{ myPlayer.trap_count }}</div>
        <div class="killer-controls">
          <span>LMB: Attack</span>
          <span>R: Place Trap</span>
          <span>Q: Drop Survivor</span>
        </div>
      </div>

      <!-- Survivor controls hint -->
      <div class="survivor-hud" v-if="myPlayer && myPlayer.role === 'survivor'">
        <div class="health-bar">
          <span v-for="i in 3" :key="i" class="health-pip" :class="{ active: i <= (myPlayer.health || 0) }"></span>
        </div>
        <div class="survivor-controls">
          <span>WASD: Move</span>
          <span>Shift: Walk</span>
          <span>Ctrl: Crouch</span>
          <span>Space: Vault/Pallet</span>
          <span>E: Interact</span>
        </div>
      </div>

      <!-- Bottom: Scoreboard toggle -->
      <div class="hud-bottom">
        <div class="scoreboard-hint">Hold TAB for scoreboard</div>
      </div>

      <!-- Scoreboard overlay (TAB held) -->
      <div class="scoreboard" v-if="showScoreboard">
        <h3>Players</h3>
        <table>
          <thead>
            <tr>
              <th>Player</th>
              <th>Role</th>
              <th>Status</th>
              <th>Ping</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in gameStore.gameState?.players || []" :key="p.user_id"
                :class="{ me: p.user_id === authStore.userId, dead: !p.is_alive, escaped: p.has_escaped }">
              <td>{{ p.username }}</td>
              <td :class="p.role">{{ p.role === 'killer' ? 'The Trapper' : 'Survivor' }}</td>
              <td>
                <span v-if="!p.is_alive" class="dead-text">DEAD</span>
                <span v-else-if="p.has_escaped" class="escaped-text">ESCAPED</span>
                <span v-else-if="p.action_state === 'hooked'" class="hooked-text">HOOKED ({{ p.hook_stage }})</span>
                <span v-else-if="p.health === 1" class="dying-text">DYING</span>
                <span v-else-if="p.health === 2" class="injured-text">INJURED</span>
                <span v-else class="healthy-text">HEALTHY</span>
              </td>
              <td>{{ p.ping }}ms</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Game Over screen -->
    <div class="game-over" v-if="gameStore.gameOver">
      <div class="game-over-card">
        <h1>{{ getResultText() }}</h1>
        <p class="result-subtitle">{{ getResultSubtitle() }}</p>
        <div class="result-players" v-if="gameStore.gameState">
          <div v-for="p in gameStore.gameState.players" :key="p.user_id" class="result-player" :class="p.role">
            <span class="rp-name">{{ p.username }}</span>
            <span class="rp-role">{{ p.role === 'killer' ? 'The Trapper' : 'Survivor' }}</span>
            <span v-if="p.role === 'survivor'">
              {{ p.has_escaped ? 'ESCAPED' : p.is_alive ? 'ALIVE' : 'DEAD' }}
            </span>
          </div>
        </div>
        <button class="btn-primary" @click="returnToLobby">Return to Lobbies</button>
      </div>
    </div>

    <!-- Loading screen -->
    <div class="loading" v-if="loading">
      <div class="loading-content">
        <h2>Loading...</h2>
        <p>{{ loadingText }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useGameStore } from '../stores/game'
import { useAuthStore } from '../stores/auth'
import { GameEngine } from '../game/Engine'

const router = useRouter()
const gameStore = useGameStore()
const authStore = useAuthStore()

const gameContainer = ref<HTMLElement | null>(null)
const loading = ref(true)
const loadingText = ref('Connecting to game server...')
const showScoreboard = ref(false)
const interactText = ref('')
let engine: GameEngine | null = null
let interactCheckInterval: any

const myPlayer = computed(() => gameStore.getMyPlayer())

onMounted(async () => {
  if (!gameStore.gameId) {
    router.push('/lobbies')
    return
  }

  // TAB for scoreboard
  document.addEventListener('keydown', onKeyDown)
  document.addEventListener('keyup', onKeyUp)

  try {
    loadingText.value = 'Initializing 3D engine...'

    engine = new GameEngine(gameContainer.value!)
    await engine.init()

    loading.value = false

    // Check for nearby interactables periodically
    interactCheckInterval = setInterval(() => {
      if (engine) {
        const nearby = engine.getNearbyInteractable()
        interactText.value = nearby ? nearby.name : ''
      }
    }, 100)
  } catch (e: any) {
    loadingText.value = `Error: ${e.message}`
    console.error('Game init error:', e)
  }
})

onUnmounted(() => {
  if (engine) {
    engine.destroy()
    engine = null
  }
  document.removeEventListener('keydown', onKeyDown)
  document.removeEventListener('keyup', onKeyUp)
  clearInterval(interactCheckInterval)
  gameStore.reset()
})

function onKeyDown(e: KeyboardEvent) {
  if (e.code === 'Tab') {
    e.preventDefault()
    showScoreboard.value = true
  }
}

function onKeyUp(e: KeyboardEvent) {
  if (e.code === 'Tab') {
    showScoreboard.value = false
  }
}

function formatTime(seconds: number): string {
  if (seconds < 0) seconds = 0
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

function getActionLabel(action: string): string {
  const labels: Record<string, string> = {
    repairing: 'Repairing...',
    healing: 'Healing...',
    unhooking: 'Unhooking...',
    opening_gate: 'Opening Gate...',
    placing_trap: 'Setting Trap...',
    picking_up_trap: 'Picking Up...',
    breaking_pallet: 'Breaking Pallet...',
  }
  return labels[action] || action
}

function getResultText(): string {
  const result = gameStore.gameResult
  const me = myPlayer.value

  if (result === 'disconnected') return 'DISCONNECTED'
  if (result === 'timeout') return 'TIME UP'

  if (!me) return result.toUpperCase()

  if (me.role === 'survivor') {
    return me.has_escaped ? 'ESCAPED!' : 'SACRIFICED'
  } else {
    return result === 'killer_win' ? 'ENTITY PLEASED' : 'ENTITY DISPLEASED'
  }
}

function getResultSubtitle(): string {
  const result = gameStore.gameResult
  if (result === 'disconnected') return 'A player disconnected. The match has been cancelled.'
  if (result === 'timeout') return 'Time ran out.'
  if (result === 'killer_win') return 'The killer has dominated this trial.'
  if (result === 'survivor_win') return 'The survivors escaped the trial.'
  return ''
}

function returnToLobby() {
  gameStore.reset()
  router.push('/lobbies')
}
</script>

<style scoped>
.game-view {
  width: 100vw;
  height: 100vh;
  position: relative;
  overflow: hidden;
}

.game-canvas {
  width: 100%;
  height: 100%;
}

.hud {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 10;
}

.hud-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
}

.gen-counter {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(0,0,0,0.7);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 18px;
  font-weight: 700;
}

.gen-icon { font-size: 24px; }

.gates-powered {
  color: #22aa22;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 2px;
  animation: pulse 1s ease-in-out infinite alternate;
}

@keyframes pulse {
  from { opacity: 0.7; }
  to { opacity: 1; }
}

.timer {
  background: rgba(0,0,0,0.7);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 18px;
  font-family: monospace;
}

.interact-prompt {
  position: absolute;
  bottom: 200px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0,0,0,0.8);
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  border: 1px solid rgba(255,255,255,0.2);
}

.action-bar {
  position: absolute;
  bottom: 150px;
  left: 50%;
  transform: translateX(-50%);
  width: 300px;
}

.action-bar-label {
  text-align: center;
  font-size: 13px;
  margin-bottom: 6px;
  color: #ccc;
}

.action-bar-track {
  height: 8px;
  background: rgba(255,255,255,0.1);
  border-radius: 4px;
  overflow: hidden;
}

.action-bar-fill {
  height: 100%;
  background: #f0c040;
  border-radius: 4px;
  transition: width 0.05s linear;
}

.special-prompt {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  background: rgba(0,0,0,0.8);
  padding: 20px 30px;
  border-radius: 8px;
  border: 1px solid var(--primary);
}

.special-prompt p:first-child {
  font-size: 24px;
  font-weight: 700;
  color: var(--primary);
  margin-bottom: 8px;
}

.killer-hud {
  position: absolute;
  bottom: 24px;
  left: 24px;
}

.trap-counter {
  background: rgba(0,0,0,0.7);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}

.killer-controls, .survivor-controls {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: var(--text-dim);
}

.survivor-hud {
  position: absolute;
  bottom: 24px;
  left: 24px;
}

.health-bar {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
}

.health-pip {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255,255,255,0.3);
  border-radius: 3px;
}

.health-pip.active {
  background: #cc2222;
  border-color: #ff4444;
}

.hud-bottom {
  position: absolute;
  bottom: 24px;
  right: 24px;
}

.scoreboard-hint {
  font-size: 11px;
  color: var(--text-dim);
}

/* Scoreboard */
.scoreboard {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(0,0,0,0.9);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  min-width: 500px;
  pointer-events: auto;
}

.scoreboard h3 {
  margin-bottom: 12px;
  font-size: 16px;
  text-align: center;
}

.scoreboard table {
  width: 100%;
  border-collapse: collapse;
}

.scoreboard th {
  text-align: left;
  padding: 8px;
  border-bottom: 1px solid var(--border);
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-dim);
}

.scoreboard td {
  padding: 8px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.scoreboard tr.me td { background: rgba(100,100,255,0.1); }
.scoreboard tr.dead td { opacity: 0.4; }
.scoreboard tr.escaped td { opacity: 0.7; }

.scoreboard .killer { color: var(--primary); font-weight: 600; }
.scoreboard .survivor { color: #4caf50; }

.dead-text { color: #888; }
.escaped-text { color: #4caf50; }
.hooked-text { color: var(--primary); }
.dying-text { color: #ff8800; }
.injured-text { color: #ffaa00; }
.healthy-text { color: #4caf50; }

/* Game Over */
.game-over {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 20;
}

.game-over-card {
  text-align: center;
  padding: 40px;
}

.game-over-card h1 {
  font-size: 48px;
  color: var(--primary);
  letter-spacing: 4px;
  margin-bottom: 8px;
}

.result-subtitle {
  color: var(--text-dim);
  margin-bottom: 32px;
  font-size: 16px;
}

.result-players {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
}

.result-player {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 10px 16px;
  background: rgba(255,255,255,0.05);
  border-radius: 4px;
  min-width: 400px;
}

.result-player.killer { border-left: 3px solid var(--primary); }
.result-player.survivor { border-left: 3px solid #4caf50; }

.rp-name { font-weight: 600; }
.rp-role { color: var(--text-dim); font-size: 13px; }

.game-over-card button {
  pointer-events: auto;
  font-size: 16px;
  padding: 12px 32px;
}

/* Loading */
.loading {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: #0a0a0a;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 30;
}

.loading-content {
  text-align: center;
}

.loading-content h2 {
  font-size: 24px;
  color: var(--primary);
  margin-bottom: 12px;
}

.loading-content p {
  color: var(--text-dim);
}
</style>
