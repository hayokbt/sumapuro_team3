const STORAGE_KEY = 'settings'

// public/audio/ に置いたファイル名に合わせる
const BGM_SRC = '/audio/bgm.mp3'
const SE_SRC = '/audio/se_click.mp3'

export type Volumes = {
  bgmVolume: number
  seVolume: number
}

function loadVolumes(): Volumes {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
    return {
      bgmVolume: typeof saved.bgmVolume === 'number' ? saved.bgmVolume : 30,
      seVolume: typeof saved.seVolume === 'number' ? saved.seVolume : 60,
    }
  } catch {
    return { bgmVolume: 30, seVolume: 60 }
  }
}

// 画面表示中の音量(スライダーを動かすとここが変わる。保存するまでは次の読み込みで元に戻る)
const volumes: Volumes = loadVolumes()

let bgm: HTMLAudioElement | null = null

function getBgm(): HTMLAudioElement {
  if (!bgm) {
    bgm = new Audio(BGM_SRC)
    bgm.loop = true
    bgm.volume = volumes.bgmVolume / 100
  }
  return bgm
}

export function getVolumes(): Volumes {
  return { ...volumes }
}

export function saveVolumes() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(volumes))
}

// BGM を流す。ブラウザが自動再生をブロックしたら、最初のクリックかキー入力で再生する
export function startBgm() {
  const audio = getBgm()
  audio.play().catch(() => {
    const retry = () => {
      audio.play().catch(() => {})
    }
    window.addEventListener('pointerdown', retry, { once: true })
    window.addEventListener('keydown', retry, { once: true })
  })
}

export function applyBgmVolume(value: number) {
  volumes.bgmVolume = value
  if (bgm) bgm.volume = value / 100
}

export function applySeVolume(value: number) {
  volumes.seVolume = value
}

// 効果音を1回鳴らす
export function playSe() {
  const se = new Audio(SE_SRC)
  se.volume = volumes.seVolume / 100
  se.play().catch(() => {})
}