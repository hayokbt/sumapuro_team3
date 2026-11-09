type HomeFormProps = {
  state: 'checking' | 'loggedIn' | 'loggedOut' | 'error'
  errorMessage: string
  isSubmitting: boolean
  onLogout: () => void
}

function HomeForm({ state, errorMessage, isSubmitting, onLogout }: HomeFormProps) {
  const labels = {
    checking: 'ログイン状態を確認中...',
    loggedIn: 'ログイン中',
    loggedOut: 'ログアウト状態',
    error: 'ログイン状態を確認できません',
  }

  return (
    <main className="home-page">
      <section className="home-menu" aria-labelledby="home-title">
        <p className="home-eyebrow">GOMOKU TAISEN</p>
        <h1 id="home-title" className="home-title">
          <span>五目並べ</span>
          <span>おんらいん</span>
        </h1>
        <p className="home-description" role="status">{labels[state]}</p>
        {errorMessage && <p className="home-status" role="alert">{errorMessage}</p>}
        {state === 'loggedIn' && (
          <>
            <a className="home-button home-button-primary" href="/game?mode=cpu">
              CPUと対戦
              <span className="home-mode-detail">ひとりで、すぐに対局を始める</span>
            </a>
            <a className="home-button home-button-secondary" href="/game?mode=online">
              オンライン対戦
              <span className="home-mode-detail">待機中のプレイヤーと自動マッチング</span>
            </a>
            <p className="home-status">15×15盤・ちょうど5個で勝利。6個以上では勝ちになりません。</p>
            <button
              className="home-logout"
              type="button"
              disabled={isSubmitting}
              onClick={onLogout}
            >
              {isSubmitting ? 'ログアウト中...' : 'ログアウト'}
            </button>
          </>
        )}
        {state === 'loggedOut' && <a href="/login">ログイン画面へ</a>}
        <a className="home-button home-button-secondary" href="/ranking">ランキングを見る</a>
      </section>
    </main>
  )
}

export default HomeForm
