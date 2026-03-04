import * as THREE from 'three'
import { Materials } from './Materials'

export function createTrapperModel(): THREE.Group {
  const group = new THREE.Group()

  // The Trapper is larger and more imposing: ~2.0m tall, broader

  // Head
  const head = new THREE.Mesh(new THREE.BoxGeometry(0.30, 0.30, 0.30), Materials.killerSkin)
  head.position.set(0, 1.82, 0)
  head.name = 'head'
  group.add(head)

  // Mask - the iconic Trapper mask
  const maskFace = new THREE.Mesh(new THREE.BoxGeometry(0.32, 0.22, 0.06), Materials.killerMask)
  maskFace.position.set(0, 1.84, 0.16)
  group.add(maskFace)

  // Mask dark eye holes
  const maskEyeL = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.06, 0.02), Materials.black)
  maskEyeL.position.set(-0.07, 1.87, 0.19)
  group.add(maskEyeL)
  const maskEyeR = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.06, 0.02), Materials.black)
  maskEyeR.position.set(0.07, 1.87, 0.19)
  group.add(maskEyeR)

  // Mask mouth grate (horizontal lines)
  for (let i = 0; i < 3; i++) {
    const line = new THREE.Mesh(new THREE.BoxGeometry(0.20, 0.02, 0.02), Materials.killerMaskDark)
    line.position.set(0, 1.79 - i * 0.04, 0.19)
    group.add(line)
  }

  // Mask forehead ridge
  const maskRidge = new THREE.Mesh(new THREE.BoxGeometry(0.34, 0.04, 0.04), Materials.killerMaskDark)
  maskRidge.position.set(0, 1.94, 0.15)
  group.add(maskRidge)

  // Torso - broad, muscular
  const torso = new THREE.Mesh(new THREE.BoxGeometry(0.48, 0.42, 0.26), Materials.killerOveralls)
  torso.position.set(0, 1.44, 0)
  torso.name = 'torso'
  group.add(torso)

  // Apron (front of torso)
  const apron = new THREE.Mesh(new THREE.BoxGeometry(0.36, 0.48, 0.04), Materials.killerApron)
  apron.position.set(0, 1.36, 0.14)
  group.add(apron)

  // Blood stains on apron
  const bloodStain = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.12, 0.02), Materials.blood)
  bloodStain.position.set(0.06, 1.40, 0.16)
  group.add(bloodStain)
  const bloodStain2 = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.08, 0.02), Materials.blood)
  bloodStain2.position.set(-0.08, 1.28, 0.16)
  group.add(bloodStain2)

  // Lower torso / hips
  const hips = new THREE.Mesh(new THREE.BoxGeometry(0.42, 0.16, 0.24), Materials.killerOveralls)
  hips.position.set(0, 1.14, 0)
  group.add(hips)

  // Left arm upper (larger)
  const armLU = new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.30, 0.16), Materials.killerSkin)
  armLU.position.set(-0.32, 1.48, 0)
  armLU.name = 'armLU'
  group.add(armLU)

  // Left arm lower
  const armLL = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.28, 0.14), Materials.killerSkin)
  armLL.position.set(-0.32, 1.18, 0)
  armLL.name = 'armLL'
  group.add(armLL)

  // Left hand
  const handL = new THREE.Mesh(new THREE.BoxGeometry(0.10, 0.10, 0.10), Materials.killerSkin)
  handL.position.set(-0.32, 1.02, 0)
  handL.name = 'handL'
  group.add(handL)

  // Right arm upper (weapon arm)
  const armRU = new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.30, 0.16), Materials.killerSkin)
  armRU.position.set(0.32, 1.48, 0)
  armRU.name = 'armRU'
  group.add(armRU)

  // Right arm lower
  const armRL = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.28, 0.14), Materials.killerSkin)
  armRL.position.set(0.32, 1.18, 0)
  armRL.name = 'armRL'
  group.add(armRL)

  // Right hand
  const handR = new THREE.Mesh(new THREE.BoxGeometry(0.10, 0.10, 0.10), Materials.killerSkin)
  handR.position.set(0.32, 1.02, 0)
  handR.name = 'handR'
  group.add(handR)

  // Cleaver / Weapon in right hand
  const weaponHandle = new THREE.Mesh(new THREE.BoxGeometry(0.04, 0.20, 0.04), Materials.woodDark)
  weaponHandle.position.set(0.32, 0.88, 0.05)
  weaponHandle.name = 'weaponHandle'
  group.add(weaponHandle)

  const weaponBlade = new THREE.Mesh(new THREE.BoxGeometry(0.02, 0.28, 0.16), Materials.killerWeapon)
  weaponBlade.position.set(0.32, 0.76, 0.14)
  weaponBlade.name = 'weaponBlade'
  group.add(weaponBlade)

  // Left leg upper
  const legLU = new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.32, 0.16), Materials.killerOveralls)
  legLU.position.set(-0.12, 0.88, 0)
  legLU.name = 'legLU'
  group.add(legLU)

  // Left leg lower
  const legLL = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.34, 0.14), Materials.killerOveralls)
  legLL.position.set(-0.12, 0.54, 0)
  legLL.name = 'legLL'
  group.add(legLL)

  // Left boot
  const bootL = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.10, 0.20), Materials.killerMaskDark)
  bootL.position.set(-0.12, 0.33, 0.03)
  bootL.name = 'footL'
  group.add(bootL)

  // Right leg upper
  const legRU = new THREE.Mesh(new THREE.BoxGeometry(0.16, 0.32, 0.16), Materials.killerOveralls)
  legRU.position.set(0.12, 0.88, 0)
  legRU.name = 'legRU'
  group.add(legRU)

  // Right leg lower
  const legRL = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.34, 0.14), Materials.killerOveralls)
  legRL.position.set(0.12, 0.54, 0)
  legRL.name = 'legRL'
  group.add(legRL)

  // Right boot
  const bootR = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.10, 0.20), Materials.killerMaskDark)
  bootR.position.set(0.12, 0.33, 0.03)
  bootR.name = 'footR'
  group.add(bootR)

  // Set shadow casting
  group.traverse((child) => {
    if (child instanceof THREE.Mesh) {
      child.castShadow = true
      child.receiveShadow = true
    }
  })

  return group
}
