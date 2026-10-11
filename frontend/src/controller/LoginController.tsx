import { useState, type SubmitEvent } from 'react'
import LoginForm from '../ui/LoginForm'

function LoginController() {
  const [accountName, setAccountName] = useState('')
  const [password, setPassword] = useState('')

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
  }

  return (
    <LoginForm
      accountName={accountName}
      password={password}
      onAccountNameChange={setAccountName}
      onPasswordChange={setPassword}
      onSubmit={handleSubmit}
    />
  )
}

export default LoginController
