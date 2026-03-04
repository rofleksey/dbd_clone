<template>
  <div class="lobby-page">
    <header class="top-bar">
      <div class="back-section">
        <button class="btn-secondary" @click="leaveLobby">Back</button>
        <h1>{{ lobbyData?.name || 'Lobby' }}</h1>
      </div>
      <div class="lobby-id">ID: {{ $route.params.id }}</div>
    </header>

    <div class="lobby-content">
      <div class="players-section">
        <h2>Players ({{ lobbyData?.players?.length || 0 }} / {{ lobbyData?.max_players || 5 }})</h2>
        <div class="player-list">
          <div v-for="p in lobbyData?.players || []" :key="p.user_id" class="player-row" :class="{ ready: p.ready, me: p.user_id === auth.userId }">
            <div class="player-name">
              {{ p.username }}
              <span v-if="p.user_id === lobbyData?.host_id" class="host-badge">HOST</span>
              <span v-if="p.user_id === auth.userId" class="me-badge">YOU</span>
            </div>
            <div class="player-status">
              <span v-if="p.ready" class="status-ready">READY</span>
              <span v-else class="status-waiting">WAITING</span>
            </div>
          </div>
        </div>

        <div v-for="i in (5 - (lobbyData?.players?.length || 0))" :key="'empty-' + i" class="player-row empty">
          <div class="player-name dim">Empty Slot</div>
        </div>
      </div>

      <div class="actions-section">
        <p class="info-text">
          Minimum 2 players required. All players must be ready to start.
          One random player will be chosen as The Trapper.
        </p>
        <button
          class="btn-ready"
          :class="{ ready: isReady }"
          @click="toggleReady"
        >
          {{ isReady ? 'CANCEL READY' : 'READY UP' }}
        </button>
      </div>
    </div>

    <div v-if="statusMessage" class="status-banner" :class="statusType">
      {{ statusMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useGameStore } from '../stores/game'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const gameStore = useGameStore()

const lobbyData = ref<any>(null)
const isReady = ref(false)
const statusMessage = ref('')
const statusType = ref('info')
let ws: WebSocket | null = null

onMounted(() => {
  connectWS()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

function connectWS() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/lobby/${route.params.id}?token=${auth.token}`

  ws = new WebSocket(wsUrl)

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data)
    switch (msg.type) {
      case 'lobby_state':
        lobbyData.value = msg.payload
        // Sync our ready state
        const me = msg.payload.players?.find((p: any) => p.user_id === auth.userId)
        if (me) isReady.value = me.ready
        break

      case 'game_start':
        handleGameStart(msg.payload)
        break

      case 'game_error':
        statusMessage.value = msg.payload
        statusType.value = 'error'
        break

      case 'error':
        statusMessage.value = msg.payload
        statusType.value = 'error'
        break
    }
  }

  ws.onclose = () => {
    statusMessage.value = 'Disconnected from lobby'
    statusType.value = 'error'
  }

  ws.onerror = () => {
    statusMessage.value = 'Connection error'
    statusType.value = 'error'
  }
}

function toggleReady() {
  isReady.value = !isReady.value
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'ready', payload: isReady.value }))
  }
}

function handleGameStart(payload: any) {
  statusMessage.value = 'Game starting...'
  statusType.value = 'success'

  const myPlayer = payload.players.find((p: any) => p.user_id === auth.userId)
  if (myPlayer) {
    gameStore.setGameInfo(payload.game_id, payload.port, myPlayer.role, auth.userId)

    setTimeout(() => {
      router.push(`/game/${payload.game_id}`)
    }, 1000)
  }
}

function leaveLobby() {
  if (ws) {
    ws.close()
    ws = null
  }
  router.push('/lobbies')
}
</script>

<style scoped>
.lobby-page { height: 100vh; display: flex; flex-direction: column; }

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
}
.back-section { display: flex; align-items: center; gap: 16px; }
.back-section h1 { font-size: 18px; }
.lobby-id { font-size: 12px; color: var(--text-dim); font-family: monospace; }

.lobby-content { flex: 1; padding: 24px; display: flex; flex-direction: column; align-items: center; }

.players-section { width: 100%; max-width: 500px; }
.players-section h2 { margin-bottom: 16px; font-size: 16px; }

.player-list { display: flex; flex-direction: column; gap: 4px; }

.player-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  margin-bottom: 4px;
}
.player-row.ready { border-color: var(--success); }
.player-row.me { background: #1a1a2a; }
.player-row.empty { opacity: 0.3; }

.player-name { display: flex; align-items: center; gap: 8px; font-weight: 500; }
.host-badge { background: var(--warning); color: #000; padding: 2px 6px; border-radius: 3px; font-size: 10px; font-weight: 700; }
.me-badge { background: var(--primary); color: white; padding: 2px 6px; border-radius: 3px; font-size: 10px; font-weight: 700; }

.status-ready { color: var(--success); font-weight: 600; font-size: 13px; }
.status-waiting { color: var(--text-dim); font-size: 13px; }

.dim { color: var(--text-dim); }

.actions-section { margin-top: 32px; text-align: center; }
.info-text { color: var(--text-dim); font-size: 13px; margin-bottom: 20px; max-width: 400px; line-height: 1.5; }

.btn-ready {
  padding: 16px 48px;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 2px;
  background: var(--success);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-ready:hover { transform: scale(1.05); }
.btn-ready.ready { background: var(--warning); }

.status-banner {
  padding: 12px;
  text-align: center;
  font-weight: 600;
  font-size: 14px;
}
.status-banner.info { background: #1a3a5c; color: #6db3f8; }
.status-banner.success { background: #1a3a1a; color: #4caf50; }
.status-banner.error { background: #3a1a1a; color: #ef5350; }
</style>
