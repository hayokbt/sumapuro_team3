import assert from 'node:assert/strict'
import { createServer } from 'node:http'
import { once } from 'node:events'
import test from 'node:test'
import { proxyAPI } from './api-proxy.mjs'

test('BFF forwards GET, auth cookies, origin and individual Set-Cookie headers', async () => {
 const backend = createServer((req, res) => {
  assert.equal(req.method, 'GET')
  assert.equal(req.headers.cookie, 'gomoku_auth=test-jwt')
  assert.equal(req.headers.origin, 'http://localhost:5173')
  res.writeHead(200, { 'content-type': 'application/json', 'cache-control': 'no-store', 'set-cookie': ['gomoku_auth=new; HttpOnly; Path=/', 'other=value; Path=/'] })
  res.end('{"userId":1}')
 })
 backend.listen(0, '127.0.0.1'); await once(backend, 'listening')
 const bff = createServer((req, res) => { void proxyAPI(req, res, '/auth/me', `http://127.0.0.1:${backend.address().port}`) })
 bff.listen(0, '127.0.0.1'); await once(bff, 'listening')
 try {
  const response = await fetch(`http://127.0.0.1:${bff.address().port}/api/auth/me`, { headers: { cookie: 'gomoku_auth=test-jwt', origin: 'http://localhost:5173' } })
  assert.equal(response.status, 200)
  assert.equal(response.headers.getSetCookie().length, 2)
  assert.equal(response.headers.get('cache-control'), 'no-store')
  assert.deepEqual(await response.json(), { userId: 1 })
 } finally { await Promise.all([new Promise(resolve => bff.close(resolve)), new Promise(resolve => backend.close(resolve))]) }
})
