import { useState, type SubmitEvent } from 'react'
import { requestAccountVerification, verifyAccountEmail } from '../api/accountRegistration'
import AccountCreateForm from '../ui/AccountCreateForm'

function AccountCreateController() {
  const [email, setEmail] = useState('')
  const [accountName, setAccountName] = useState('')
  const [password, setPassword] = useState('')
  const [verificationCode, setVerificationCode] = useState('')
  const [verificationRequested, setVerificationRequested] = useState(false)
  const [accountCreated, setAccountCreated] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isVerifying, setIsVerifying] = useState(false)
  const [statusMessage, setStatusMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsSubmitting(true)
    setErrorMessage('')
    setStatusMessage('')
    try {
      const message = await requestAccountVerification({ email, accountName, password })
      setVerificationRequested(true)
      setStatusMessage(message)
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '確認メールを送信できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleVerify(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsVerifying(true)
    setErrorMessage('')
    setStatusMessage('')
    try {
      const message = await verifyAccountEmail(email, verificationCode)
      setAccountCreated(true)
      setStatusMessage(message)
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '確認コードを検証できませんでした。')
    } finally {
      setIsVerifying(false)
    }
  }

  async function handleResend() {
    setIsSubmitting(true)
    setVerificationCode('')
    setErrorMessage('')
    setStatusMessage('')
    try {
      const message = await requestAccountVerification({ email, accountName, password })
      setStatusMessage(message)
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '確認メールを送信できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <AccountCreateForm
      email={email}
      accountName={accountName}
      password={password}
      verificationCode={verificationCode}
      onEmailChange={setEmail}
      onAccountNameChange={setAccountName}
      onPasswordChange={setPassword}
      onVerificationCodeChange={setVerificationCode}
      onSubmit={handleSubmit}
      onVerify={handleVerify}
      onResend={handleResend}
      verificationRequested={verificationRequested}
      accountCreated={accountCreated}
      isSubmitting={isSubmitting}
      isVerifying={isVerifying}
      statusMessage={statusMessage}
      errorMessage={errorMessage}
    />
  )
}

export default AccountCreateController
