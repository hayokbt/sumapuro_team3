import assert from 'node:assert/strict'
import { test } from 'node:test'

test('internal navigation keeps BGM playing and supports browser history and native links', async (t) => {
  const descriptors = Object.getOwnPropertyDescriptors(globalThis)
  const location = new URL('http://localhost:5173/login')
  const reloads = []
  location.assign = url => reloads.push(['assign', url])
  location.replace = url => reloads.push(['replace', url])
  const entries = [location.href]
  let index = 0
  const window = new EventTarget()
  window.location = location
  window.scrollTo = () => {}
  window.history = {
    pushState(_state, _title, url) {
      entries.splice(++index, entries.length, url)
      location.href = url
    },
    replaceState(_state, _title, url) { entries[index] = url; location.href = url },
    go(offset) {
      index += offset
      location.href = entries[index]
      window.dispatchEvent(new Event('popstate'))
    },
  }
  class Element {
    constructor(link = null) { this.link = link }
    closest(selector) { return selector === 'a[href]' ? this.link : null }
  }
  class Anchor extends Element {
    constructor(href, target = '', download = false) {
      super(); this.link = this; this.href = href; this.target = target; this.download = download
    }
    hasAttribute(name) { return name === 'download' && this.download }
  }
  const storage = () => {
    const values = new Map()
    return { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value) }
  }
  const players = []
  class Audio extends EventTarget {
    paused = true
    readyState = 1
    currentTime = 0
    starts = 0
    constructor(src) { super(); this.src = src; players.push(this) }
    load() {}
    play() { this.starts++; this.paused = false; return Promise.resolve() }
    pause() { this.paused = true }
  }
  const replacements = { window, Element, HTMLAnchorElement: Anchor, Audio, localStorage: storage(), sessionStorage: storage() }
  for (const [key, value] of Object.entries(replacements)) {
    Object.defineProperty(globalThis, key, { value, writable: true, configurable: true })
  }
  let stop = () => {}
  t.after(() => {
    stop()
    for (const key of Object.keys(replacements)) {
      if (descriptors[key]) Object.defineProperty(globalThis, key, descriptors[key])
      else delete globalThis[key]
    }
  })
  const { navigate, installNavigation, subscribeToNavigation, getRoute } = await import('../src/router/navigation.ts')
  const { installAudio } = await import('../src/audio/audioManager.ts')
  const cleanupAudio = installAudio()
  const cleanupNavigation = installNavigation()
  const routes = []
  const unsubscribe = subscribeToNavigation(() => routes.push(getRoute()))
  stop = () => { unsubscribe(); cleanupNavigation(); cleanupAudio() }
  const bgm = players.find(player => player.src === '/audio/bgm.mp3')
  bgm.currentTime = 25
  function click(link, props = {}) {
    const event = new Event('click', { cancelable: true })
    Object.defineProperty(event, 'target', { value: new Element(link) })
    Object.assign(event, { button: 0, detail: 1, ...props })
    window.dispatchEvent(event)
    return event
  }

  navigate('/home') // Login's programmatic redirect.
  assert.equal(getRoute(), '/home')
  assert.equal(click(new Anchor('/settings')).defaultPrevented, true)
  navigate('/home') // Saving settings.
  click(new Anchor('/game?mode=cpu'))
  click(new Anchor('/game?mode=online'))
  assert.deepEqual(routes.slice(-2), ['/game?mode=cpu', '/game?mode=online'])
  window.history.go(-1)
  assert.equal(routes.at(-1), '/game?mode=cpu')
  window.history.go(1)
  assert.equal(routes.at(-1), '/game?mode=online')
  const length = entries.length
  navigate('/login', { replace: true })
  assert.equal(entries.length, length)
  assert.equal(routes.at(-1), '/login')
  assert.deepEqual(reloads, [], 'screen changes must not reload the document')
  assert.equal(bgm.paused, false)
  assert.equal(bgm.currentTime, 25)
  assert.equal(bgm.starts, 1, 'screen changes must not restart BGM')
  assert.equal(players.filter(player => player.src === '/audio/bgm.mp3').length, 1)

  for (const [link, props] of [
    [new Anchor('/home'), { ctrlKey: true }],
    [new Anchor('/home'), { metaKey: true }],
    [new Anchor('/home'), { shiftKey: true }],
    [new Anchor('/home'), { altKey: true }],
    [new Anchor('/home'), { button: 1 }],
    [new Anchor('/home', '_blank'), {}],
    [new Anchor('/home', '', true), {}],
    [new Anchor('https://example.com/home'), {}],
    [new Anchor('/audio/bgm.mp3'), {}],
    [new Anchor('#help'), {}],
  ]) {
    assert.equal(click(link, props).defaultPrevented, false, 'native link behavior must be preserved')
  }
  const updates = routes.length
  cleanupNavigation()
  assert.equal(click(new Anchor('/home')).defaultPrevented, false)
  unsubscribe()
  navigate('/home')
  assert.equal(routes.length, updates)
})
