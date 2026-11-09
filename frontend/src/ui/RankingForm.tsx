import type { RankingEntry } from '../api/ranking'

type RankingFormProps = {
  entries: RankingEntry[]
  isLoading: boolean
  errorMessage: string
}

function RankingForm({ entries, isLoading, errorMessage }: RankingFormProps) {
  return (
    <main className="ranking-page">
      <section className="ranking-screen" aria-labelledby="ranking-title">
        <a className="game-back-link" href="/home">ホームへ戻る</a>
        <h1 id="ranking-title">ランキング</h1>
        {isLoading && <p role="status">読み込み中…</p>}
        {errorMessage && <p className="home-status" role="alert">{errorMessage}</p>}
        {!isLoading && !errorMessage && entries.length === 0 && <p role="status">ランキングデータがありません。</p>}
        {!isLoading && !errorMessage && entries.length > 0 && (
          <table className="ranking-table">
            <thead>
              <tr>
                <th scope="col">順位</th>
                <th scope="col">アカウント名</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry) => (
                <tr key={entry.position}>
                  <td>{entry.position}位</td>
                  <td>{entry.accountName}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </main>
  )
}

export default RankingForm
