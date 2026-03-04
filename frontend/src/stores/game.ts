import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface PlayerState {
  user_id: number
  username: string
  role: string
  pos_x: number
  pos_y: number
  pos_z: number
  rot_y: number
  health: number
  move_state: string
  action_state: string
  action_target: string
  action_progress: number
  carrying_id: number
  hooked_on: string
  hook_stage: number
  trapped_in: string
  ping: number
  is_alive: boolean
  has_escaped: boolean
  trap_count: number
}

export interface GenState {
  id: string
  pos_x: number
  pos_y: number
  pos_z: number
  progress: number
  done: boolean
  regressing: boolean
  being_repaired: boolean
}

export interface PalletState {
  id: string
  pos_x: number
  pos_y: number
  pos_z: number
  rot_y: number
  dropped: boolean
  broken: boolean
}

export interface TrapState {
  id: string
  pos_x: number
  pos_y: number
  pos_z: number
  placed: boolean
  triggered: boolean
  visible: boolean
}

export interface GateState {
  id: string
  pos_x: number
  pos_y: number
  pos_z: number
  rot_y: number
  progress: number
  open: boolean
  powered: boolean
}

export interface HookState {
  id: string
  pos_x: number
  pos_y: number
  pos_z: number
  occupied: boolean
  player_id: number
}

export interface WindowState {
  id: string
  pos_x: number
  pos_y: number
  pos_z: number
  rot_y: number
}

export interface ScratchMark {
  pos_x: number
  pos_z: number
  age: number
}

export interface BloodTrail {
  pos_x: number
  pos_z: number
  age: number
}

export interface GameStateData {
  tick: number
  players: PlayerState[]
  generators: GenState[]
  pallets: PalletState[]
  traps: TrapState[]
  gates: GateState[]
  hooks: HookState[]
  windows: WindowState[]
  scratch_marks: ScratchMark[]
  blood_trails: BloodTrail[]
  time_remaining: number
  gens_completed: number
  gens_required: number
  gates_powered: boolean
}

export const useGameStore = defineStore('game', () => {
  const gameState = ref<GameStateData | null>(null)
  const myRole = ref('')
  const myUserId = ref(0)
  const gameId = ref(0)
  const gamePort = ref(0)
  const isConnected = ref(false)
  const gameResult = ref('')
  const gameOver = ref(false)

  function updateState(state: GameStateData) {
    gameState.value = state
  }

  function setGameInfo(id: number, port: number, role: string, uid: number) {
    gameId.value = id
    gamePort.value = port
    myRole.value = role
    myUserId.value = uid
  }

  function endGame(result: string) {
    gameResult.value = result
    gameOver.value = true
  }

  function reset() {
    gameState.value = null
    myRole.value = ''
    gameId.value = 0
    gamePort.value = 0
    isConnected.value = false
    gameResult.value = ''
    gameOver.value = false
  }

  function getMyPlayer(): PlayerState | null {
    if (!gameState.value) return null
    return gameState.value.players.find(p => p.user_id === myUserId.value) || null
  }

  return {
    gameState, myRole, myUserId, gameId, gamePort, isConnected,
    gameResult, gameOver, updateState, setGameInfo, endGame, reset, getMyPlayer
  }
})
