import { createServer } from 'node:http'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { apiRoutes, proxyAPI } from './server/api-proxy.mjs'
import { serveStatic } from './server/static-files.mjs'

const staticDirectory = resolve(fileURLToPath(new URL('./dist', import.meta.url)))
const backendURL = process.env.BACKEND_URL ?? 'http://backend:8080'
const port = Number(process.env.PORT ?? 3000)

if (!Number.isInteger(port) || port < 1 || port > 65535) {
  throw new Error('PORT must be a valid TCP port')
}

const server = createServer((request, response) => {
  let pathname
  try {
    pathname = new URL(request.url ?? '/', 'http://localhost').pathname
  } catch {
    response.writeHead(400)
    response.end()
    return
  }

  const backendPath = apiRoutes.get(pathname)
  if (backendPath) {
    void proxyAPI(request, response, backendPath, backendURL)
    return
  }
  serveStatic(request, response, pathname, staticDirectory)
})

server.listen(port, '0.0.0.0', () => {
  console.log(`Frontend BFF listening on :${port}`)
})
