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
        <h1 id="home-title">五目並べ</h1>
        <p className="home-description" role="status">{labels[state]}</p>
        {errorMessage && <p className="home-status" role="alert">{errorMessage}</p>}
        {state === 'loggedIn' && (
          <>
            <a className="home-button home-button-primary" href="/game">
              ゲームを始める
            </a>
            <a className="home-button home-button-secondary" href="/account/create">
              アカウントを作って始める
            </a>
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
      </section>
    </main>
  )
}

export default HomeForm
