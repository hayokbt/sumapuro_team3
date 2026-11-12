export type RankingEntry = {
  rating: number
  position: number
  accountName: string
}

export async function fetchRanking(signal?: AbortSignal): Promise<RankingEntry[]> {
  const response = await fetch('/api/ranking', { signal, cache: 'no-store' })
  if (!response.ok) {
    throw new Error('ランキングを取得できませんでした。')
  }

  const result: unknown = await response.json()
  if (!Array.isArray(result) || !result.every(isRankingEntry)) {
    throw new Error('ランキングの応答が正しくありません。')
  }
  return result
}

function isRankingEntry(value: unknown): value is RankingEntry {
  return typeof value === 'object'
    && value !== null
    && 'rating' in value
    && typeof value.rating === 'number'
    && Number.isFinite(value.rating)
    && 'position' in value
    && typeof value.position === 'number'
    && 'accountName' in value
    && typeof value.accountName === 'string'
}
