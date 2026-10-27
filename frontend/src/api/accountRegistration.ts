type RegistrationData = {
  email: string
  accountName: string
  password: string
}

type APIResponse = {
  message?: string
  error?: string
}

function isAPIResponse(value: unknown): value is APIResponse {
  return typeof value === 'object'
    && value !== null
    && (!('message' in value) || typeof value.message === 'string')
    && (!('error' in value) || typeof value.error === 'string')
}

async function postRegistrationRequest(path: string, body: unknown): Promise<string> {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'gomoku-web' },
    body: JSON.stringify(body),
  })

  let result: unknown
  try {
    result = await response.json()
  } catch {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }
  if (!isAPIResponse(result)) {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }

  if (!response.ok) {
    throw new Error(result.error ?? 'リクエストを処理できませんでした。')
  }
  if (typeof result.message !== 'string') {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }
  return result.message
}

export function requestAccountVerification(data: RegistrationData): Promise<string> {
  return postRegistrationRequest('/api/account/create', data)
}

export function verifyAccountEmail(email: string, code: string): Promise<string> {
  return postRegistrationRequest('/api/account/verify', { email, code })
}
