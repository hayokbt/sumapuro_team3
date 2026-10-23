const bodyLimit = 16 * 1024

export const apiRoutes = new Map([
  ['/api/account/create', '/account/create'],
  ['/api/account/verify', '/account/verify'],
  ['/api/login', '/login'],
])

function sendJSON(response, status, payload) {
  response.writeHead(status, { 'content-type': 'application/json; charset=utf-8' })
  response.end(JSON.stringify(payload))
}

export async function proxyAPI(request, response, backendPath, backendURL) {
  if (request.method !== 'POST') {
    response.setHeader('allow', 'POST')
    sendJSON(response, 405, { error: 'method not allowed' })
    return
  }

  const chunks = []
  let bodySize = 0
  try {
    for await (const chunk of request) {
      bodySize += chunk.length
      if (bodySize > bodyLimit) {
        sendJSON(response, 413, { error: 'request body is too large' })
        return
      }
      chunks.push(chunk)
    }
  } catch (error) {
    console.error('Failed to read client request:', error)
    if (!response.destroyed) {
      sendJSON(response, 400, { error: 'リクエストを読み取れませんでした。' })
    }
    return
  }

  try {
    const backendResponse = await fetch(new URL(backendPath, backendURL), {
      method: 'POST',
      headers: { 'content-type': request.headers['content-type'] ?? 'application/json' },
      body: Buffer.concat(chunks),
      signal: AbortSignal.timeout(30_000),
    })
    response.writeHead(backendResponse.status, {
      'content-type': backendResponse.headers.get('content-type') ?? 'application/json',
    })
    response.end(Buffer.from(await backendResponse.arrayBuffer()))
  } catch (error) {
    console.error('Backend request failed:', error)
    if (!response.destroyed) {
      sendJSON(response, 502, { error: 'バックエンドに接続できませんでした。' })
    }
  }
}
