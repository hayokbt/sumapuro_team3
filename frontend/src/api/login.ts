type LoginData = {
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

export async function requestLogin(data: LoginData): Promise<string> {
  let response: Response
  try {
    response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
  } catch {
    throw new Error('サーバーに接続できませんでした。時間をおいて再度お試しください。')
  }

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
    throw new Error(result.error ?? 'ログインを処理できませんでした。')
  }
  if (typeof result.message !== 'string') {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }
  return result.message
}
