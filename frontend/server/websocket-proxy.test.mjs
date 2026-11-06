import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createServer, request } from 'node:http'
import { once } from 'node:events'
import { proxyWebSocket } from './websocket-proxy.mjs'

test('WebSocket handshake, credentials and bidirectional frames pass through Node', async (t) => {
  const backend = createServer()
  const peers = new Set()
  backend.on('upgrade', (req, socket, head) => {
    assert.equal(req.url, '/game/ws?mode=online')
    assert.equal(req.headers.cookie, 'gomoku_auth=test')
    assert.equal(req.headers.origin, 'http://game.test')
    socket.write('HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n')
    if (head.length) socket.write(head)
    socket.on('data', (data) => socket.write(data))
  })
  backend.listen(0, '127.0.0.1')
  await once(backend, 'listening')
  const bff = createServer()
  bff.on('upgrade', (req, socket, head) => proxyWebSocket(req, socket, head, `http://127.0.0.1:${backend.address().port}`))
  for (const server of [backend, bff]) server.on('connection', (socket) => { peers.add(socket); socket.on('close', () => peers.delete(socket)) })
  t.after(() => { for (const socket of peers) socket.destroy(); backend.close(); bff.close() })
  bff.listen(0, '127.0.0.1')
  await once(bff, 'listening')
  const req = request(`http://127.0.0.1:${bff.address().port}/api/game/ws?mode=online`, {
    headers: { Connection: 'Upgrade', Upgrade: 'websocket', Cookie: 'gomoku_auth=test', Origin: 'http://game.test' },
  })
  req.end()
  const [response, socket] = await once(req, 'upgrade')
  assert.equal(response.statusCode, 101)
  const frame = Buffer.from([0x81, 0x82, 1, 2, 3, 4, 0x69, 0x6b])
  socket.write(frame)
  const [data] = await once(socket, 'data')
  assert.deepEqual(data, frame)
  socket.destroy()
})

test('WebSocket backend authentication rejection is forwarded', async (t) => {
  const backend = createServer((_req, res) => { res.writeHead(401); res.end() })
  backend.listen(0, '127.0.0.1'); await once(backend, 'listening')
  const bff = createServer()
  bff.on('upgrade', (req, socket, head) => proxyWebSocket(req, socket, head, `http://127.0.0.1:${backend.address().port}`))
  bff.listen(0, '127.0.0.1'); await once(bff, 'listening')
  t.after(() => { backend.closeAllConnections(); backend.close(); bff.closeAllConnections(); bff.close() })
  const req = request(`http://127.0.0.1:${bff.address().port}/api/game/ws?mode=cpu`, { headers: { Connection: 'Upgrade', Upgrade: 'websocket' } })
  req.end()
  const [response] = await once(req, 'response')
  assert.equal(response.statusCode, 401)
  response.resume()
})
