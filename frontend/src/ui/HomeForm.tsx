function HomeForm() {
  return (
    <main className="home-page">
      <section className="home-menu" aria-labelledby="home-title">
        <p className="home-eyebrow">GOMOKU TAISEN</p>
        <h1 id="home-title">五目並べ</h1>
        <p className="home-description">対局の準備ができました。</p>
        <a className="home-button home-button-primary" href="/game">
          ゲームを始める
        </a>
        <a className="home-button home-button-secondary" href="/account/create">
          アカウントを作って始める
        </a>
        <button
          className="home-logout"
          type="button"
          onClick={() => window.location.replace('/login')}
        >
          ログアウト
        </button>
      </section>
    </main>
  )
}

export default HomeForm
