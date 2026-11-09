const STORAGE_KEY = 'settings'
const BGM_POSITION_KEY = 'bgmPosition'
const BGM_SAVED_AT_KEY = 'bgmSavedAt'

// public/audio/ に置いたファイル名に合わせる
const BGM_SRC = '/audio/bgm.mp3'
const SE_SRC = '/audio/se_click.mp3'

// 効果音が小さいときの増幅倍率(1 なら元の音量。大きすぎると音が割れます)
const SE_BOOST = 3

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

// 再生中のBGMの位置と時刻を記録する(ページを移ったあと、続きから流すため)
function savePosition(audio: HTMLAudioElement) {
  if (audio.paused) return // 鳴っていないときは、前の記録を消さない
  try {
    sessionStorage.setItem(BGM_POSITION_KEY, String(audio.currentTime))
    sessionStorage.setItem(BGM_SAVED_AT_KEY, String(Date.now()))
  } catch {
    // 保存できなくても、BGM は止めない
  }
}

// 記録した位置に、ページを移っている間の経過時間を足して返す
function loadPosition(duration: number): number {
  try {
    const position = Number(sessionStorage.getItem(BGM_POSITION_KEY))
    const savedAt = Number(sessionStorage.getItem(BGM_SAVED_AT_KEY))
    if (position > 0 && savedAt > 0 && duration > 0) {
      const elapsed = (Date.now() - savedAt) / 1000
      return (position + elapsed) % duration
    }
  } catch {
    // 読めなければ頭から
  }
  return 0
}

function getBgm(): HTMLAudioElement {
  if (!bgm) {
    const audio = new Audio(BGM_SRC)
    audio.loop = true
    audio.volume = volumes.bgmVolume / 100

    // 読み込めたら、前のページの続きの位置まで進める
    audio.addEventListener('loadedmetadata', () => {
      audio.currentTime = loadPosition(audio.duration)
    })
    // 再生中は位置を記録し続ける
    audio.addEventListener('timeupdate', () => savePosition(audio))
    // ページを離れる直前にも記録する
    window.addEventListener('pagehide', () => savePosition(audio))

    bgm = audio
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

let audioContext: AudioContext | null = null
let seBuffer: AudioBuffer | null = null

// 効果音を1回鳴らす(Web Audio API で、SE_BOOST 倍まで増幅できる)
export async function playSe() {
  try {
    if (!audioContext) audioContext = new AudioContext()
    if (audioContext.state === 'suspended') await audioContext.resume()

    if (!seBuffer) {
      const response = await fetch(SE_SRC)
      seBuffer = await audioContext.decodeAudioData(await response.arrayBuffer())
    }

    const source = audioContext.createBufferSource()
    source.buffer = seBuffer

    const gain = audioContext.createGain()
    gain.gain.value = (volumes.seVolume / 100) * SE_BOOST

    source.connect(gain)
    gain.connect(audioContext.destination)
    source.start()
  } catch {
    // 再生できなくても、ゲームは止めない
  }
}