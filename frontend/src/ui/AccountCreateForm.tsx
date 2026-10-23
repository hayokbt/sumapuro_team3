type AccountCreateFormProps = {
  email: string
  accountName: string
  password: string
  verificationCode: string
  onEmailChange: (value: string) => void
  onAccountNameChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onVerificationCodeChange: (value: string) => void
  onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
  onVerify: (event: React.SubmitEvent<HTMLFormElement>) => void
  onResend: () => void
  verificationRequested: boolean
  accountCreated: boolean
  isSubmitting: boolean
  isVerifying: boolean
  statusMessage: string
  errorMessage: string
}

function AccountCreateForm({
  email,
  accountName,
  password,
  verificationCode,
  onEmailChange,
  onAccountNameChange,
  onPasswordChange,
  onVerificationCodeChange,
  onSubmit,
  onVerify,
  onResend,
  verificationRequested,
  accountCreated,
  isSubmitting,
  isVerifying,
  statusMessage,
  errorMessage,
}: AccountCreateFormProps) {
  return (
    <div className="form-container account-create-form">
      <form onSubmit={onSubmit}>
        <label>
          メールアドレス
          <input
            type="email"
            name="email"
            autoComplete="email"
            value={email}
            onChange={(event) => onEmailChange(event.target.value)}
            disabled={verificationRequested || accountCreated}
            required
          />
        </label>
        <label>
          アカウント名
          <input
            type="text"
            name="accountName"
            autoComplete="username"
            maxLength={50}
            value={accountName}
            onChange={(event) => onAccountNameChange(event.target.value)}
            disabled={verificationRequested || accountCreated}
            required
          />
        </label>
        <label>
          パスワード（8文字以上）
          <input
            type="password"
            name="password"
            autoComplete="new-password"
            minLength={8}
            maxLength={72}
            value={password}
            onChange={(event) => onPasswordChange(event.target.value)}
            disabled={verificationRequested || accountCreated}
            required
          />
        </label>
        {!verificationRequested && (
          <button type="submit" disabled={isSubmitting || accountCreated}>
            {isSubmitting ? '送信中…' : '確認コードをメールで送信'}
          </button>
        )}
      </form>

      {verificationRequested && !accountCreated && (
        <form onSubmit={onVerify}>
          <label>
            メールで受け取った確認コード
            <input
              type="text"
              name="verificationCode"
              autoComplete="one-time-code"
              inputMode="numeric"
              pattern="[0-9]{6}"
              maxLength={6}
              value={verificationCode}
              onChange={(event) => onVerificationCodeChange(event.target.value.replace(/\D/g, '').slice(0, 6))}
              disabled={isVerifying}
              required
            />
          </label>
          <button type="submit" disabled={isVerifying || verificationCode.length !== 6}>
            {isVerifying ? '確認中…' : 'コードを確認してアカウントを作成'}
          </button>
          <button type="button" onClick={onResend} disabled={isSubmitting || isVerifying}>
            {isSubmitting ? '送信中…' : '確認コードを再送する'}
          </button>
        </form>
      )}

      {statusMessage && <p role="status">{statusMessage}</p>}
      {errorMessage && <p role="alert">{errorMessage}</p>}
      <button type="button" onClick={() => window.location.assign('/login')}>
        戻る
      </button>
    </div>
  )
}

export default AccountCreateForm