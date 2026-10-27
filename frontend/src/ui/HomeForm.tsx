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
    <>
      <p>home</p>
      <p role="status">{labels[state]}</p>
      {errorMessage && <p role="alert">{errorMessage}</p>}
      {state === 'loggedIn' && (
        <button type="button" disabled={isSubmitting} onClick={onLogout}>
          {isSubmitting ? 'ログアウト中...' : 'ログアウト'}
        </button>
      )}
      {state === 'loggedOut' && <a href="/login">ログイン画面へ</a>}
    </>
  )
}

export default HomeForm
