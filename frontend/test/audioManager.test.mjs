import assert from 'node:assert/strict'
import { test } from 'node:test'

test('BGM retries after autoplay rejection, loops, follows volume, and cleans up', async (t) => {
  const original = Object.getOwnPropertyDescriptors(globalThis)
  const storage = () => {
    const values = new Map()
    return { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value) }
  }
  const players = []
  let blocked = true
  class FakeAudio extends EventTarget {
    paused = true
    readyState = 1
    duration = 120
    currentTime = 0
    attempts = 0
    constructor(src) { super(); this.src = src; players.push(this) }
    load() {}
    play() {
      this.attempts++
      if (blocked) return Promise.reject(new Error('NotAllowedError'))
      this.paused = false
      return Promise.resolve()
    }
    pause() { this.paused = true }
  }
  const replacements = {
    window: new EventTarget(), Audio: FakeAudio, Element: class {},
    localStorage: storage(), sessionStorage: storage(),
  }
  for (const [key, value] of Object.entries(replacements)) {
    Object.defineProperty(globalThis, key, { configurable: true, writable: true, value })
  }
  t.after(() => {
    for (const key of Object.keys(replacements)) {
      if (original[key]) Object.defineProperty(globalThis, key, original[key])
      else delete globalThis[key]
    }
  })
  sessionStorage.setItem('bgmPosition', '42')
  const { installAudio, applyBgmVolume, applySeVolume, saveVolumes } = await import('../src/audio/audioManager.ts')
  const cleanup = installAudio()
  const bgm = players.find(player => player.src === '/audio/bgm.mp3')
  const se = players.find(player => player.src === '/audio/se_click.mp3')
  assert.equal(bgm.loop, true)
  assert.equal(bgm.volume, 0.5)
  await Promise.resolve() // A rejected autoplay attempt must be handled.
  assert.equal(bgm.paused, true)
  bgm.dispatchEvent(new Event('loadedmetadata'))
  assert.equal(bgm.currentTime, 42)

  blocked = false
  window.dispatchEvent(new Event('click'))
  assert.equal(bgm.paused, false)
  const attempts = bgm.attempts
  window.dispatchEvent(new Event('click'))
  assert.equal(bgm.attempts, attempts, 'clicks must not restart playing music')
  assert.equal(bgm.currentTime, 42)
  applyBgmVolume(25)
  assert.equal(bgm.volume, 0.25)
  applySeVolume(70)
  assert.equal(se.volume, 0.7)
  assert.equal(bgm.volume, 0.25, 'SE and BGM volumes are independent')
  applyBgmVolume(0)
  assert.equal(bgm.paused, true)
  window.dispatchEvent(new Event('keydown'))
  assert.equal(bgm.paused, true, 'gestures must not unmute music')
  applyBgmVolume(30)
  assert.equal(bgm.paused, false)
  assert.equal(bgm.volume, 0.3)
  saveVolumes()
  assert.equal(JSON.parse(localStorage.getItem('settings')).bgmVolume, 30)

  bgm.currentTime = 65
  window.dispatchEvent(new Event('pagehide'))
  assert.equal(sessionStorage.getItem('bgmPosition'), '65')
  assert.equal(bgm.paused, true)
  window.dispatchEvent(new Event('pageshow'))
  assert.equal(bgm.paused, false)
  cleanup()
  assert.equal(bgm.paused, true)
  const stoppedAttempts = bgm.attempts
  window.dispatchEvent(new Event('click'))
  window.dispatchEvent(new Event('keydown'))
  assert.equal(bgm.attempts, stoppedAttempts)
  const cleanupAgain = installAudio()
  assert.equal(players.filter(player => player.src === '/audio/bgm.mp3').length, 1)
  assert.equal(bgm.paused, false)
  cleanupAgain()
})
