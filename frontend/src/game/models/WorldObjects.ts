import * as THREE from 'three'
import { Materials } from './Materials'

export function createGeneratorModel(): THREE.Group {
  const group = new THREE.Group()

  // Base frame
  const base = new THREE.Mesh(new THREE.BoxGeometry(1.2, 0.2, 0.8), Materials.metalDark)
  base.position.set(0, 0.1, 0)
  group.add(base)

  // Engine block
  const engine = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.6, 0.6), Materials.metal)
  engine.position.set(0, 0.5, 0)
  group.add(engine)

  // Pistons (top)
  for (let i = -1; i <= 1; i += 2) {
    const piston = new THREE.Mesh(new THREE.BoxGeometry(0.15, 0.3, 0.15), Materials.metalDark)
    piston.position.set(i * 0.2, 0.95, 0)
    piston.name = 'piston'
    group.add(piston)
  }

  // Exhaust pipe
  const exhaust = new THREE.Mesh(new THREE.BoxGeometry(0.1, 0.5, 0.1), Materials.rust)
  exhaust.position.set(0.5, 1.0, 0.3)
  group.add(exhaust)

  // Light indicator
  const light = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.12, 0.04), Materials.genRed)
  light.position.set(0, 0.85, 0.32)
  light.name = 'indicator'
  group.add(light)

  // Side panels
  const panelL = new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.5, 0.7), Materials.rust)
  panelL.position.set(-0.42, 0.45, 0)
  group.add(panelL)

  const panelR = new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.5, 0.7), Materials.rust)
  panelR.position.set(0.42, 0.45, 0)
  group.add(panelR)

  group.traverse((c) => { if (c instanceof THREE.Mesh) { c.castShadow = true; c.receiveShadow = true } })
  return group
}

export function createPalletModel(): THREE.Group {
  const group = new THREE.Group()

  // DBD-style pallet: two vertical supports with horizontal planks between. ~2.2m wide, ~1m tall when standing.

  // Vertical posts (left and right)
  const postL = new THREE.Mesh(new THREE.BoxGeometry(0.12, 1.0, 0.12), Materials.woodDark)
  postL.position.set(-1.04, 0.5, 0)
  postL.name = 'postL'
  group.add(postL)

  const postR = new THREE.Mesh(new THREE.BoxGeometry(0.12, 1.0, 0.12), Materials.woodDark)
  postR.position.set(1.04, 0.5, 0)
  postR.name = 'postR'
  group.add(postR)

  // Horizontal planks (width 2.0 in X, thin in Z)
  const plankW = 2.08
  const plankH = 0.06
  const plankD = 0.1
  const heights = [0.15, 0.42, 0.68, 0.92]
  heights.forEach((y, i) => {
    const plank = new THREE.Mesh(
      new THREE.BoxGeometry(plankW, plankH, plankD),
      i % 2 === 0 ? Materials.wood : Materials.woodDark
    )
    plank.position.set(0, y, 0)
    group.add(plank)
  })

  group.traverse((c) => { if (c instanceof THREE.Mesh) { c.castShadow = true; c.receiveShadow = true } })
  return group
}

export function createHookModel(): THREE.Group {
  const group = new THREE.Group()

  // Wooden pole (smaller hook overall)
  const pole = new THREE.Mesh(new THREE.BoxGeometry(0.12, 2.4, 0.12), Materials.woodDark)
  pole.position.set(0, 1.2, 0)
  group.add(pole)

  // Cross beam
  const beam = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.08, 0.08), Materials.woodDark)
  beam.position.set(0.2, 2.2, 0)
  group.add(beam)

  // Hook (metal)
  const hookBase = new THREE.Mesh(new THREE.BoxGeometry(0.05, 0.2, 0.05), Materials.metal)
  hookBase.position.set(0.4, 2.05, 0)
  group.add(hookBase)

  const hookCurve = new THREE.Mesh(new THREE.BoxGeometry(0.1, 0.05, 0.05), Materials.metal)
  hookCurve.position.set(0.36, 1.92, 0)
  group.add(hookCurve)

  const hookPoint = new THREE.Mesh(new THREE.BoxGeometry(0.05, 0.1, 0.05), Materials.metal)
  hookPoint.position.set(0.32, 1.86, 0)
  group.add(hookPoint)

  // Entity indicator (red glow)
  const indicator = new THREE.Mesh(new THREE.BoxGeometry(0.06, 0.06, 0.06), Materials.redLight)
  indicator.position.set(0, 2.4, 0)
  indicator.name = 'indicator'
  group.add(indicator)

  group.traverse((c) => { if (c instanceof THREE.Mesh) { c.castShadow = true; c.receiveShadow = true } })
  return group
}

export function createTrapModel(): THREE.Group {
  const group = new THREE.Group()

  // Bear trap - circular base with teeth
  const base = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.06, 0.5), Materials.trapMetal)
  base.position.set(0, 0.03, 0)
  group.add(base)

  // Jaw left
  const jawL = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.15, 0.06), Materials.trapMetal)
  jawL.position.set(0, 0.12, -0.2)
  jawL.name = 'jawL'
  group.add(jawL)

  // Jaw right
  const jawR = new THREE.Mesh(new THREE.BoxGeometry(0.5, 0.15, 0.06), Materials.trapMetal)
  jawR.position.set(0, 0.12, 0.2)
  jawR.name = 'jawR'
  group.add(jawR)

  // Teeth
  for (let i = 0; i < 5; i++) {
    const tooth = new THREE.Mesh(new THREE.BoxGeometry(0.03, 0.1, 0.03), Materials.trapTeeth)
    tooth.position.set(-0.2 + i * 0.1, 0.2, -0.18)
    tooth.name = 'tooth'
    group.add(tooth)

    const tooth2 = new THREE.Mesh(new THREE.BoxGeometry(0.03, 0.1, 0.03), Materials.trapTeeth)
    tooth2.position.set(-0.2 + i * 0.1, 0.2, 0.18)
    tooth2.name = 'tooth'
    group.add(tooth2)
  }

  // Spring mechanism
  const spring = new THREE.Mesh(new THREE.BoxGeometry(0.06, 0.08, 0.3), Materials.metalDark)
  spring.position.set(0, 0.1, 0)
  group.add(spring)

  group.traverse((c) => { if (c instanceof THREE.Mesh) { c.castShadow = true; c.receiveShadow = true } })
  return group
}

export function createExitGateModel(): THREE.Group {
  const group = new THREE.Group()

  // Gate frame - two tall posts and crossbeam
  const postL = new THREE.Mesh(new THREE.BoxGeometry(0.4, 5.0, 0.4), Materials.gateFrame)
  postL.position.set(-3, 2.5, 0)
  group.add(postL)

  const postR = new THREE.Mesh(new THREE.BoxGeometry(0.4, 5.0, 0.4), Materials.gateFrame)
  postR.position.set(3, 2.5, 0)
  group.add(postR)

  const crossbeam = new THREE.Mesh(new THREE.BoxGeometry(6.4, 0.3, 0.3), Materials.gateFrame)
  crossbeam.position.set(0, 4.85, 0)
  group.add(crossbeam)

  // Gate door (slides open)
  const door = new THREE.Mesh(new THREE.BoxGeometry(5.6, 4.5, 0.15), Materials.gateDoor)
  door.position.set(0, 2.25, 0)
  door.name = 'door'
  group.add(door)

  // Gate switch panel
  const switchPanel = new THREE.Mesh(new THREE.BoxGeometry(0.3, 0.5, 0.15), Materials.metalDark)
  switchPanel.position.set(-3.4, 1.5, 0.3)
  group.add(switchPanel)

  // Switch handle
  const handle = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.2, 0.08), Materials.metal)
  handle.position.set(-3.4, 1.6, 0.42)
  handle.name = 'handle'
  group.add(handle)

  // Lights (3 lights that fill up as gate opens)
  for (let i = 0; i < 3; i++) {
    const light = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.12, 0.06), Materials.genRed)
    light.position.set(-3.4, 1.9 + i * 0.2, 0.38)
    light.name = `light_${i}`
    group.add(light)
  }

  // Warning signs
  const sign = new THREE.Mesh(new THREE.BoxGeometry(0.6, 0.4, 0.04), Materials.yellow)
  sign.position.set(0, 4.5, 0.17)
  group.add(sign)

  group.traverse((c) => { if (c instanceof THREE.Mesh) { c.castShadow = true; c.receiveShadow = true } })
  return group
}

export function createWindowModel(): THREE.Group {
  const group = new THREE.Group()

  // Window frame - a wall segment with a vaultable opening
  const wallL = new THREE.Mesh(new THREE.BoxGeometry(0.5, 2.5, 0.3), Materials.wood)
  wallL.position.set(-0.75, 1.25, 0)
  group.add(wallL)

  const wallR = new THREE.Mesh(new THREE.BoxGeometry(0.5, 2.5, 0.3), Materials.wood)
  wallR.position.set(0.75, 1.25, 0)
  group.add(wallR)

  // Top of window
  const top = new THREE.Mesh(new THREE.BoxGeometry(2.0, 0.5, 0.3), Materials.wood)
  top.position.set(0, 2.25, 0)
  group.add(top)

  // Window sill
  const sill = new THREE.Mesh(new THREE.BoxGeometry(1.0, 0.1, 0.4), Materials.woodDark)
  sill.position.set(0, 0.9, 0)
  group.add(sill)

  group.traverse((c) => { if (c instanceof THREE.Mesh) { c.castShadow = true; c.receiveShadow = true } })
  return group
}
