const STORAGE_KEY = 'settings'
const DEFAULT_VOLUME = 50
const SE_SRC = '/audio/se_click.mp3'

export type Volumes = {
  bgmVolume: number
  seVolume: number
}

function storedVolume(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 100
    ? value
    : DEFAULT_VOLUME
}

function loadVolumes(): Volumes {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
    return {
      bgmVolume: storedVolume(saved.bgmVolume),
      seVolume: storedVolume(saved.seVolume),
    }
  } catch {
    return { bgmVolume: DEFAULT_VOLUME, seVolume: DEFAULT_VOLUME }
  }
}

const volumes: Volumes = loadVolumes()
let se: HTMLAudioElement | null = null

function getSe(): HTMLAudioElement {
  if (!se) {
    se = new Audio(SE_SRC)
    se.preload = 'auto'
    se.volume = volumes.seVolume / 100
  }
  return se
}

export function getVolumes(): Volumes {
  return { ...volumes }
}

export function saveVolumes() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(volumes))
}

export function applyBgmVolume(value: number) {
  volumes.bgmVolume = storedVolume(value)
}

export function applySeVolume(value: number) {
  volumes.seVolume = storedVolume(value)
  if (se) se.volume = volumes.seVolume / 100
}

export function playSe() {
  if (volumes.seVolume === 0) return
  const audio = getSe()
  audio.currentTime = 0
  audio.volume = volumes.seVolume / 100
  audio.play().catch(() => {})
}

export function installButtonSounds(): () => void {
  getSe().load()
  const isButton = (target: EventTarget | null) => {
    if (!(target instanceof Element)) return false
    const button = target.closest('button, a[href], [role="button"], input[type="submit"], input[type="button"], input[type="reset"]')
    return !!button && !button.matches(':disabled') && !button.closest('[aria-disabled="true"]')
  }
  // Play on press so navigation does not cut off the start of the sound.
  const pointerDown = (event: PointerEvent) => {
    if (event.button === 0 && event.isPrimary && isButton(event.target)) playSe()
  }
  const click = (event: MouseEvent) => {
    // Keyboard and assistive-technology clicks have no pointerdown event.
    if (event.detail === 0 && isButton(event.target)) playSe()
  }
  window.addEventListener('pointerdown', pointerDown, true)
  window.addEventListener('click', click, true)
  return () => {
    window.removeEventListener('pointerdown', pointerDown, true)
    window.removeEventListener('click', click, true)
  }
}
