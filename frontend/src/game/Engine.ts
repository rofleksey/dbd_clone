import * as THREE from 'three'
import { GameNetwork } from './Network'
import { useGameStore, type PlayerState } from '../stores/game'
import { useAuthStore } from '../stores/auth'
import { createSurvivorModel } from './models/Survivor'
import { createTrapperModel } from './models/Trapper'
import {
  createGeneratorModel,
  createPalletModel,
  createHookModel,
  createTrapModel,
  createExitGateModel,
  createWindowModel
} from './models/WorldObjects'
import { Materials } from './models/Materials'
import { CharacterAnimator } from './animations/CharacterAnimator'
import { buildAzarovMap } from './map/AzarovRealm'

const MAP_WIDTH = 80
const MAP_HEIGHT = 80

export class GameEngine {
  private renderer!: THREE.WebGLRenderer
  private scene!: THREE.Scene
  private camera!: THREE.PerspectiveCamera
  private container: HTMLElement

  private network: GameNetwork
  private gameStore = useGameStore()
  private authStore = useAuthStore()

  // Player models
  private playerModels: Map<number, THREE.Group> = new Map()
  private playerAnimators: Map<number, CharacterAnimator> = new Map()

  // Object meshes
  private generatorMeshes: Map<string, THREE.Group> = new Map()
  private palletMeshes: Map<string, THREE.Group> = new Map()
  private hookMeshes: Map<string, THREE.Group> = new Map()
  private trapMeshes: Map<string, THREE.Group> = new Map()
  private gateMeshes: Map<string, THREE.Group> = new Map()
  private windowMeshes: Map<string, THREE.Group> = new Map()

  // Scratch marks and blood
  private scratchMeshes: THREE.Mesh[] = []
  private bloodMeshes: THREE.Mesh[] = []

  // Input state
  private keys: Set<string> = new Set()
  private mouseX: number = 0
  private mouseY: number = 0
  private mouseDeltaX: number = 0
  private mouseDeltaY: number = 0
  private isPointerLocked: boolean = false

  // Camera
  private cameraYaw: number = 0
  private cameraPitch: number = 0
  private thirdPersonOffset = new THREE.Vector3(0, 2.5, -3.5)

  // Timing
  private clock: THREE.Clock = new THREE.Clock()
  private lastSendTime: number = 0
  private lastPingTime: number = 0

  // Smooth display position for local POV: lerps toward server position (server remains source of truth)
  private smoothDisplayPos: { x: number; y: number; z: number; rotY: number } | null = null
  private static readonly SMOOTH_POS_K = 20
  private static readonly SMOOTH_ROT_K = 24
  private cameraRaycaster = new THREE.Raycaster()
  private cameraRayOrigin = new THREE.Vector3()
  private cameraRayDir = new THREE.Vector3()

  // Interaction
  private nearbyInteractable: { type: string; id: string; name: string } | null = null

  private running: boolean = false

  constructor(container: HTMLElement) {
    this.container = container
    this.network = new GameNetwork()
  }

  async init() {
    // Setup renderer
    this.renderer = new THREE.WebGLRenderer({ antialias: false })
    this.renderer.setSize(this.container.clientWidth, this.container.clientHeight)
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    this.renderer.shadowMap.enabled = true
    this.renderer.shadowMap.type = THREE.BasicShadowMap
    this.renderer.setClearColor(0x0a0a0a)
    this.container.appendChild(this.renderer.domElement)

    // Setup scene
    this.scene = new THREE.Scene()
    this.scene.fog = new THREE.FogExp2(0x0a0a0a, 0.04)

    // Setup camera
    this.camera = new THREE.PerspectiveCamera(
      this.gameStore.myRole === 'killer' ? 90 : 70,
      this.container.clientWidth / this.container.clientHeight,
      0.1,
      100
    )

    // Lighting
    this.setupLighting()

    // Build map
    buildAzarovMap(this.scene)

    // Setup input
    this.setupInput()

    // Connect to game server
    await this.network.connect(
      this.gameStore.gameId,
      this.gameStore.gamePort,
      this.authStore.token,
      this.authStore.userId
    )

    this.network.setOnDisconnect(() => {
      this.running = false
    })

    // Handle resize
    window.addEventListener('resize', () => this.onResize())

    this.running = true
    this.animate()
  }

  private setupLighting() {
    // Ambient - brighter for visibility while keeping night feel
    const ambient = new THREE.AmbientLight(0x445566, 1.05)
    this.scene.add(ambient)

    // Moonlight - directional from above
    const moon = new THREE.DirectionalLight(0x5588bb, 1.0)
    moon.position.set(30, 50, 30)
    moon.castShadow = true
    moon.shadow.mapSize.width = 2048
    moon.shadow.mapSize.height = 2048
    moon.shadow.camera.near = 0.5
    moon.shadow.camera.far = 100
    moon.shadow.camera.left = -50
    moon.shadow.camera.right = 50
    moon.shadow.camera.top = 50
    moon.shadow.camera.bottom = -50
    this.scene.add(moon)

    // Hemisphere for sky/ground color
    const hemi = new THREE.HemisphereLight(0x334455, 0x2a2a2a, 0.58)
    this.scene.add(hemi)
  }

  private setupInput() {
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey || e.metaKey) e.preventDefault()
      this.keys.add(e.code)

      // Interaction keys
      if (e.code === 'KeyE') this.handleInteractPress()
      if (e.code === 'Space') this.handleSpacePress()
      if (e.code === 'KeyQ') this.handleQPress()
      if (e.code === 'KeyR' && !e.repeat) this.handleRPress()
    })

    document.addEventListener('keyup', (e) => {
      this.keys.delete(e.code)

      if (e.code === 'KeyE') this.handleInteractRelease()
      if (e.code === 'KeyR') this.network.sendStopInteract()
    })

    // Mouse for camera control
    document.addEventListener('mousemove', (e) => {
      if (this.isPointerLocked) {
        this.mouseDeltaX += e.movementX
        this.mouseDeltaY += e.movementY
      }
    })

    // Left click for attack (killer)
    document.addEventListener('mousedown', (e) => {
      if (e.button === 0) {
        if (!this.isPointerLocked) {
          this.container.requestPointerLock()
        } else {
          this.handleLeftClick()
        }
      }
      if (e.button === 2) {
        this.handleRightClick()
      }
    })

    document.addEventListener('contextmenu', (e) => e.preventDefault())

    document.addEventListener('pointerlockchange', () => {
      this.isPointerLocked = document.pointerLockElement === this.container
    })

    // Click to lock pointer
    this.container.addEventListener('click', () => {
      if (!this.isPointerLocked) {
        this.container.requestPointerLock()
      }
    })
  }

  private handleInteractPress() {
    if (!this.nearbyInteractable) return

    const { type, id } = this.nearbyInteractable
    switch (type) {
      case 'generator':
        this.network.sendInteract('repair', id)
        break
      case 'gate':
        this.network.sendInteract('open_gate', id)
        break
      case 'hook_occupied':
        this.network.sendInteract('unhook', id)
        break
      case 'trap_pickup':
        this.network.sendInteract('pickup_trap', id)
        break
      case 'survivor_dying':
        this.network.sendInteract('pickup')
        break
      case 'hook_empty':
        this.network.sendInteract('hook', id)
        break
      case 'gen_kick':
        this.network.sendInteract('kick_gen', id)
        break
      case 'pallet_break':
        this.network.sendInteract('break_pallet', id)
        break
      case 'heal':
        this.network.sendInteract('heal', id)
        break
    }
  }

  private handleInteractRelease() {
    this.network.sendStopInteract()
  }

  private handleSpacePress() {
    const me = this.gameStore.getMyPlayer()
    if (!me) return

    // Vault or drop pallet
    if (!this.nearbyInteractable) return

    if (this.nearbyInteractable.type === 'pallet' && me.role === 'survivor') {
      this.network.sendInteract('drop_pallet', this.nearbyInteractable.id)
    } else if (this.nearbyInteractable.type === 'window' || this.nearbyInteractable.type === 'pallet_vault') {
      this.network.sendInteract('vault', this.nearbyInteractable.id)
    }

  }

  private handleQPress() {
    const me = this.gameStore.getMyPlayer()
    if (!me || me.role !== 'killer') return

    if (me.carrying_id > 0) {
      this.network.sendInteract('drop_survivor')
    }
  }

  private handleRPress() {
    const me = this.gameStore.getMyPlayer()
    if (!me || me.role !== 'killer') return

    if (me.trap_count > 0) {
      this.network.sendInteract('place_trap')
    }
  }

  private handleLeftClick() {
    const me = this.gameStore.getMyPlayer()
    if (!me) return

    if (me.role === 'killer') {
      this.network.sendInteract('attack')
    }
  }

  private handleRightClick() {
    // Could be used for secondary actions
  }

  private animate() {
    if (!this.running) return
    requestAnimationFrame(() => this.animate())

    const dt = this.clock.getDelta()
    const now = performance.now()

    // Process camera rotation from mouse
    this.updateCamera(dt)

    // Process movement input and send to server
    this.updateMovement(dt, now)

    // Update game objects from state
    this.updateGameObjects(dt)

    // Check nearby interactables
    this.updateNearbyInteractable()

    // Ping
    if (now - this.lastPingTime > 2000) {
      this.network.sendPing()
      this.lastPingTime = now
    }

    // Render
    this.renderer.render(this.scene, this.camera)
  }

  private updateCamera(dt: number) {
    const sensitivity = 0.002

    this.cameraYaw += this.mouseDeltaX * sensitivity
    this.cameraPitch -= this.mouseDeltaY * sensitivity
    this.cameraPitch = Math.max(-Math.PI / 3, Math.min(Math.PI / 3, this.cameraPitch))

    this.mouseDeltaX = 0
    this.mouseDeltaY = 0

    const me = this.gameStore.getMyPlayer()
    if (!me) return

    const myModel = this.playerModels.get(me.user_id)

    // Lerp smooth display position toward server (source of truth) for extremely smooth POV
    const serverPos = { x: me.pos_x, y: me.pos_y, z: me.pos_z, rotY: me.rot_y }
    if (!this.smoothDisplayPos) {
      this.smoothDisplayPos = { ...serverPos }
    } else {
      const tPos = 1 - Math.exp(-GameEngine.SMOOTH_POS_K * dt)
      const tRot = 1 - Math.exp(-GameEngine.SMOOTH_ROT_K * dt)
      this.smoothDisplayPos.x += (serverPos.x - this.smoothDisplayPos.x) * tPos
      this.smoothDisplayPos.y += (serverPos.y - this.smoothDisplayPos.y) * tPos
      this.smoothDisplayPos.z += (serverPos.z - this.smoothDisplayPos.z) * tPos
      let rotDiff = serverPos.rotY - this.smoothDisplayPos.rotY
      while (rotDiff > Math.PI) rotDiff -= 2 * Math.PI
      while (rotDiff < -Math.PI) rotDiff += 2 * Math.PI
      this.smoothDisplayPos.rotY += rotDiff * tRot
    }
    const pos = this.smoothDisplayPos

    if (me.role === 'killer') {
      // First person camera (smooth position, server is source of truth)
      this.camera.position.set(pos.x, pos.y + 1.7, pos.z)

      const lookDir = new THREE.Vector3(
        Math.sin(this.cameraYaw) * Math.cos(this.cameraPitch),
        Math.sin(this.cameraPitch),
        Math.cos(this.cameraYaw) * Math.cos(this.cameraPitch)
      )
      this.camera.lookAt(
        this.camera.position.x + lookDir.x,
        this.camera.position.y + lookDir.y,
        this.camera.position.z + lookDir.z
      )

      // Hide own model in first person
      if (myModel) myModel.visible = false
    } else {
      // Third person camera: reduced distance, and collision so camera doesn't go through walls
      const playerPos = new THREE.Vector3(pos.x, pos.y, pos.z)
      const lookAtY = 1.2
      const horizDist = 2.4
      const vertBase = 2.0
      const vertSwing = 1.2
      const offset = new THREE.Vector3(
        -Math.sin(this.cameraYaw) * horizDist,
        vertBase + Math.sin(this.cameraPitch) * vertSwing,
        -Math.cos(this.cameraYaw) * horizDist
      )
      const desiredCamPos = playerPos.clone().add(offset)
      this.cameraRayOrigin.set(playerPos.x, playerPos.y + lookAtY, playerPos.z)
      this.cameraRayDir.copy(desiredCamPos).sub(this.cameraRayOrigin).normalize()
      const maxCamDist = this.cameraRayOrigin.distanceTo(desiredCamPos)
      this.cameraRaycaster.set(this.cameraRayOrigin, this.cameraRayDir)
      this.cameraRaycaster.camera = this.camera
      this.cameraRaycaster.far = maxCamDist + 0.5
      const playerRoots = new Set(this.playerModels.values())
      const hits = this.cameraRaycaster.intersectObject(this.scene, true)
      let useDist = maxCamDist
      const minHitY = playerPos.y + 0.3
      for (const hit of hits) {
        if (hit.object instanceof THREE.Sprite) continue
        if (hit.point.y < minHitY) continue
        let obj: THREE.Object3D | null = hit.object
        while (obj) {
          if (playerRoots.has(obj as THREE.Group)) break
          obj = obj.parent
        }
        if (obj) continue
        const d = hit.distance
        if (d < useDist) useDist = Math.max(0.2, d - 0.1)
        break
      }
      this.camera.position.copy(this.cameraRayOrigin).addScaledVector(this.cameraRayDir, useDist)
      this.camera.lookAt(playerPos.x, playerPos.y + lookAtY, playerPos.z)

      // Show own model
      if (myModel) myModel.visible = true
    }
  }

  private updateMovement(dt: number, now: number) {
    const me = this.gameStore.getMyPlayer()
    if (!me || !me.is_alive || me.has_escaped) return

    // Can't move if doing certain actions
    if (['hooked', 'trapped', 'being_carried', 'stunned'].includes(me.action_state)) return

    let moveState = 'idle'
    let dx = 0, dz = 0

    // Movement relative to camera direction
    const forward = new THREE.Vector3(Math.sin(this.cameraYaw), 0, Math.cos(this.cameraYaw))
    const right = new THREE.Vector3(Math.cos(this.cameraYaw), 0, -Math.sin(this.cameraYaw))

    if (this.keys.has('KeyW')) { dx += forward.x; dz += forward.z }
    if (this.keys.has('KeyS')) { dx -= forward.x; dz -= forward.z }
    if (this.keys.has('KeyA')) { dx += right.x; dz += right.z }
    if (this.keys.has('KeyD')) { dx -= right.x; dz -= right.z }

    // Normalize
    const len = Math.sqrt(dx * dx + dz * dz)
    if (len > 0) {
      dx /= len
      dz /= len

      if (me.role === 'survivor') {
        if (this.keys.has('ControlLeft') || this.keys.has('ControlRight')) {
          moveState = 'crouching'
        } else if (this.keys.has('ShiftLeft') || this.keys.has('ShiftRight')) {
          moveState = 'running'
        } else {
          moveState = 'walking'
        }
      } else {
        moveState = 'running' // Killer always runs
      }

      // Calculate speed (must match game-server/internal/player/player.go) — 3x multiplier, then 1.5x
      let speed = 63.0 // default run
      if (me.role === 'survivor') {
        if (moveState === 'running') speed = 63.0
        else if (moveState === 'walking') speed = 36.0
        else if (moveState === 'crouching') speed = 18.0
      } else {
        speed = me.carrying_id > 0 ? 66.24 : 72.0
      }

      const newX = me.pos_x + dx * speed * dt
      const newZ = me.pos_z + dz * speed * dt

      // Send to server at tick rate (45 Hz for smoother input)
      if (now - this.lastSendTime > 1000 / 45) {
        this.network.sendMove(newX, me.pos_y, newZ, this.cameraYaw, moveState)
        this.lastSendTime = now
      }
    } else {
      if (now - this.lastSendTime > 200) {
        this.network.sendMove(me.pos_x, me.pos_y, me.pos_z, this.cameraYaw, 'idle')
        this.lastSendTime = now
      }
    }
  }

  private updateGameObjects(dt: number) {
    const state = this.gameStore.gameState
    if (!state) return

    // Update players
    for (const p of state.players) {
      this.updatePlayerModel(p, dt)
    }

    // Remove disconnected player models
    for (const [uid, model] of this.playerModels) {
      if (!state.players.find(p => p.user_id === uid)) {
        this.scene.remove(model)
        this.playerModels.delete(uid)
        this.playerAnimators.delete(uid)
      }
    }

    // Update generators
    for (const gen of state.generators) {
      this.updateGenerator(gen)
    }

    // Update pallets
    for (const pal of state.pallets) {
      this.updatePallet(pal)
    }

    // Update hooks
    for (const hook of state.hooks) {
      this.updateHook(hook)
    }

    // Update traps
    for (const trap of state.traps) {
      this.updateTrap(trap)
    }

    // Update gates
    for (const gate of state.gates) {
      this.updateGate(gate)
    }

    // Update windows (static, just ensure they exist)
    for (const win of state.windows) {
      this.ensureWindow(win)
    }

    // Update scratch marks
    this.updateScratchMarks(state.scratch_marks || [])

    // Update blood trails
    this.updateBloodTrails(state.blood_trails || [])
  }

  private updatePlayerModel(p: PlayerState, dt: number) {
    let model = this.playerModels.get(p.user_id)
    let animator = this.playerAnimators.get(p.user_id)

    if (!model) {
      // Create new model
      model = p.role === 'killer' ? createTrapperModel() : createSurvivorModel()
      animator = new CharacterAnimator(model)
      this.playerModels.set(p.user_id, model)
      this.playerAnimators.set(p.user_id, animator)
      this.scene.add(model)

      // Add name label (floating text isn't easy in Three.js, so we'll use a small indicator)
      const labelSprite = this.createNameSprite(p.username)
      labelSprite.position.set(0, p.role === 'killer' ? 2.3 : 2.0, 0)
      labelSprite.name = 'nameLabel'
      model.add(labelSprite)
    }

    // Feet offset: model origin is at waist; bottom of feet ~0.24 (survivor) / 0.2 (killer)
    const feetOffset = p.role === 'killer' ? 0.2 : 0.24
    const dyingFloorOffset = p.action_state === 'dying' ? 0.2 : 0

    // Local player: use smooth display position (lerped toward server in updateCamera); remote: use server position
    const isLocal = p.user_id === this.authStore.userId
    const displayX = isLocal && this.smoothDisplayPos ? this.smoothDisplayPos.x : p.pos_x
    const displayY = isLocal && this.smoothDisplayPos ? this.smoothDisplayPos.y : p.pos_y
    const displayZ = isLocal && this.smoothDisplayPos ? this.smoothDisplayPos.z : p.pos_z
    let displayRotY = isLocal && this.smoothDisplayPos ? this.smoothDisplayPos.rotY : p.rot_y
    const displayYFeet = displayY - feetOffset + dyingFloorOffset

    // When repairing, face the generator (client-side display; server also sets RotY in handleRepair)
    if (p.action_state === 'repairing' && p.action_target) {
      const state = this.gameStore.gameState
      const gen = state?.generators.find((g) => g.id === p.action_target)
      if (gen) {
        displayRotY = Math.atan2(gen.pos_x - displayX, gen.pos_z - displayZ)
      }
    }

    const lerpFactor = 1 - Math.exp(-GameEngine.SMOOTH_POS_K * dt)
    model.position.x += (displayX - model.position.x) * lerpFactor
    model.position.y += (displayYFeet - model.position.y) * lerpFactor
    model.position.z += (displayZ - model.position.z) * lerpFactor
    let angleDiff = displayRotY - model.rotation.y
    while (angleDiff > Math.PI) angleDiff -= 2 * Math.PI
    while (angleDiff < -Math.PI) angleDiff += 2 * Math.PI
    model.rotation.y += angleDiff * lerpFactor

    // Update animation
    if (animator) {
      const isInjured = p.role === 'survivor' && p.health === 2
      animator.update(dt, p.move_state, p.action_state, p.role === 'killer', isInjured)
    }

    // Death / escape visibility
    if (!p.is_alive || p.has_escaped) {
      model.visible = false
    } else {
      // Only hide own model if killer first person
      if (p.user_id === this.authStore.userId && p.role === 'killer') {
        model.visible = false
      } else {
        model.visible = true
      }
    }

    // Injured tint
    if (p.role === 'survivor' && p.health === 2) {
      // Could tint model red slightly
    }
  }

  private createNameSprite(name: string): THREE.Sprite {
    const canvas = document.createElement('canvas')
    canvas.width = 256
    canvas.height = 64
    const ctx = canvas.getContext('2d')!
    ctx.fillStyle = 'rgba(0,0,0,0.5)'
    ctx.fillRect(0, 0, 256, 64)
    ctx.fillStyle = 'white'
    ctx.font = 'bold 28px monospace'
    ctx.textAlign = 'center'
    ctx.fillText(name, 128, 40)

    const texture = new THREE.CanvasTexture(canvas)
    const material = new THREE.SpriteMaterial({ map: texture, transparent: true })
    const sprite = new THREE.Sprite(material)
    sprite.scale.set(1.5, 0.4, 1)
    return sprite
  }

  private updateGenerator(gen: any) {
    let mesh = this.generatorMeshes.get(gen.id)
    if (!mesh) {
      mesh = createGeneratorModel()
      mesh.position.set(gen.pos_x, gen.pos_y, gen.pos_z)
      this.generatorMeshes.set(gen.id, mesh)
      this.scene.add(mesh)

      // Add point light
      const light = new THREE.PointLight(gen.done ? 0x22aa22 : 0xaa6622, 0.8, 8)
      light.position.set(0, 1.2, 0)
      light.name = 'genLight'
      mesh.add(light)

      // World-space progress bar (track + fill)
      const track = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.06, 0.04), Materials.metalDark)
      track.position.set(0, 1.35, 0.4)
      track.name = 'progressTrack'
      mesh.add(track)
      const fill = new THREE.Mesh(new THREE.BoxGeometry(0.76, 0.05, 0.02), Materials.genRed)
      fill.position.set(-0.38, 0, 0)
      fill.name = 'progressFill'
      track.add(fill)
    }

    // Update indicator color
    const indicator = mesh.getObjectByName('indicator') as THREE.Mesh
    if (indicator) {
      (indicator as THREE.Mesh).material = gen.done ? Materials.genGreen : Materials.genRed
    }

    // Update light color
    const light = mesh.getObjectByName('genLight') as THREE.PointLight
    if (light) {
      if (gen.done) {
        light.color.setHex(0x22aa22)
        light.intensity = 1.2
      } else if (gen.being_repaired) {
        light.color.setHex(0xaaaa22)
        light.intensity = 0.6 + Math.sin(performance.now() * 0.005) * 0.3
      } else {
        light.color.setHex(0xaa6622)
        light.intensity = 0.3
      }
    }

    // Animate pistons when being repaired
    if (gen.being_repaired && !gen.done) {
      mesh.traverse((child) => {
        if (child.name === 'piston') {
          child.position.y = 0.95 + Math.sin(performance.now() * 0.01 + child.id) * 0.05
        }
      })
    }

    // Progress bar visibility and fill (scale fill in X by progress; center so it grows rightward)
    const track = mesh.getObjectByName('progressTrack')
    const fill = mesh.getObjectByName('progressFill')
    if (track && fill) {
      if (gen.being_repaired && !gen.done) {
        track.visible = true
        fill.visible = true
        const p = Math.max(0, Math.min(1, gen.progress ?? 0))
        fill.scale.x = p
        fill.position.x = -0.38 + p * 0.38
      } else if (gen.done) {
        track.visible = true
        fill.visible = true
        fill.scale.x = 1
        fill.position.x = 0
        if (fill instanceof THREE.Mesh && fill.material) {
          (fill.material as THREE.MeshLambertMaterial).color?.setHex(0x22aa22)
        }
      } else {
        track.visible = false
        fill.visible = false
      }
    }
  }

  private updatePallet(pal: any) {
    let mesh = this.palletMeshes.get(pal.id)
    if (!mesh) {
      mesh = createPalletModel()
      mesh.position.set(pal.pos_x, pal.pos_y, pal.pos_z)
      mesh.rotation.y = pal.rot_y
      this.palletMeshes.set(pal.id, mesh)
      this.scene.add(mesh)
    }

    if (pal.broken) {
      mesh.visible = false
    } else {
      mesh.visible = true
      mesh.rotation.y = pal.rot_y
      if (pal.dropped) {
        mesh.rotation.x = Math.PI / 2
        mesh.position.set(pal.pos_x, 0.15, pal.pos_z)
      } else {
        mesh.rotation.x = 0
        mesh.position.set(pal.pos_x, pal.pos_y, pal.pos_z)
      }
    }
  }

  private updateHook(hook: any) {
    let mesh = this.hookMeshes.get(hook.id)
    if (!mesh) {
      mesh = createHookModel()
      mesh.position.set(hook.pos_x, hook.pos_y, hook.pos_z)
      this.hookMeshes.set(hook.id, mesh)
      this.scene.add(mesh)

      // Hook light (matches smaller hook model height)
      const light = new THREE.PointLight(0xaa2222, 0.5, 6)
      light.position.set(0, 2.4, 0)
      mesh.add(light)
    }

    // Indicator glow when occupied
    const indicator = mesh.getObjectByName('indicator') as THREE.Mesh
    if (indicator) {
      if (hook.occupied) {
        (indicator as THREE.Mesh).material = Materials.redLight
      } else {
        (indicator as THREE.Mesh).material = Materials.red
      }
    }
  }

  private updateTrap(trap: any) {
    let mesh = this.trapMeshes.get(trap.id)
    if (!mesh) {
      mesh = createTrapModel()
      this.trapMeshes.set(trap.id, mesh)
      this.scene.add(mesh)
    }

    mesh.position.set(trap.pos_x, trap.pos_y, trap.pos_z)

    if (!trap.placed) {
      mesh.visible = false
    } else {
      mesh.visible = true

      // Animate triggered trap (jaws closed)
      if (trap.triggered) {
        const jawL = mesh.getObjectByName('jawL')
        const jawR = mesh.getObjectByName('jawR')
        if (jawL) jawL.position.z = -0.05
        if (jawR) jawR.position.z = 0.05
      }
    }
  }

  private updateGate(gate: any) {
    let mesh = this.gateMeshes.get(gate.id)
    if (!mesh) {
      mesh = createExitGateModel()
      mesh.position.set(gate.pos_x, gate.pos_y, gate.pos_z)
      mesh.rotation.y = gate.rot_y
      this.gateMeshes.set(gate.id, mesh)
      this.scene.add(mesh)
    }

    // Animate door sliding when open
    const door = mesh.getObjectByName('door') as THREE.Mesh
    if (door && gate.open) {
      door.position.y = Math.min(door.position.y + 0.05, 6.0) // Slide up
    }

    // Update lights based on progress
    for (let i = 0; i < 3; i++) {
      const light = mesh.getObjectByName(`light_${i}`) as THREE.Mesh
      if (light) {
        const threshold = (i + 1) / 3
        ;(light as THREE.Mesh).material = gate.progress >= threshold ? Materials.genGreen : Materials.genRed
      }
    }
  }

  private ensureWindow(win: any) {
    let mesh = this.windowMeshes.get(win.id)
    if (!mesh) {
      mesh = createWindowModel()
      mesh.position.set(win.pos_x, win.pos_y, win.pos_z)
      mesh.rotation.y = win.rot_y
      this.windowMeshes.set(win.id, mesh)
      this.scene.add(mesh)
    }
  }

  private updateScratchMarks(marks: any[]) {
    // Remove old
    for (const m of this.scratchMeshes) {
      this.scene.remove(m)
    }
    this.scratchMeshes = []

    // Only show to killer
    if (this.gameStore.myRole !== 'killer') return

    for (const mark of marks) {
      const mesh = new THREE.Mesh(
        new THREE.PlaneGeometry(0.3, 0.3),
        Materials.scratchMark.clone()
      )
      mesh.rotation.x = -Math.PI / 2
      mesh.position.set(mark.pos_x, 0.02, mark.pos_z)
      const mat = mesh.material as THREE.MeshBasicMaterial
      mat.opacity = Math.max(0, 1 - mark.age / 7) * 0.6
      this.scene.add(mesh)
      this.scratchMeshes.push(mesh)
    }
  }

  private updateBloodTrails(trails: any[]) {
    for (const m of this.bloodMeshes) {
      this.scene.remove(m)
    }
    this.bloodMeshes = []

    for (const trail of trails) {
      const mesh = new THREE.Mesh(
        new THREE.PlaneGeometry(0.2, 0.2),
        Materials.bloodTrail.clone()
      )
      mesh.rotation.x = -Math.PI / 2
      mesh.position.set(trail.pos_x, 0.01, trail.pos_z)
      const mat = mesh.material as THREE.MeshBasicMaterial
      mat.opacity = Math.max(0, 1 - trail.age / 10) * 0.7
      this.scene.add(mesh)
      this.bloodMeshes.push(mesh)
    }
  }

  private updateNearbyInteractable() {
    const me = this.gameStore.getMyPlayer()
    const state = this.gameStore.gameState
    if (!me || !state) {
      this.nearbyInteractable = null
      return
    }

    const interactDist = 1.6
    this.nearbyInteractable = null

    if (me.role === 'survivor') {
      // Check generators
      for (const gen of state.generators) {
        if (!gen.done && this.dist2d(me, gen) < interactDist) {
          this.nearbyInteractable = { type: 'generator', id: gen.id, name: 'Repair Generator [E]' }
          return
        }
      }

      // Check gates
      if (state.gates_powered) {
        for (const gate of state.gates) {
          if (!gate.open && gate.powered && this.dist2d(me, gate) < interactDist + 0.6) {
            this.nearbyInteractable = { type: 'gate', id: gate.id, name: 'Open Gate [E]' }
            return
          }
        }
      }

      // Check hooks with survivors
      for (const hook of state.hooks) {
        if (hook.occupied && this.dist2d(me, hook) < interactDist) {
          this.nearbyInteractable = { type: 'hook_occupied', id: hook.id, name: 'Unhook [E]' }
          return
        }
      }

      // Check pallets
      for (const pal of state.pallets) {
        if (!pal.dropped && !pal.broken && this.dist2d(me, pal) < interactDist) {
          this.nearbyInteractable = { type: 'pallet', id: pal.id, name: 'Drop Pallet [Space]' }
          return
        }
        if (pal.dropped && !pal.broken && this.dist2d(me, pal) < interactDist) {
          this.nearbyInteractable = { type: 'pallet_vault', id: pal.id, name: 'Vault Pallet [Space]' }
          return
        }
      }

      // Check windows
      for (const win of state.windows) {
        if (this.dist2d(me, win) < interactDist) {
          this.nearbyInteractable = { type: 'window', id: win.id, name: 'Vault [Space]' }
          return
        }
      }

      // Check injured survivors for healing
      for (const p of state.players) {
        if (p.role === 'survivor' && p.user_id !== me.user_id && p.health === 2 && p.is_alive) {
          if (this.dist2dP(me, p) < interactDist) {
            this.nearbyInteractable = { type: 'heal', id: p.user_id.toString(), name: `Heal ${p.username} [E]` }
            return
          }
        }
      }
    } else {
      // Killer interactions
      // Pickup dying survivor (not hooked)
      for (const p of state.players) {
        if (p.role === 'survivor' && p.health === 1 && p.is_alive && p.action_state !== 'being_carried' && p.action_state !== 'hooked') {
          if (this.dist2dP(me, p) < interactDist) {
            this.nearbyInteractable = { type: 'survivor_dying', id: p.user_id.toString(), name: `Pickup ${p.username} [E]` }
            return
          }
        }
      }

      // Hook (when carrying)
      if (me.carrying_id > 0) {
        for (const hook of state.hooks) {
          if (!hook.occupied && this.dist2d(me, hook) < interactDist) {
            this.nearbyInteractable = { type: 'hook_empty', id: hook.id, name: 'Hook Survivor [E]' }
            return
          }
        }
      }

      // Break pallet
      for (const pal of state.pallets) {
        if (pal.dropped && !pal.broken && this.dist2d(me, pal) < interactDist) {
          this.nearbyInteractable = { type: 'pallet_break', id: pal.id, name: 'Break Pallet [E]' }
          return
        }
      }

      // Kick gen
      for (const gen of state.generators) {
        if (!gen.done && gen.progress > 0 && this.dist2d(me, gen) < interactDist) {
          this.nearbyInteractable = { type: 'gen_kick', id: gen.id, name: 'Kick Generator [E]' }
          return
        }
      }

      // Pickup trap
      for (const trap of state.traps) {
        if (trap.placed && !trap.triggered && this.dist2d(me, trap) < interactDist) {
          this.nearbyInteractable = { type: 'trap_pickup', id: trap.id, name: 'Pickup Trap [E]' }
          return
        }
      }

      // Vault
      for (const win of state.windows) {
        if (this.dist2d(me, win) < interactDist) {
          this.nearbyInteractable = { type: 'window', id: win.id, name: 'Vault [Space]' }
          return
        }
      }
    }
  }

  private dist2d(me: any, obj: any): number {
    const dx = me.pos_x - obj.pos_x
    const dz = me.pos_z - obj.pos_z
    return Math.sqrt(dx * dx + dz * dz)
  }

  private dist2dP(me: any, obj: any): number {
    const dx = me.pos_x - obj.pos_x
    const dz = me.pos_z - obj.pos_z
    return Math.sqrt(dx * dx + dz * dz)
  }

  getNearbyInteractable() {
    return this.nearbyInteractable
  }

  private onResize() {
    const w = this.container.clientWidth
    const h = this.container.clientHeight
    this.camera.aspect = w / h
    this.camera.updateProjectionMatrix()
    this.renderer.setSize(w, h)
  }

  destroy() {
    this.running = false
    this.smoothDisplayPos = null
    this.network.disconnect()

    // Clean up Three.js
    this.renderer.dispose()
    this.scene.traverse((obj) => {
      if (obj instanceof THREE.Mesh) {
        obj.geometry.dispose()
        if (Array.isArray(obj.material)) {
          obj.material.forEach(m => m.dispose())
        } else {
          obj.material.dispose()
        }
      }
    })

    // Remove canvas
    if (this.renderer.domElement.parentElement) {
      this.renderer.domElement.parentElement.removeChild(this.renderer.domElement)
    }

    // Remove event listeners
    document.exitPointerLock()
  }
}
