import { useState, type SubmitEvent } from 'react'
import { requestAccountRecoveryCode, verifyAccountRecoveryCode, accountRecoveryStorageKey } from '../api/accountRecovery'
import AccountRecoveryForm from '../ui/AccountRecoveryForm'

function AccountRecoveryController() {
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [codeRequested, setCodeRequested] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [statusMessage, setStatusMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsSubmitting(true)
    setErrorMessage('')
    setStatusMessage('')
    try {
      if (!codeRequested) {
        const message = await requestAccountRecoveryCode(email)
        setCodeRequested(true)
        setStatusMessage(message)
        return
      }

      const resetToken = await verifyAccountRecoveryCode(email, code)
      try {
        sessionStorage.setItem(accountRecoveryStorageKey, JSON.stringify({ email, resetToken }))
      } catch {
        setErrorMessage('復旧情報を保存できませんでした。ブラウザーの設定を確認して再度お試しください。')
        return
      }
      window.location.assign('/account/reset')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'アカウント復旧を処理できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  function handleEditEmail() {
    setCodeRequested(false)
    setCode('')
    setStatusMessage('')
    setErrorMessage('')
  }

  return (
    <AccountRecoveryForm
      email={email}
      code={code}
      codeRequested={codeRequested}
      isSubmitting={isSubmitting}
      statusMessage={statusMessage}
      errorMessage={errorMessage}
      onEmailChange={setEmail}
      onCodeChange={setCode}
      onSubmit={handleSubmit}
      onEditEmail={handleEditEmail}
    />
  )
}

export default AccountRecoveryController
