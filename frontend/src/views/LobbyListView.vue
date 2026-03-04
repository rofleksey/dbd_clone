<template>
  <div class="lobby-list-page">
    <header class="top-bar">
      <h1>DEAD BY DAYLIGHT <span class="dim">CLONE</span></h1>
      <div class="user-info">
        <span>{{ auth.username }}</span>
        <button class="btn-secondary" @click="logout">Logout</button>
      </div>
    </header>

    <nav class="nav-bar">
      <router-link to="/lobbies" class="nav-active">Lobbies</router-link>
      <router-link to="/leaderboard">Leaderboard</router-link>
      <router-link to="/games">Live Games</router-link>
    </nav>

    <div class="content">
      <div class="section-header">
        <h2>Game Lobbies</h2>
        <div class="actions">
          <button class="btn-secondary" @click="refreshLobbies">Refresh</button>
          <button class="btn-primary" @click="showCreateModal = true">Create Lobby</button>
        </div>
      </div>

      <div v-if="lobbies.length === 0" class="empty-state">
        <p>No lobbies available. Create one to get started!</p>
      </div>

      <div v-else class="lobby-grid">
        <div v-for="lobby in lobbies" :key="lobby.id" class="lobby-card" @click="joinLobby(lobby.id)">
          <div class="lobby-name">{{ lobby.name }}</div>
          <div class="lobby-info">
            <span class="lobby-host">Host: {{ lobby.host_name }}</span>
            <span class="lobby-players">{{ lobby.players?.length || 0 }} / {{ lobby.max_players }} players</span>
          </div>
          <div class="lobby-player-list">
            <span v-for="p in lobby.players" :key="p.user_id" class="player-tag" :class="{ ready: p.ready }">
              {{ p.username }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Lobby Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal">
        <h3>Create Lobby</h3>
        <div class="form-group">
          <label>Lobby Name</label>
          <input v-model="newLobbyName" type="text" placeholder="My Lobby" maxlength="50" />
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showCreateModal = false">Cancel</button>
          <button class="btn-primary" @click="createLobby">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore, api } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const lobbies = ref<any[]>([])
const showCreateModal = ref(false)
const newLobbyName = ref('')
let refreshTimer: any

onMounted(() => {
  refreshLobbies()
  refreshTimer = setInterval(refreshLobbies, 3000)
})

onUnmounted(() => {
  clearInterval(refreshTimer)
})

async function refreshLobbies() {
  try {
    const res = await api.get('/lobbies')
    lobbies.value = res.data
  } catch (e) {
    console.error('Failed to fetch lobbies', e)
  }
}

async function createLobby() {
  try {
    const res = await api.post('/lobbies', { name: newLobbyName.value || `${auth.username}'s lobby` })
    showCreateModal.value = false
    newLobbyName.value = ''
    router.push(`/lobby/${res.data.id}`)
  } catch (e: any) {
    alert(e.response?.data?.error || 'Failed to create lobby')
  }
}

function joinLobby(id: string) {
  router.push(`/lobby/${id}`)
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.lobby-list-page { height: 100vh; display: flex; flex-direction: column; }

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
}
.top-bar h1 { font-size: 18px; color: var(--primary); letter-spacing: 2px; }
.top-bar .dim { color: var(--text-dim); font-size: 12px; }
.user-info { display: flex; align-items: center; gap: 12px; }

.nav-bar {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border);
}
.nav-bar a {
  padding: 12px 24px;
  color: var(--text-dim);
  text-decoration: none;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 1px;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}
.nav-bar a:hover, .nav-bar .nav-active { color: var(--text); border-bottom-color: var(--primary); }

.content { flex: 1; padding: 24px; overflow-y: auto; }

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.section-header h2 { font-size: 20px; }
.actions { display: flex; gap: 8px; }

.empty-state {
  text-align: center;
  padding: 60px;
  color: var(--text-dim);
}

.lobby-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
}

.lobby-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
}
.lobby-card:hover { border-color: var(--primary); background: var(--bg-hover); }

.lobby-name { font-size: 16px; font-weight: 600; margin-bottom: 8px; }
.lobby-info { display: flex; justify-content: space-between; font-size: 12px; color: var(--text-dim); margin-bottom: 10px; }

.lobby-player-list { display: flex; flex-wrap: wrap; gap: 6px; }
.player-tag {
  background: var(--bg-hover);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
}
.player-tag.ready { background: var(--success); color: white; }

.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.modal {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
  width: 400px;
}
.modal h3 { margin-bottom: 16px; }
.modal .form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.modal .form-group label { font-size: 12px; color: var(--text-dim); text-transform: uppercase; }
.modal .form-group input { width: 100%; }
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; }
</style>
