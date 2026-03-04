import * as THREE from 'three'

// Procedural character animation using sin-wave limb rotation
export class CharacterAnimator {
  private model: THREE.Group
  private time: number = 0

  // Cached part references
  private parts: Map<string, THREE.Object3D> = new Map()

  // Original positions for resetting
  private originalPositions: Map<string, THREE.Vector3> = new Map()

  constructor(model: THREE.Group) {
    this.model = model
    this.cacheChildren()
  }

  private cacheChildren() {
    this.model.traverse((child) => {
      if (child.name) {
        this.parts.set(child.name, child)
        this.originalPositions.set(child.name, child.position.clone())
      }
    })
  }

  private getPart(name: string): THREE.Object3D | undefined {
    return this.parts.get(name)
  }

  private getOrigPos(name: string): THREE.Vector3 {
    return this.originalPositions.get(name) || new THREE.Vector3()
  }

  update(dt: number, moveState: string, actionState: string, isKiller: boolean, isInjured: boolean = false) {
    this.time += dt

    switch (actionState) {
      case 'attacking':
        this.animateAttack()
        return
      case 'repairing':
        this.animateRepair()
        return
      case 'healing':
      case 'unhooking':
        this.animateHeal()
        return
      case 'hooked':
        this.animateHooked()
        return
      case 'being_carried':
        this.animateBeingCarried()
        return
      case 'dying':
        this.animateDowned()
        return
      case 'trapped':
        this.animateTrapped()
        return
      case 'stunned':
        this.animateStunned()
        return
      case 'carrying':
        this.animateCarrying(moveState)
        return
      case 'breaking_pallet':
        this.animateBreakPallet()
        return
      case 'placing_trap':
      case 'picking_up_trap':
        this.animatePlaceTrap()
        return
    }

    switch (moveState) {
      case 'running':
        if (isInjured) this.animateInjuredRun()
        else this.animateRun(isKiller ? 5.0 : 6.0)
        break
      case 'walking':
        if (isInjured) this.animateInjuredWalk()
        else this.animateWalk()
        break
      case 'crouching':
        this.animateCrouch()
        break
      default:
        this.animateIdle()
    }
  }

  private animateIdle() {
    const breathe = Math.sin(this.time * 2) * 0.005

    this.resetLimbs()

    const torso = this.getPart('torso')
    if (torso) {
      const orig = this.getOrigPos('torso')
      torso.position.y = orig.y + breathe
    }
  }

  private animateWalk() {
    const speed = 3.0
    const swing = 0.3
    const t = this.time * speed

    this.animateLegSwing(t, swing)
    this.animateArmSwing(t, swing * 0.6)
    this.animateBodyBob(t, 0.02)
  }

  private animateInjuredWalk() {
    const speed = 2.2
    const swing = 0.25
    const t = this.time * speed
    const limp = Math.sin(t * 2) * 0.04

    this.animateLegSwing(t, swing)
    this.animateArmSwing(t, swing * 0.5)
    this.animateBodyBob(t, 0.03)
    const torso = this.getPart('torso')
    if (torso) {
      const orig = this.getOrigPos('torso')
      torso.position.x = orig.x + limp
      torso.position.y = orig.y - 0.05
    }
  }

  private animateInjuredRun() {
    const speed = 4.0
    const swing = 0.35
    const t = this.time * speed
    const limp = Math.sin(t * 2) * 0.06

    this.animateLegSwing(t, swing)
    this.animateArmSwing(t, swing * 0.6)
    this.animateBodyBob(t, 0.05)
    const torso = this.getPart('torso')
    if (torso) {
      const orig = this.getOrigPos('torso')
      torso.position.x = orig.x + limp
      torso.position.y = orig.y - 0.08
    }
  }

  private animateRun(speed: number) {
    const swing = 0.5
    const t = this.time * speed

    this.animateLegSwing(t, swing)
    this.animateArmSwing(t, swing * 0.8)
    this.animateBodyBob(t, 0.04)
  }

  private animateCrouch() {
    const speed = 1.8
    const swing = 0.08
    const t = this.time * speed

    // Lower the body more for a compact crouch
    this.model.traverse((child) => {
      if (child !== this.model && child.name) {
        const orig = this.getOrigPos(child.name)
        if (orig) {
          child.position.y = orig.y - 0.48
        }
      }
    })

    // Arms slightly forward, legs tucked
    const armLU = this.getPart('armLU')
    const armRU = this.getPart('armRU')
    if (armLU) {
      const o = this.getOrigPos('armLU')
      armLU.position.set(o.x - 0.05, o.y - 0.1, o.z + 0.12)
    }
    if (armRU) {
      const o = this.getOrigPos('armRU')
      armRU.position.set(o.x + 0.05, o.y - 0.1, o.z + 0.12)
    }
    this.animateLegSwing(t, swing)
  }

  private animateAttack() {
    const t = this.time * 8
    const attackPhase = (Math.sin(t) + 1) / 2

    const armRU = this.getPart('armRU')
    const armRL = this.getPart('armRL')

    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.y = orig.y + attackPhase * 0.3
      armRU.position.z = orig.z + attackPhase * 0.3
    }
    if (armRL) {
      const orig = this.getOrigPos('armRL')
      armRL.position.y = orig.y + attackPhase * 0.2
      armRL.position.z = orig.z + attackPhase * 0.4
    }

    const weaponBlade = this.getPart('weaponBlade')
    const weaponHandle = this.getPart('weaponHandle')
    if (weaponBlade) {
      const orig = this.getOrigPos('weaponBlade')
      weaponBlade.position.y = orig.y + attackPhase * 0.2
      weaponBlade.position.z = orig.z + attackPhase * 0.5
    }
    if (weaponHandle) {
      const orig = this.getOrigPos('weaponHandle')
      weaponHandle.position.y = orig.y + attackPhase * 0.2
      weaponHandle.position.z = orig.z + attackPhase * 0.4
    }
  }

  private animateRepair() {
    const t = this.time * 4
    const handMove = Math.sin(t) * 0.1

    this.resetLimbs()

    // Arms reach forward
    const armLU = this.getPart('armLU')
    const armLL = this.getPart('armLL')
    const armRU = this.getPart('armRU')
    const armRL = this.getPart('armRL')

    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.z = orig.z + 0.15 + handMove
      armLU.position.y = orig.y + 0.05
    }
    if (armLL) {
      const orig = this.getOrigPos('armLL')
      armLL.position.z = orig.z + 0.2 + handMove
      armLL.position.y = orig.y + 0.1
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.z = orig.z + 0.15 - handMove
      armRU.position.y = orig.y + 0.05
    }
    if (armRL) {
      const orig = this.getOrigPos('armRL')
      armRL.position.z = orig.z + 0.2 - handMove
      armRL.position.y = orig.y + 0.1
    }
  }

  private animateHeal() {
    const t = this.time * 3
    const handMove = Math.sin(t) * 0.08

    this.resetLimbs()

    const armLU = this.getPart('armLU')
    const armRU = this.getPart('armRU')

    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.z = orig.z + 0.2
      armLU.position.y = orig.y + handMove
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.z = orig.z + 0.2
      armRU.position.y = orig.y - handMove
    }
  }

  private animateHooked() {
    const t = this.time * 2
    const struggle = Math.sin(t) * 0.03

    this.resetLimbs()

    // Arms up
    const armLU = this.getPart('armLU')
    const armRU = this.getPart('armRU')
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.y = orig.y + 0.5 + struggle
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.y = orig.y + 0.5 - struggle
    }

    // Legs dangle
    const legLU = this.getPart('legLU')
    const legRU = this.getPart('legRU')
    if (legLU) {
      const orig = this.getOrigPos('legLU')
      legLU.position.z = orig.z + struggle * 2
    }
    if (legRU) {
      const orig = this.getOrigPos('legRU')
      legRU.position.z = orig.z - struggle * 2
    }
  }

  private animateBeingCarried() {
    // Limp body draped over shoulder - rotate model
    this.model.rotation.x = Math.PI / 6
    this.model.rotation.z = Math.sin(this.time * 2) * 0.05
  }

  private animateDowned() {
    // Crawl / downed pose: body low, arms forward, legs bent, subtle struggle motion
    const t = this.time * 2
    const struggle = Math.sin(t) * 0.02

    this.resetLimbs()

    // Tilt model forward into crawl (on hands and knees)
    this.model.rotation.x = Math.PI / 2.2
    this.model.rotation.z = struggle

    const torso = this.getPart('torso')
    const head = this.getPart('head')
    const armLU = this.getPart('armLU')
    const armLL = this.getPart('armLL')
    const armRU = this.getPart('armRU')
    const armRL = this.getPart('armRL')
    const legLU = this.getPart('legLU')
    const legLL = this.getPart('legLL')
    const legRU = this.getPart('legRU')
    const legRL = this.getPart('legRL')

    // Lower torso and head slightly (in local space after tilt)
    if (torso) {
      const orig = this.getOrigPos('torso')
      torso.position.y = orig.y - 0.15 + struggle
    }
    if (head) {
      const orig = this.getOrigPos('head')
      head.position.y = orig.y - 0.1
    }

    // Arms forward (crawling)
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.z = orig.z + 0.25
      armLU.position.y = orig.y - 0.2
    }
    if (armLL) {
      const orig = this.getOrigPos('armLL')
      armLL.position.z = orig.z + 0.3
      armLL.position.y = orig.y - 0.25
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.z = orig.z + 0.25 - struggle
      armRU.position.y = orig.y - 0.2
    }
    if (armRL) {
      const orig = this.getOrigPos('armRL')
      armRL.position.z = orig.z + 0.3 + struggle
      armRL.position.y = orig.y - 0.25
    }

    // Legs bent (knees up)
    if (legLU) {
      const orig = this.getOrigPos('legLU')
      legLU.position.y = orig.y - 0.1
      legLU.position.z = orig.z - 0.15
    }
    if (legLL) {
      const orig = this.getOrigPos('legLL')
      legLL.position.z = orig.z - 0.2
    }
    if (legRU) {
      const orig = this.getOrigPos('legRU')
      legRU.position.y = orig.y - 0.1
      legRU.position.z = orig.z - 0.15
    }
    if (legRL) {
      const orig = this.getOrigPos('legRL')
      legRL.position.z = orig.z - 0.2
    }
  }

  private animateTrapped() {
    const t = this.time * 6
    const struggle = Math.sin(t) * 0.05

    this.resetLimbs()

    const armLU = this.getPart('armLU')
    const armRU = this.getPart('armRU')
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.y = orig.y + struggle
      armLU.position.x = orig.x + struggle
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.y = orig.y - struggle
      armRU.position.x = orig.x - struggle
    }
  }

  private animateStunned() {
    const wobble = Math.sin(this.time * 8) * 0.05
    this.model.rotation.z = wobble

    const head = this.getPart('head')
    if (head) {
      const orig = this.getOrigPos('head')
      head.position.x = orig.x + wobble * 2
    }
  }

  private animateCarrying(moveState: string) {
    const speed = moveState === 'running' ? 4.5 : 3.0
    const swing = moveState === 'running' ? 0.35 : 0.2
    const t = this.time * speed

    this.animateLegSwing(t, swing)

    // Left arm holds survivor (up)
    const armLU = this.getPart('armLU')
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.y = orig.y + 0.3
    }

    // Right arm swing reduced
    this.animateArmSwing(t, swing * 0.3)
  }

  private animateBreakPallet() {
    const t = this.time * 6
    const smash = Math.abs(Math.sin(t)) * 0.4

    this.resetLimbs()

    const armRU = this.getPart('armRU')
    const armLU = this.getPart('armLU')
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.y = orig.y + smash
      armRU.position.z = orig.z + smash * 0.5
    }
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.y = orig.y + smash
      armLU.position.z = orig.z + smash * 0.5
    }
  }

  private animatePlaceTrap() {
    const t = this.time * 3
    const handMove = Math.sin(t) * 0.05

    this.resetLimbs()

    // Crouch down
    this.model.traverse((child) => {
      if (child !== this.model && child.name) {
        const orig = this.getOrigPos(child.name)
        if (orig) child.position.y = orig.y - 0.4
      }
    })

    // Arms reaching down
    const armLU = this.getPart('armLU')
    const armRU = this.getPart('armRU')
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.y = orig.y - 0.5 + handMove
      armLU.position.z = orig.z + 0.2
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.y = orig.y - 0.5 - handMove
      armRU.position.z = orig.z + 0.2
    }
  }

  // Helpers
  private animateLegSwing(t: number, amplitude: number) {
    const legLU = this.getPart('legLU')
    const legLL = this.getPart('legLL')
    const legRU = this.getPart('legRU')
    const legRL = this.getPart('legRL')
    const footL = this.getPart('footL')
    const footR = this.getPart('footR')

    const swing = Math.sin(t) * amplitude

    if (legLU) {
      const orig = this.getOrigPos('legLU')
      legLU.position.z = orig.z + swing * 0.4
      legLU.position.y = orig.y + Math.abs(swing) * 0.05
    }
    if (legLL) {
      const orig = this.getOrigPos('legLL')
      legLL.position.z = orig.z + swing * 0.5
    }
    if (footL) {
      const orig = this.getOrigPos('footL')
      footL.position.z = orig.z + swing * 0.5
      footL.position.y = orig.y + Math.max(0, swing) * 0.1
    }

    if (legRU) {
      const orig = this.getOrigPos('legRU')
      legRU.position.z = orig.z - swing * 0.4
      legRU.position.y = orig.y + Math.abs(swing) * 0.05
    }
    if (legRL) {
      const orig = this.getOrigPos('legRL')
      legRL.position.z = orig.z - swing * 0.5
    }
    if (footR) {
      const orig = this.getOrigPos('footR')
      footR.position.z = orig.z - swing * 0.5
      footR.position.y = orig.y + Math.max(0, -swing) * 0.1
    }
  }

  private animateArmSwing(t: number, amplitude: number) {
    const armLU = this.getPart('armLU')
    const armLL = this.getPart('armLL')
    const armRU = this.getPart('armRU')
    const armRL = this.getPart('armRL')

    const swing = Math.sin(t) * amplitude

    // Arms swing opposite to legs
    if (armLU) {
      const orig = this.getOrigPos('armLU')
      armLU.position.z = orig.z - swing * 0.5
    }
    if (armLL) {
      const orig = this.getOrigPos('armLL')
      armLL.position.z = orig.z - swing * 0.6
    }
    if (armRU) {
      const orig = this.getOrigPos('armRU')
      armRU.position.z = orig.z + swing * 0.5
    }
    if (armRL) {
      const orig = this.getOrigPos('armRL')
      armRL.position.z = orig.z + swing * 0.6
    }
  }

  private animateBodyBob(t: number, amplitude: number) {
    const bob = Math.abs(Math.sin(t * 2)) * amplitude
    const torso = this.getPart('torso')
    const head = this.getPart('head')

    if (torso) {
      const orig = this.getOrigPos('torso')
      torso.position.y = orig.y + bob
    }
    if (head) {
      const orig = this.getOrigPos('head')
      head.position.y = orig.y + bob
    }
  }

  private resetLimbs() {
    this.model.rotation.x = 0
    this.model.rotation.z = 0

    this.parts.forEach((part, name) => {
      const orig = this.originalPositions.get(name)
      if (orig) {
        part.position.copy(orig)
      }
    })
  }

  reset() {
    this.time = 0
    this.resetLimbs()
  }
}
