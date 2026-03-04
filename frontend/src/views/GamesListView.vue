<template>
  <div class="page">
    <header class="top-bar">
      <h1>DEAD BY DAYLIGHT <span class="dim">CLONE</span></h1>
      <div class="nav-links">
        <router-link to="/lobbies" v-if="auth.isLoggedIn">Lobbies</router-link>
        <router-link to="/leaderboard">Leaderboard</router-link>
        <router-link to="/games">Live Games</router-link>
        <router-link to="/login" v-if="!auth.isLoggedIn">Login</router-link>
      </div>
    </header>

    <div class="content">
      <div class="section-header">
        <h2>Live Games</h2>
        <button class="btn-secondary" @click="refresh">Refresh</button>
      </div>

      <div v-if="games.length === 0" class="empty">
        <p>No active games right now.</p>
      </div>

      <div v-else class="games-grid">
        <div v-for="g in games" :key="g.id" class="game-card">
          <div class="game-header">
            <span class="game-id">Game #{{ g.id }}</span>
            <span class="game-status" :class="g.status">{{ g.status }}</span>
          </div>
          <div class="game-info">
            <div class="info-row">
              <span class="label">Killer:</span>
              <span>{{ g.killer_name }}</span>
            </div>
            <div class="info-row" v-if="g.progress">
              <span class="label">Gens:</span>
              <span>{{ g.progress.gens_completed }} / 5</span>
            </div>
            <div class="info-row" v-if="g.progress">
              <span class="label">Survivors:</span>
              <span>{{ g.progress.survivors_alive }} alive</span>
            </div>
            <div class="info-row" v-if="g.progress">
              <span class="label">Time:</span>
              <span>{{ formatTime(g.progress.elapsed_seconds) }}</span>
            </div>
          </div>
          <div class="game-players">
            <span v-for="p in g.players" :key="p.user_id" class="player-tag" :class="p.role">
              {{ p.username }} ({{ p.role }})
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import axios from 'axios'

const auth = useAuthStore()
const games = ref<any[]>([])
let timer: any

onMounted(() => {
  refresh()
  timer = setInterval(refresh, 5000)
})

onUnmounted(() => {
  clearInterval(timer)
})

async function refresh() {
  try {
    const res = await axios.get('/api/games')
    games.value = res.data
  } catch (e) {
    console.error('Failed to fetch games', e)
  }
}

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}
</script>

<style scoped>
.page { height: 100vh; display: flex; flex-direction: column; }
.top-bar { display: flex; justify-content: space-between; align-items: center; padding: 16px 24px; border-bottom: 1px solid var(--border); }
.top-bar h1 { font-size: 18px; color: var(--primary); letter-spacing: 2px; }
.dim { color: var(--text-dim); font-size: 12px; }
.nav-links { display: flex; gap: 16px; font-size: 13px; }

.content { flex: 1; padding: 24px; overflow-y: auto; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }

.empty { text-align: center; padding: 60px; color: var(--text-dim); }

.games-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); gap: 12px; }

.game-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
}
.game-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.game-id { font-weight: 700; }
.game-status { font-size: 11px; text-transform: uppercase; padding: 3px 8px; border-radius: 3px; }
.game-status.in_progress { background: var(--success); color: white; }

.game-info { margin-bottom: 12px; }
.info-row { display: flex; gap: 8px; font-size: 13px; padding: 3px 0; }
.info-row .label { color: var(--text-dim); min-width: 70px; }

.game-players { display: flex; flex-wrap: wrap; gap: 6px; }
.player-tag {
  padding: 3px 8px;
  border-radius: 3px;
  font-size: 11px;
  background: var(--bg-hover);
}
.player-tag.killer { background: var(--primary); color: white; }
.player-tag.survivor { background: #1a3a1a; color: #4caf50; }
</style>
