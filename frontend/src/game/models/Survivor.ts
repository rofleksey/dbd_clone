import * as THREE from 'three'
import { Materials } from './Materials'

export function createSurvivorModel(): THREE.Group {
  const group = new THREE.Group()

  // Scale: 1 unit = 1 meter, survivor is ~1.7m tall

  // Head
  const head = new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.28, 0.28), Materials.skin)
  head.position.set(0, 1.54, 0)
  head.name = 'head'
  group.add(head)

  // Hair (on top and back of head)
  const hairTop = new THREE.Mesh(new THREE.BoxGeometry(0.30, 0.08, 0.30), Materials.hair)
  hairTop.position.set(0, 1.72, 0)
  group.add(hairTop)

  const hairBack = new THREE.Mesh(new THREE.BoxGeometry(0.30, 0.20, 0.06), Materials.hair)
  hairBack.position.set(0, 1.58, -0.16)
  group.add(hairBack)

  // Eyes (small black cubes)
  const eyeGeo = new THREE.BoxGeometry(0.05, 0.04, 0.02)
  const eyeL = new THREE.Mesh(eyeGeo, Materials.black)
  eyeL.position.set(-0.07, 1.56, 0.14)
  group.add(eyeL)
  const eyeR = new THREE.Mesh(eyeGeo, Materials.black)
  eyeR.position.set(0.07, 1.56, 0.14)
  group.add(eyeR)

  // Torso (upper body)
  const torso = new THREE.Mesh(new THREE.BoxGeometry(0.38, 0.36, 0.22), Materials.survivorShirt)
  torso.position.set(0, 1.22, 0)
  torso.name = 'torso'
  group.add(torso)

  // Lower torso / hips
  const hips = new THREE.Mesh(new THREE.BoxGeometry(0.34, 0.14, 0.20), Materials.survivorPants)
  hips.position.set(0, 0.97, 0)
  group.add(hips)

  // Left arm upper
  const armLU = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.26, 0.12), Materials.survivorShirt)
  armLU.position.set(-0.25, 1.27, 0)
  armLU.name = 'armLU'
  group.add(armLU)

  // Left arm lower
  const armLL = new THREE.Mesh(new THREE.BoxGeometry(0.10, 0.24, 0.10), Materials.skin)
  armLL.position.set(-0.25, 1.01, 0)
  armLL.name = 'armLL'
  group.add(armLL)

  // Left hand
  const handL = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.08, 0.08), Materials.skin)
  handL.position.set(-0.25, 0.87, 0)
  handL.name = 'handL'
  group.add(handL)

  // Right arm upper
  const armRU = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.26, 0.12), Materials.survivorShirt)
  armRU.position.set(0.25, 1.27, 0)
  armRU.name = 'armRU'
  group.add(armRU)

  // Right arm lower
  const armRL = new THREE.Mesh(new THREE.BoxGeometry(0.10, 0.24, 0.10), Materials.skin)
  armRL.position.set(0.25, 1.01, 0)
  armRL.name = 'armRL'
  group.add(armRL)

  // Right hand
  const handR = new THREE.Mesh(new THREE.BoxGeometry(0.08, 0.08, 0.08), Materials.skin)
  handR.position.set(0.25, 0.87, 0)
  handR.name = 'handR'
  group.add(handR)

  // Left leg upper
  const legLU = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.30, 0.14), Materials.survivorPants)
  legLU.position.set(-0.09, 0.75, 0)
  legLU.name = 'legLU'
  group.add(legLU)

  // Left leg lower
  const legLL = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.30, 0.12), Materials.survivorPants)
  legLL.position.set(-0.09, 0.45, 0)
  legLL.name = 'legLL'
  group.add(legLL)

  // Left foot
  const footL = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.08, 0.18), Materials.survivorShoes)
  footL.position.set(-0.09, 0.28, 0.03)
  footL.name = 'footL'
  group.add(footL)

  // Right leg upper
  const legRU = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.30, 0.14), Materials.survivorPants)
  legRU.position.set(0.09, 0.75, 0)
  legRU.name = 'legRU'
  group.add(legRU)

  // Right leg lower
  const legRL = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.30, 0.12), Materials.survivorPants)
  legRL.position.set(0.09, 0.45, 0)
  legRL.name = 'legRL'
  group.add(legRL)

  // Right foot
  const footR = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.08, 0.18), Materials.survivorShoes)
  footR.position.set(0.09, 0.28, 0.03)
  footR.name = 'footR'
  group.add(footR)

  // Set shadow casting
  group.traverse((child) => {
    if (child instanceof THREE.Mesh) {
      child.castShadow = true
      child.receiveShadow = true
    }
  })

  return group
}
