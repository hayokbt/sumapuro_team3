import { useState, type SubmitEvent } from 'react'
import AccountRecoveryForm from '../ui/AccountRecoveryForm'

function AccountRecoveryController() {
  const [email, setEmail] = useState('')

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
  }

  return (
    <AccountRecoveryForm
      email={email}
      onEmailChange={setEmail}
      onSubmit={handleSubmit}
    />
  )
}

export default AccountRecoveryController
