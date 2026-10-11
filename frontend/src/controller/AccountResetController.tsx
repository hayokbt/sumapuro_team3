import { useState, type SubmitEvent } from 'react'
import AccountResetForm from '../ui/AccountResetForm'

function AccountResetController() {
  const [accountName, setAccountName] = useState('')
  const [password, setPassword] = useState('')

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
  }

  return (
    <AccountResetForm
      accountName={accountName}
      password={password}
      onAccountNameChange={setAccountName}
      onPasswordChange={setPassword}
      onSubmit={handleSubmit}
    />
  )
}

export default AccountResetController
