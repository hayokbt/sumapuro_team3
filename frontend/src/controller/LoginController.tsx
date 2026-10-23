import { useState, type SubmitEvent } from 'react'
import { requestLogin } from '../api/login'
import LoginForm from '../ui/LoginForm'

function LoginController() {
  const [accountName, setAccountName] = useState('')
  const [password, setPassword] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsSubmitting(true)
    setErrorMessage('')
    try {
      await requestLogin({ accountName, password })
      window.location.assign('/home')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'ログインを処理できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <LoginForm
      accountName={accountName}
      password={password}
      onAccountNameChange={setAccountName}
      onPasswordChange={setPassword}
      onSubmit={handleSubmit}
      isSubmitting={isSubmitting}
      errorMessage={errorMessage}
    />
  )
}

export default LoginController
