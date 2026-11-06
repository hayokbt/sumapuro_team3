import { request as httpRequest } from 'node:http'
import { request as httpsRequest } from 'node:https'

export function proxyWebSocket(request, socket, head, backendURL) {
  const url = new URL(request.url ?? '/', 'http://localhost')
  if (url.pathname !== '/api/game/ws') {
    socket.end('HTTP/1.1 404 Not Found\r\nConnection: close\r\n\r\n')
    return
  }
  const target = new URL(`/game/ws${url.search}`, backendURL)
  const upstream = (target.protocol === 'https:' ? httpsRequest : httpRequest)(target, {
    method: 'GET', headers: { ...request.headers, host: target.host },
  })
  let backendSocket
  upstream.setTimeout(10_000, () => upstream.destroy(new Error('Upgrade timed out')))
  socket.on('error', () => { upstream.destroy(); backendSocket?.destroy() })
  socket.on('close', () => { upstream.destroy(); backendSocket?.destroy() })
  upstream.on('upgrade', (response, peer, backendHead) => {
    backendSocket = peer
    upstream.setTimeout(0)
    if (socket.destroyed) { peer.destroy(); return }
    const headers = []
    for (let i = 0; i < response.rawHeaders.length; i += 2) headers.push(`${response.rawHeaders[i]}: ${response.rawHeaders[i + 1]}`)
    socket.write(`HTTP/1.1 101 Switching Protocols\r\n${headers.join('\r\n')}\r\n\r\n`)
    peer.on('error', () => socket.destroy())
    peer.on('close', () => socket.destroy())
    if (backendHead.length) socket.write(backendHead)
    if (head.length) peer.write(head)
    socket.pipe(peer).pipe(socket)
  })
  upstream.on('response', (response) => {
    socket.end(`HTTP/1.1 ${response.statusCode} ${response.statusMessage}\r\nConnection: close\r\nContent-Length: 0\r\n\r\n`)
    response.resume()
  })
  upstream.on('error', () => {
    if (!socket.destroyed) socket.end('HTTP/1.1 502 Bad Gateway\r\nConnection: close\r\n\r\n')
  })
  upstream.end()
}
