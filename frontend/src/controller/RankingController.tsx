import { useEffect, useState } from 'react'
import { fetchRanking, type RankingEntry } from '../api/ranking'
import RankingForm from '../ui/RankingForm'

function RankingController() {
  const [entries, setEntries] = useState<RankingEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    async function loadRanking() {
      try {
        setEntries(await fetchRanking(controller.signal))
      } catch (error) {
        if (!controller.signal.aborted) {
          setErrorMessage(error instanceof Error ? error.message : 'ランキングを取得できませんでした。')
        }
      } finally {
        if (!controller.signal.aborted) setIsLoading(false)
      }
    }

    void loadRanking()
    return () => controller.abort()
  }, [])

  return <RankingForm entries={entries} isLoading={isLoading} errorMessage={errorMessage} />
}

export default RankingController
