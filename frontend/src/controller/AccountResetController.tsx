import { useEffect, useState, type SubmitEvent } from 'react'
import {
  accountRecoveryStorageKey,
  readAccountRecoveryState,
  resetAccount,
} from '../api/accountRecovery'
import AccountResetForm from '../ui/AccountResetForm'

function AccountResetController() {
  const [recoveryState] = useState(readAccountRecoveryState)
  const [accountName, setAccountName] = useState('')
  const [password, setPassword] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isCompleted, setIsCompleted] = useState(false)
  const [statusMessage, setStatusMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    if (!recoveryState) {
      window.location.replace('/account/recovery')
    }
  }, [recoveryState])

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!recoveryState) {
      setErrorMessage('確認コードを確認してから、アカウント復旧を続けてください。')
      return
    }

    setIsSubmitting(true)
    setErrorMessage('')
    setStatusMessage('')
    try {
      const message = await resetAccount({
        email: recoveryState.email,
        resetToken: recoveryState.resetToken,
        accountName,
        password,
      })
      try {
        sessionStorage.removeItem(accountRecoveryStorageKey)
      } catch (error) {
        console.error('Failed to clear account recovery state:', error)
      }
      setIsCompleted(true)
      setStatusMessage(message)
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'アカウント情報を更新できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!recoveryState) {
    return <p role="status">確認コードを確認してください。</p>
  }

  return (
    <AccountResetForm
      accountName={accountName}
      password={password}
      onAccountNameChange={setAccountName}
      onPasswordChange={setPassword}
      onSubmit={handleSubmit}
      isSubmitting={isSubmitting}
      isCompleted={isCompleted}
      statusMessage={statusMessage}
      errorMessage={errorMessage}
    />
  )
}

export default AccountResetController
