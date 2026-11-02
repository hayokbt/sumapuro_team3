import { useEffect, useRef, useState } from 'react'
import { currentUser, logout } from '../api/authentication'
import HomeForm from '../ui/HomeForm'

type LoginState = 'checking' | 'loggedIn' | 'loggedOut' | 'error'

function HomeController() {
  const generation = useRef(0)
  const loggingOut = useRef(false)
  const [state, setState] = useState<LoginState>('checking')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    const controller = new AbortController()

    async function check() {
      if (loggingOut.current) return
      // Only the latest check may update the login state.
      const currentGeneration = ++generation.current
      try {
        const user = await currentUser(controller.signal)
        if (!controller.signal.aborted && generation.current === currentGeneration) {
          setState(user === null ? 'loggedOut' : 'loggedIn')
          setError('')
        }
      } catch {
        if (!controller.signal.aborted && generation.current === currentGeneration) {
          setState('error')
          setError('ログイン状態を確認できませんでした。')
        }
      }
    }

    void check()
    const recheck = () => { void check() }
    window.addEventListener('focus', recheck)
    const interval = window.setInterval(recheck, 30_000)
    return () => {
      controller.abort()
      window.removeEventListener('focus', recheck)
      window.clearInterval(interval)
    }
  }, [])

  async function handleLogout() {
    if (loggingOut.current) return
    // Ignore responses from checks started before this logout.
    loggingOut.current = true
    generation.current += 1
    setIsSubmitting(true)
    setError('')
    try {
      await logout()
      setState('loggedOut')
    } catch (error) {
      setError(error instanceof Error ? error.message : 'ログアウトできませんでした。')
    } finally {
      loggingOut.current = false
      setIsSubmitting(false)
    }
  }

  return (
    <HomeForm
      state={state}
      errorMessage={error}
      isSubmitting={isSubmitting}
      onLogout={handleLogout}
    />
  )
}

export default HomeController
