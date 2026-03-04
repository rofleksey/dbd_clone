import * as THREE from 'three'
import { Materials } from '../models/Materials'

const MAP_W = 80
const MAP_H = 80
const WALL_H = 3
const SECOND_FLOOR_Y = 3.5

// Build the Azarov's Realm map geometry
export function buildAzarovMap(scene: THREE.Scene) {
  // Ground plane
  buildGround(scene)

  // Map boundary walls (invisible but collidable in server, visual fencing here)
  buildBoundaryFence(scene)

  // Main building
  buildMainBuilding(scene)

  // Killer Shack
  buildShack(scene)

  // Hill
  buildHill(scene)

  // Debris walls / loop walls
  buildLoopWalls(scene)

  // Scatter environmental props
  buildEnvironment(scene)
}

function buildGround(scene: THREE.Scene) {
  // Create a textured ground using a canvas for pixel-art dirt/grass
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 256
  const ctx = canvas.getContext('2d')!

  // Base dirt color
  ctx.fillStyle = '#3a2a1a'
  ctx.fillRect(0, 0, 256, 256)

  // Add grass patches
  for (let i = 0; i < 2000; i++) {
    const x = Math.random() * 256
    const y = Math.random() * 256
    const shade = Math.random() * 0.3
    ctx.fillStyle = `rgb(${30 + shade * 40}, ${50 + shade * 60}, ${20 + shade * 30})`
    ctx.fillRect(x, y, 2 + Math.random() * 4, 2 + Math.random() * 4)
  }

  // Add dirt patches
  for (let i = 0; i < 500; i++) {
    const x = Math.random() * 256
    const y = Math.random() * 256
    ctx.fillStyle = `rgb(${50 + Math.random() * 30}, ${35 + Math.random() * 20}, ${20 + Math.random() * 15})`
    ctx.fillRect(x, y, 3 + Math.random() * 6, 3 + Math.random() * 6)
  }

  const texture = new THREE.CanvasTexture(canvas)
  texture.wrapS = THREE.RepeatWrapping
  texture.wrapT = THREE.RepeatWrapping
  texture.repeat.set(16, 16)
  texture.magFilter = THREE.NearestFilter
  texture.minFilter = THREE.NearestFilter

  const groundMat = new THREE.MeshLambertMaterial({ map: texture })
  const ground = new THREE.Mesh(new THREE.PlaneGeometry(MAP_W, MAP_H), groundMat)
  ground.rotation.x = -Math.PI / 2
  ground.position.set(MAP_W / 2, 0, MAP_H / 2)
  ground.receiveShadow = true
  scene.add(ground)
}

function buildBoundaryFence(scene: THREE.Scene) {
  const fenceMat = Materials.metalDark

  // Create fence posts around perimeter
  const spacing = 4
  for (let i = 0; i < MAP_W; i += spacing) {
    addFencePost(scene, i, 0, fenceMat)
    addFencePost(scene, i, MAP_H, fenceMat)
  }
  for (let i = 0; i < MAP_H; i += spacing) {
    addFencePost(scene, 0, i, fenceMat)
    addFencePost(scene, MAP_W, i, fenceMat)
  }

  // Horizontal bars
  addWall(scene, -0.5, 0.5, 0, MAP_W + 0.5, 1.0, 0.15, Materials.metalDark)
  addWall(scene, -0.5, 1.5, 0, MAP_W + 0.5, 2.0, 0.15, Materials.metalDark)
  addWall(scene, -0.5, 0.5, MAP_H - 0.15, MAP_W + 0.5, 1.0, MAP_H, Materials.metalDark)
  addWall(scene, -0.5, 1.5, MAP_H - 0.15, MAP_W + 0.5, 2.0, MAP_H, Materials.metalDark)
  addWall(scene, 0, 0.5, 0, 0.15, 1.0, MAP_H, Materials.metalDark)
  addWall(scene, 0, 1.5, 0, 0.15, 2.0, MAP_H, Materials.metalDark)
  addWall(scene, MAP_W - 0.15, 0.5, 0, MAP_W, 1.0, MAP_H, Materials.metalDark)
  addWall(scene, MAP_W - 0.15, 1.5, 0, MAP_W, 2.0, MAP_H, Materials.metalDark)
}

function addFencePost(scene: THREE.Scene, x: number, z: number, mat: THREE.Material) {
  const post = new THREE.Mesh(new THREE.BoxGeometry(0.15, 2.5, 0.15), mat)
  post.position.set(x, 1.25, z)
  post.castShadow = true
  scene.add(post)
}

function buildMainBuilding(scene: THREE.Scene) {
  const mbX = 45, mbZ = 30
  const mbW = 20, mbD = 18
  const wt = 0.5 // wall thickness

  const wallMat = Materials.concrete
  const wallDarkMat = Materials.concreteDark
  const floorMat = Materials.concreteDark

  // Ground floor
  // South wall with door gap
  addWall(scene, mbX, 0, mbZ, mbX + 8, WALL_H, mbZ + wt, wallMat)
  addWall(scene, mbX + 10, 0, mbZ, mbX + mbW, WALL_H, mbZ + wt, wallMat)

  // North wall with door gap
  addWall(scene, mbX, 0, mbZ + mbD - wt, mbX + 7, WALL_H, mbZ + mbD, wallMat)
  addWall(scene, mbX + 9, 0, mbZ + mbD - wt, mbX + mbW, WALL_H, mbZ + mbD, wallMat)

  // West wall
  addWall(scene, mbX, 0, mbZ, mbX + wt, WALL_H, mbZ + mbD, wallMat)

  // East wall with window gap
  addWall(scene, mbX + mbW - wt, 0, mbZ, mbX + mbW, WALL_H, mbZ + 6, wallMat)
  addWall(scene, mbX + mbW - wt, 0, mbZ + 8, mbX + mbW, WALL_H, mbZ + mbD, wallMat)

  // Internal dividing wall
  addWall(scene, mbX + 10 - wt / 2, 0, mbZ, mbX + 10 + wt / 2, WALL_H, mbZ + 7, wallDarkMat)
  addWall(scene, mbX + 10 - wt / 2, 0, mbZ + 9, mbX + 10 + wt / 2, WALL_H, mbZ + mbD, wallDarkMat)

  // Ground floor floor slab (concrete)
  const gFloor = new THREE.Mesh(new THREE.BoxGeometry(mbW, 0.1, mbD), floorMat)
  gFloor.position.set(mbX + mbW / 2, 0.05, mbZ + mbD / 2)
  gFloor.receiveShadow = true
  scene.add(gFloor)

  // Second floor slab
  const sFloor = new THREE.Mesh(new THREE.BoxGeometry(mbW, 0.2, mbD), floorMat)
  sFloor.position.set(mbX + mbW / 2, SECOND_FLOOR_Y - 0.1, mbZ + mbD / 2)
  sFloor.receiveShadow = true
  sFloor.castShadow = true
  scene.add(sFloor)

  // Second floor walls
  addWall(scene, mbX, SECOND_FLOOR_Y, mbZ, mbX + mbW, SECOND_FLOOR_Y + WALL_H, mbZ + wt, wallMat)
  addWall(scene, mbX, SECOND_FLOOR_Y, mbZ + mbD - wt, mbX + 6, SECOND_FLOOR_Y + WALL_H, mbZ + mbD, wallMat)
  addWall(scene, mbX + 8, SECOND_FLOOR_Y, mbZ + mbD - wt, mbX + mbW, SECOND_FLOOR_Y + WALL_H, mbZ + mbD, wallMat)
  addWall(scene, mbX, SECOND_FLOOR_Y, mbZ, mbX + wt, SECOND_FLOOR_Y + WALL_H, mbZ + mbD, wallMat)
  addWall(scene, mbX + mbW - wt, SECOND_FLOOR_Y, mbZ, mbX + mbW, SECOND_FLOOR_Y + WALL_H, mbZ + mbD, wallMat)

  // Roof
  const roof = new THREE.Mesh(new THREE.BoxGeometry(mbW + 1, 0.3, mbD + 1), wallDarkMat)
  roof.position.set(mbX + mbW / 2, SECOND_FLOOR_Y + WALL_H + 0.15, mbZ + mbD / 2)
  roof.castShadow = true
  roof.receiveShadow = true
  scene.add(roof)

  // Stairs (ramp visual) from ground to second floor
  const stairSteps = 8
  for (let i = 0; i < stairSteps; i++) {
    const stepH = (SECOND_FLOOR_Y / stairSteps)
    const step = new THREE.Mesh(
      new THREE.BoxGeometry(3, stepH, (mbD - 14) / stairSteps),
      Materials.concreteDark
    )
    step.position.set(
      mbX + 2.5,
      stepH * i + stepH / 2,
      mbZ + 12 + (i + 0.5) * ((mbD - 13) / stairSteps)
    )
    step.castShadow = true
    step.receiveShadow = true
    scene.add(step)
  }

  // Interior lights
  const interiorLight1 = new THREE.PointLight(0xaa7744, 0.6, 15)
  interiorLight1.position.set(mbX + 5, 2.5, mbZ + 9)
  scene.add(interiorLight1)

  const interiorLight2 = new THREE.PointLight(0xaa7744, 0.6, 15)
  interiorLight2.position.set(mbX + 15, 2.5, mbZ + 9)
  scene.add(interiorLight2)

  const interiorLight3 = new THREE.PointLight(0xaa7744, 0.4, 12)
  interiorLight3.position.set(mbX + 10, SECOND_FLOOR_Y + 2, mbZ + 9)
  scene.add(interiorLight3)
}

function buildShack(scene: THREE.Scene) {
  const skX = 8, skZ = 60
  const skW = 6, skD = 6
  const wt = 0.5

  const wallMat = Materials.wood

  // Walls
  addWall(scene, skX, 0, skZ, skX + skW, WALL_H, skZ + wt, wallMat) // South
  addWall(scene, skX, 0, skZ + skD - wt, skX + 3, WALL_H, skZ + skD, wallMat) // North left
  addWall(scene, skX + 4.5, 0, skZ + skD - wt, skX + skW, WALL_H, skZ + skD, wallMat) // North right
  addWall(scene, skX, 0, skZ, skX + wt, WALL_H, skZ + skD, wallMat) // West
  addWall(scene, skX + skW - wt, 0, skZ, skX + skW, WALL_H, skZ + 2, wallMat) // East bottom
  addWall(scene, skX + skW - wt, 0, skZ + 3.5, skX + skW, WALL_H, skZ + skD, wallMat) // East top

  // Floor
  const floor = new THREE.Mesh(new THREE.BoxGeometry(skW, 0.1, skD), Materials.woodDark)
  floor.position.set(skX + skW / 2, 0.05, skZ + skD / 2)
  floor.receiveShadow = true
  scene.add(floor)

  // Roof
  const roof = new THREE.Mesh(new THREE.BoxGeometry(skW + 0.6, 0.2, skD + 0.6), Materials.woodDark)
  roof.position.set(skX + skW / 2, WALL_H + 0.1, skZ + skD / 2)
  roof.castShadow = true
  scene.add(roof)

  // Interior light
  const light = new THREE.PointLight(0xaa6622, 0.5, 10)
  light.position.set(skX + skW / 2, 2.5, skZ + skD / 2)
  scene.add(light)
}

function buildHill(scene: THREE.Scene) {
  const hX = 35, hZ = 10
  const hW = 10, hD = 8, hH = 2

  // Main hill body
  const hill = new THREE.Mesh(new THREE.BoxGeometry(hW, hH, hD), Materials.grass)
  hill.position.set(hX + hW / 2, hH / 2, hZ + hD / 2)
  hill.castShadow = true
  hill.receiveShadow = true
  scene.add(hill)

  // Slope ramps (visual) on two sides
  const slopeGeo = new THREE.BoxGeometry(3, 0.3, hD)

  const slopeL = new THREE.Mesh(slopeGeo, Materials.grassDark)
  slopeL.position.set(hX - 1, hH * 0.3, hZ + hD / 2)
  slopeL.rotation.z = Math.PI / 6
  scene.add(slopeL)

  const slopeR = new THREE.Mesh(slopeGeo, Materials.grassDark)
  slopeR.position.set(hX + hW + 1, hH * 0.3, hZ + hD / 2)
  slopeR.rotation.z = -Math.PI / 6
  scene.add(slopeR)

  // Some rocks on top
  for (let i = 0; i < 3; i++) {
    const rock = new THREE.Mesh(
      new THREE.BoxGeometry(0.5 + Math.random() * 0.5, 0.3 + Math.random() * 0.3, 0.5 + Math.random() * 0.5),
      Materials.concreteDark
    )
    rock.position.set(hX + 2 + i * 3, hH + 0.2, hZ + hD / 2 + (Math.random() - 0.5) * 4)
    rock.rotation.y = Math.random() * Math.PI
    rock.castShadow = true
    scene.add(rock)
  }
}

function buildLoopWalls(scene: THREE.Scene) {
  const wallMat = Materials.concrete

  // Loop 1 - L-shaped wall (southwest)
  addWall(scene, 12, 0, 15, 12.5, WALL_H, 22, wallMat)
  addWall(scene, 12, 0, 15, 17, WALL_H, 15.5, wallMat)

  // Loop 2 - T-wall (south-center-east)
  addWall(scene, 55, 0, 12, 55.5, WALL_H, 19, wallMat)
  addWall(scene, 53, 0, 15, 58, WALL_H, 15.5, wallMat)

  // Loop 3 - Straight wall (northwest)
  addWall(scene, 20, 0, 55, 20.5, WALL_H, 62, wallMat)

  // Loop 4 - L-shape (northeast)
  addWall(scene, 60, 0, 58, 60.5, WALL_H, 65, wallMat)
  addWall(scene, 60, 0, 58, 66, WALL_H, 58.5, wallMat)

  // Loop 5 - Corner walls (center-west)
  addWall(scene, 18, 0, 35, 18.5, WALL_H, 42, wallMat)
  addWall(scene, 18, 0, 38, 24, WALL_H, 38.5, wallMat)

  // Loop 6 - Car/debris (center) - shorter
  const carMat = Materials.rust
  addWall(scene, 32, 0, 38, 38, WALL_H * 0.6, 41, carMat)
  // Car details
  const wheels = [
    [32.5, 0.3, 38.5], [32.5, 0.3, 40.5],
    [37.5, 0.3, 38.5], [37.5, 0.3, 40.5]
  ]
  for (const [wx, wy, wz] of wheels) {
    const wheel = new THREE.Mesh(new THREE.BoxGeometry(0.2, 0.5, 0.5), Materials.black)
    wheel.position.set(wx, wy, wz)
    scene.add(wheel)
  }

  // Loop 7 - Wall near main building
  addWall(scene, 40, 0, 25, 40.5, WALL_H, 30, wallMat)

  // Loop 8 - Wall near shack
  addWall(scene, 16, 0, 55, 16.5, WALL_H, 59, wallMat)

  // Extra debris
  addWall(scene, 70, 0, 20, 72, WALL_H * 0.5, 22, Materials.rust)
  addWall(scene, 5, 0, 40, 7, WALL_H * 0.5, 42, Materials.rust)
  addWall(scene, 50, 0, 55, 52, WALL_H * 0.5, 57, Materials.rust)
}

function buildEnvironment(scene: THREE.Scene) {
  // Scatter dead trees
  const treePositions = [
    [5, 20], [15, 45], [25, 70], [35, 55],
    [50, 10], [60, 50], [70, 65], [75, 35],
    [10, 35], [30, 5], [55, 70], [68, 15],
    [42, 45], [22, 25], [65, 38]
  ]

  for (const [tx, tz] of treePositions) {
    buildDeadTree(scene, tx, tz)
  }

  // Scatter rocks
  const rockPositions = [
    [8, 12], [28, 42], [48, 58], [62, 22],
    [18, 68], [72, 55], [38, 28], [52, 42],
    [15, 8], [68, 72], [3, 58], [76, 12]
  ]

  for (const [rx, rz] of rockPositions) {
    const rock = new THREE.Mesh(
      new THREE.BoxGeometry(
        0.6 + Math.random() * 1.0,
        0.4 + Math.random() * 0.6,
        0.6 + Math.random() * 1.0
      ),
      Materials.concreteDark
    )
    rock.position.set(rx, (0.4 + Math.random() * 0.6) / 2, rz)
    rock.rotation.y = Math.random() * Math.PI
    rock.castShadow = true
    rock.receiveShadow = true
    scene.add(rock)
  }

  // Tall grass patches (visual only)
  for (let i = 0; i < 80; i++) {
    const gx = Math.random() * MAP_W
    const gz = Math.random() * MAP_H

    // Skip if inside buildings
    if (gx > 44 && gx < 66 && gz > 29 && gz < 49) continue
    if (gx > 7 && gx < 15 && gz > 59 && gz < 67) continue

    const grassPatch = new THREE.Mesh(
      new THREE.BoxGeometry(
        1 + Math.random() * 2,
        0.3 + Math.random() * 0.4,
        1 + Math.random() * 2
      ),
      Math.random() > 0.5 ? Materials.grass : Materials.grassDark
    )
    grassPatch.position.set(gx, (0.3 + Math.random() * 0.4) / 2, gz)
    grassPatch.receiveShadow = true
    scene.add(grassPatch)
  }

  // Add barrel props
  const barrelPositions = [
    [13, 20], [57, 14], [62, 60], [22, 60],
    [47, 32], [35, 50]
  ]

  for (const [bx, bz] of barrelPositions) {
    const barrel = new THREE.Mesh(new THREE.BoxGeometry(0.6, 1.0, 0.6), Materials.rust)
    barrel.position.set(bx, 0.5, bz)
    barrel.castShadow = true
    scene.add(barrel)

    // Barrel top
    const top = new THREE.Mesh(new THREE.BoxGeometry(0.55, 0.05, 0.55), Materials.metalDark)
    top.position.set(bx, 1.0, bz)
    scene.add(top)
  }
}

function buildDeadTree(scene: THREE.Scene, x: number, z: number) {
  const height = 3 + Math.random() * 3
  const trunk = new THREE.Mesh(
    new THREE.BoxGeometry(0.3, height, 0.3),
    Materials.woodDark
  )
  trunk.position.set(x, height / 2, z)
  trunk.castShadow = true
  scene.add(trunk)

  // A few dead branches
  const branchCount = 2 + Math.floor(Math.random() * 3)
  for (let i = 0; i < branchCount; i++) {
    const bLen = 0.8 + Math.random() * 1.5
    const bAngle = Math.random() * Math.PI * 2
    const bHeight = height * 0.4 + Math.random() * height * 0.5

    const branch = new THREE.Mesh(
      new THREE.BoxGeometry(bLen, 0.12, 0.12),
      Materials.woodDark
    )
    branch.position.set(
      x + Math.cos(bAngle) * bLen / 2,
      bHeight,
      z + Math.sin(bAngle) * bLen / 2
    )
    branch.rotation.y = bAngle
    branch.rotation.z = (Math.random() - 0.5) * 0.5
    branch.castShadow = true
    scene.add(branch)
  }
}

function addWall(scene: THREE.Scene, x1: number, y1: number, z1: number, x2: number, y2: number, z2: number, mat: THREE.Material) {
  const w = x2 - x1
  const h = y2 - y1
  const d = z2 - z1
  if (w <= 0 || h <= 0 || d <= 0) return

  const mesh = new THREE.Mesh(new THREE.BoxGeometry(w, h, d), mat)
  mesh.position.set(x1 + w / 2, y1 + h / 2, z1 + d / 2)
  mesh.castShadow = true
  mesh.receiveShadow = true
  scene.add(mesh)
}
