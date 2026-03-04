import { useGameStore, type GameStateData } from '../stores/game'

export class GameNetwork {
  private ws: WebSocket | null = null
  private store = useGameStore()
  private onDisconnect: (() => void) | null = null

  connect(gameId: number, gamePort: number, token: string, userId: number): Promise<void> {
    return new Promise((resolve, reject) => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws/game/${gameId}?token=${token}`

      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        // Send auth message
        this.send({ type: 'auth', user_id: userId, token: token })
      }

      this.ws.onmessage = (event) => {
        const msg = JSON.parse(event.data)
        switch (msg.type) {
          case 'connected':
            this.store.isConnected = true
            resolve()
            break
          case 'state':
            this.store.updateState(msg.payload as GameStateData)
            break
          case 'game_end':
            this.store.endGame(msg.payload.result)
            break
          case 'error':
            console.error('Game error:', msg.payload)
            reject(new Error(msg.payload))
            break
        }
      }

      this.ws.onclose = () => {
        this.store.isConnected = false
        if (this.onDisconnect) this.onDisconnect()
      }

      this.ws.onerror = () => {
        reject(new Error('WebSocket connection failed'))
      }

      // Timeout
      setTimeout(() => {
        if (!this.store.isConnected) {
          reject(new Error('Connection timeout'))
        }
      }, 10000)
    })
  }

  send(msg: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  sendMove(posX: number, posY: number, posZ: number, rotY: number, state: string) {
    this.send({
      type: 'move',
      pos_x: posX,
      pos_y: posY,
      pos_z: posZ,
      rot_y: rotY,
      state: state,
    })
  }

  sendInteract(action: string, target?: string) {
    this.send({ type: 'interact', action, target })
  }

  sendStopInteract() {
    this.send({ type: 'stop_interact' })
  }

  sendPing() {
    this.send({ type: 'ping' })
  }

  setOnDisconnect(cb: () => void) {
    this.onDisconnect = cb
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}
