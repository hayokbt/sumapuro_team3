export type Settings = { bgmVolume: number; seVolume: number }
export type Profile = Settings & { userId: number; accountName: string; email: string; wins: number; losses: number; draws: number; rating: number }
async function request<T>(path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
 const response = await fetch(path, {
  method: body === undefined ? 'GET' : 'POST', cache: 'no-store', signal,
  headers: body === undefined ? undefined : { 'Content-Type': 'application/json', 'X-Requested-With': 'gomoku-web' },
  body: body === undefined ? undefined : JSON.stringify(body),
 })
 if (response.status === 401) throw new Error('ログインが必要です。ログイン画面から入り直してください。')
 const result = await response.json()
 if (!response.ok) throw new Error(result.error || '処理に失敗しました。')
 return result as T
}
export const getProfile = (signal?: AbortSignal) => request<Profile>('/api/account', undefined, signal)
export const updateProfile = (body: { accountName: string; currentPassword: string; password: string }) => request('/api/account', body)
export const getSettings = (signal?: AbortSignal) => request<Settings>('/api/account/settings', undefined, signal)
export const saveSettings = (body: Settings) => request('/api/account/settings', body)
