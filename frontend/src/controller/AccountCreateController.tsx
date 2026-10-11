import { useState, type SubmitEvent } from 'react'
import AccountCreateForm from '../ui/AccountCreateForm'

function AccountCreateController() {
  const [email, setEmail] = useState('')
  const [accountName, setAccountName] = useState('')
  const [password, setPassword] = useState('')

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
  }

  return (
    <AccountCreateForm
      email={email}
      accountName={accountName}
      password={password}
      onEmailChange={setEmail}
      onAccountNameChange={setAccountName}
      onPasswordChange={setPassword}
      onSubmit={handleSubmit}
    />
  )
}

export default AccountCreateController
