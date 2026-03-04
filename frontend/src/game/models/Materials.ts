import * as THREE from 'three'

// Shared materials for voxel-style rendering
export const Materials = {
  // Survivor
  skin: new THREE.MeshLambertMaterial({ color: 0xc4956a }),
  survivorShirt: new THREE.MeshLambertMaterial({ color: 0x3d5c8a }),
  survivorPants: new THREE.MeshLambertMaterial({ color: 0x2a3a4a }),
  survivorShoes: new THREE.MeshLambertMaterial({ color: 0x3a2a1a }),
  hair: new THREE.MeshLambertMaterial({ color: 0x2a1a0a }),

  // Trapper / Killer
  killerSkin: new THREE.MeshLambertMaterial({ color: 0x8a7a6a }),
  killerOveralls: new THREE.MeshLambertMaterial({ color: 0x4a3a2a }),
  killerMask: new THREE.MeshLambertMaterial({ color: 0xd4c4a4 }),
  killerMaskDark: new THREE.MeshLambertMaterial({ color: 0x5a4a3a }),
  killerWeapon: new THREE.MeshLambertMaterial({ color: 0x6a6a6a }),
  killerApron: new THREE.MeshLambertMaterial({ color: 0x5a3a2a }),

  // World objects
  metal: new THREE.MeshLambertMaterial({ color: 0x6a6a6a }),
  metalDark: new THREE.MeshLambertMaterial({ color: 0x3a3a3a }),
  rust: new THREE.MeshLambertMaterial({ color: 0x8a4a2a }),
  wood: new THREE.MeshLambertMaterial({ color: 0x6a4a2a }),
  woodDark: new THREE.MeshLambertMaterial({ color: 0x4a3218 }),
  concrete: new THREE.MeshLambertMaterial({ color: 0x7a7a7a }),
  concreteDark: new THREE.MeshLambertMaterial({ color: 0x4a4a4a }),
  dirt: new THREE.MeshLambertMaterial({ color: 0x4a3a2a }),
  grass: new THREE.MeshLambertMaterial({ color: 0x2a4a1a }),
  grassDark: new THREE.MeshLambertMaterial({ color: 0x1a3a0a }),
  red: new THREE.MeshLambertMaterial({ color: 0xcc2222 }),
  redLight: new THREE.MeshLambertMaterial({ color: 0xff4444 }),
  yellow: new THREE.MeshLambertMaterial({ color: 0xccaa22 }),
  white: new THREE.MeshLambertMaterial({ color: 0xdddddd }),
  black: new THREE.MeshLambertMaterial({ color: 0x111111 }),
  blood: new THREE.MeshLambertMaterial({ color: 0x8a1111 }),
  rope: new THREE.MeshLambertMaterial({ color: 0x8a7a5a }),
  genGreen: new THREE.MeshLambertMaterial({ color: 0x22aa22, emissive: 0x116611, emissiveIntensity: 0.3 }),
  genRed: new THREE.MeshLambertMaterial({ color: 0xaa2222, emissive: 0x661111, emissiveIntensity: 0.3 }),
  trapMetal: new THREE.MeshLambertMaterial({ color: 0x5a5a5a }),
  trapTeeth: new THREE.MeshLambertMaterial({ color: 0x8a8a8a }),
  gateFrame: new THREE.MeshLambertMaterial({ color: 0x5a5a5a }),
  gateDoor: new THREE.MeshLambertMaterial({ color: 0x4a4a4a }),
  scratchMark: new THREE.MeshBasicMaterial({ color: 0xff4400, transparent: true, opacity: 0.6 }),
  bloodTrail: new THREE.MeshBasicMaterial({ color: 0x880000, transparent: true, opacity: 0.7 }),
}
