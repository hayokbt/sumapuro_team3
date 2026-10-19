import { createReadStream, existsSync, statSync } from 'node:fs'
import { extname, relative, resolve, sep } from 'node:path'

const contentTypes = new Map([
  ['.css', 'text/css; charset=utf-8'],
  ['.html', 'text/html; charset=utf-8'],
  ['.ico', 'image/x-icon'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.png', 'image/png'],
  ['.svg', 'image/svg+xml'],
  ['.txt', 'text/plain; charset=utf-8'],
  ['.woff', 'font/woff'],
  ['.woff2', 'font/woff2'],
])

export function serveStatic(request, response, pathname, staticDirectory) {
  if (request.method !== 'GET' && request.method !== 'HEAD') {
    response.setHeader('allow', 'GET, HEAD')
    response.writeHead(405)
    response.end()
    return
  }

  let decodedPath
  try {
    decodedPath = decodeURIComponent(pathname)
  } catch {
    response.writeHead(400)
    response.end()
    return
  }
  if (decodedPath.includes('\0')) {
    response.writeHead(400)
    response.end()
    return
  }

  const requestedFile = resolve(staticDirectory, `.${decodedPath}`)
  const relativePath = relative(staticDirectory, requestedFile)
  if (relativePath.startsWith(`..${sep}`) || relativePath === '..') {
    response.writeHead(404)
    response.end()
    return
  }

  let requestedIsFile = false
  try {
    requestedIsFile = existsSync(requestedFile) && statSync(requestedFile).isFile()
  } catch (error) {
    console.error('Failed to inspect static file:', error)
    response.writeHead(500)
    response.end()
    return
  }
  const filePath = requestedIsFile
    ? requestedFile
    : decodedPath.startsWith('/api/')
      ? null
      : resolve(staticDirectory, 'index.html')
  if (!filePath || !existsSync(filePath)) {
    response.writeHead(404)
    response.end()
    return
  }

  response.writeHead(200, {
    'content-type': contentTypes.get(extname(filePath)) ?? 'application/octet-stream',
    'cache-control': filePath.endsWith('index.html') ? 'no-cache' : 'public, max-age=31536000, immutable',
  })
  if (request.method === 'HEAD') {
    response.end()
    return
  }
  createReadStream(filePath)
    .on('error', (error) => {
      console.error('Failed to stream static file:', error)
      if (!response.headersSent) {
        response.writeHead(500)
        response.end()
      } else {
        response.destroy(error)
      }
    })
    .pipe(response)
}
