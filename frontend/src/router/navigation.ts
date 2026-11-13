const NAVIGATION_EVENT = 'app:navigation'
const APP_PATHS = new Set([
  '/', '/login', '/home', '/ranking', '/game', '/settings',
  '/account', '/account/create', '/account/recovery', '/account/reset',
])

export function getRoute(): string {
  return window.location.pathname + window.location.search
}

export function subscribeToNavigation(listener: () => void): () => void {
  window.addEventListener('popstate', listener)
  window.addEventListener(NAVIGATION_EVENT, listener)
  return () => {
    window.removeEventListener('popstate', listener)
    window.removeEventListener(NAVIGATION_EVENT, listener)
  }
}

export function navigate(to: string, options: { replace?: boolean } = {}) {
  const url = new URL(to, window.location.href)
  if (url.origin !== window.location.origin || !APP_PATHS.has(url.pathname)) {
    if (options.replace) window.location.replace(url.href)
    else window.location.assign(url.href)
    return
  }
  if (url.href === window.location.href) return
  if (options.replace) window.history.replaceState(null, '', url.href)
  else window.history.pushState(null, '', url.href)
  window.dispatchEvent(new Event(NAVIGATION_EVENT))
  window.scrollTo(0, 0)
}

// Keep native links accessible, including new tabs, downloads and hash links.
// Only ordinary clicks between app screens use the persistent React root.
export function installNavigation(): () => void {
  const click = (event: MouseEvent) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    const link = event.target instanceof Element ? event.target.closest('a[href]') : null
    if (!(link instanceof HTMLAnchorElement) || link.hasAttribute('download') || (link.target && link.target.toLowerCase() !== '_self')) return
    const url = new URL(link.href, window.location.href)
    if (url.origin !== window.location.origin || !APP_PATHS.has(url.pathname)) return
    if (url.pathname === window.location.pathname && url.search === window.location.search && url.hash) return
    event.preventDefault()
    navigate(url.href)
  }
  window.addEventListener('click', click)
  return () => window.removeEventListener('click', click)
}
