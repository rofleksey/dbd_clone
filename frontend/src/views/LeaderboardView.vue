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
      <h2>Leaderboard</h2>
      <table class="leaderboard-table" v-if="stats.length > 0">
        <thead>
          <tr>
            <th>#</th>
            <th>Player</th>
            <th>Wins</th>
            <th>Games</th>
            <th>Kills</th>
            <th>Escapes</th>
            <th>Gens</th>
            <th>Win Rate</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(s, i) in stats" :key="s.user_id" :class="{ highlight: s.username === auth.username }">
            <td class="rank">{{ i + 1 }}</td>
            <td class="name">{{ s.username }}</td>
            <td>{{ s.games_won }}</td>
            <td>{{ s.games_played }}</td>
            <td>{{ s.kills }}</td>
            <td>{{ s.escapes }}</td>
            <td>{{ s.generators_done }}</td>
            <td>{{ s.games_played > 0 ? Math.round(s.games_won / s.games_played * 100) : 0 }}%</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">No players yet.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import axios from 'axios'

const auth = useAuthStore()
const stats = ref<any[]>([])

onMounted(async () => {
  try {
    const res = await axios.get('/api/leaderboard?limit=50')
    stats.value = res.data
  } catch (e) {
    console.error('Failed to load leaderboard', e)
  }
})
</script>

<style scoped>
.page { height: 100vh; display: flex; flex-direction: column; }
.top-bar { display: flex; justify-content: space-between; align-items: center; padding: 16px 24px; border-bottom: 1px solid var(--border); }
.top-bar h1 { font-size: 18px; color: var(--primary); letter-spacing: 2px; }
.dim { color: var(--text-dim); font-size: 12px; }
.nav-links { display: flex; gap: 16px; font-size: 13px; }

.content { flex: 1; padding: 24px; overflow-y: auto; }
.content h2 { margin-bottom: 20px; }

.leaderboard-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
.leaderboard-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid var(--border);
  color: var(--text-dim);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 1px;
}
.leaderboard-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}
.leaderboard-table tr:hover { background: var(--bg-hover); }
.leaderboard-table tr.highlight { background: #1a1a2a; }
.rank { color: var(--primary); font-weight: 700; }
.name { font-weight: 600; }

.empty { text-align: center; padding: 60px; color: var(--text-dim); }
</style>
