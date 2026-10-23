type APIResponse = {
  message?: string
  error?: string
  resetToken?: string
}

type ResetAccountData = {
  email: string
  resetToken: string
  accountName: string
  password: string
}

export type AccountRecoveryState = {
  email: string
  resetToken: string
}

export const accountRecoveryStorageKey = 'accountRecovery'

function isAPIResponse(value: unknown): value is APIResponse {
  return typeof value === 'object'
    && value !== null
    && (!('message' in value) || typeof value.message === 'string')
    && (!('error' in value) || typeof value.error === 'string')
    && (!('resetToken' in value) || typeof value.resetToken === 'string')
}

async function postRecoveryRequest(path: string, body: unknown): Promise<APIResponse> {
  let response: Response
  try {
    response = await fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
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
    throw new Error(result.error ?? 'アカウント復旧を処理できませんでした。')
  }
  return result
}

export async function requestAccountRecoveryCode(email: string): Promise<string> {
  const result = await postRecoveryRequest('/api/account/recovery/request', { email })
  if (typeof result.message !== 'string') {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }
  return result.message
}

export async function verifyAccountRecoveryCode(email: string, code: string): Promise<string> {
  const result = await postRecoveryRequest('/api/account/recovery/verify', { email, code })
  if (typeof result.resetToken !== 'string' || result.resetToken.length === 0) {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }
  return result.resetToken
}

export async function resetAccount(data: ResetAccountData): Promise<string> {
  const result = await postRecoveryRequest('/api/account/recovery/reset', data)
  if (typeof result.message !== 'string') {
    throw new Error('サーバーから正しい応答を受け取れませんでした。')
  }
  return result.message
}

export function readAccountRecoveryState(): AccountRecoveryState | null {
  try {
    const storedState = sessionStorage.getItem(accountRecoveryStorageKey)
    if (!storedState) {
      return null
    }
    const parsed: unknown = JSON.parse(storedState)
    if (
      typeof parsed === 'object'
      && parsed !== null
      && 'email' in parsed
      && typeof parsed.email === 'string'
      && 'resetToken' in parsed
      && typeof parsed.resetToken === 'string'
      && parsed.resetToken.length > 0
    ) {
      return { email: parsed.email, resetToken: parsed.resetToken }
    }
  } catch {
    return null
  }
  return null
}
