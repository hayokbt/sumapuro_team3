export async function currentUser(signal?: AbortSignal): Promise<number | null> {
 const response = await fetch('/api/auth/me', { signal, cache: 'no-store' })
 if (response.status === 401) return null
 if (!response.ok) throw new Error('ログイン状態を確認できませんでした。')
 const result: unknown = await response.json()
 if (typeof result !== 'object' || result === null || !('userId' in result) || typeof result.userId !== 'number') throw new Error('サーバーから正しい応答を受け取れませんでした。')
 return result.userId
}
export async function logout(): Promise<void> {
 const response = await fetch('/api/logout', { method: 'POST', headers: { 'X-Requested-With': 'gomoku-web' } })
 if (!response.ok) throw new Error('ログアウトできませんでした。再度お試しください。')
}
