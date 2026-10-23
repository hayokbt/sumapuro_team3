import type { SubmitEvent } from 'react'

type AccountRecoveryFormProps = {
  email: string
  code: string
  codeRequested: boolean
  isSubmitting: boolean
  statusMessage: string
  errorMessage: string
  onEmailChange: (value: string) => void
  onCodeChange: (value: string) => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
  onEditEmail: () => void
}

function AccountRecoveryForm({
  email,
  code,
  codeRequested,
  isSubmitting,
  statusMessage,
  errorMessage,
  onEmailChange,
  onCodeChange,
  onSubmit,
  onEditEmail,
}: AccountRecoveryFormProps) {
  return (
    <form onSubmit={onSubmit}>
      <label>
        メールアドレス
        <input
          type="email"
          name="email"
          autoComplete="email"
          value={email}
          onChange={(event) => onEmailChange(event.target.value)}
          required
          readOnly={codeRequested}
          disabled={isSubmitting}
        />
      </label>
      {codeRequested && (
        <label>
          確認コード
          <input
            type="text"
            name="code"
            autoComplete="one-time-code"
            inputMode="numeric"
            pattern="[0-9]{6}"
            maxLength={6}
            value={code}
            onChange={(event) => onCodeChange(event.target.value)}
            required
            disabled={isSubmitting}
          />
        </label>
      )}
      {statusMessage && <p role="status">{statusMessage}</p>}
      {errorMessage && <p role="alert">{errorMessage}</p>}
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting
          ? '処理中...'
          : codeRequested
            ? '確認コードを確認'
            : '確認コードを送信'}
      </button>
      {codeRequested && (
        <button type="button" onClick={onEditEmail} disabled={isSubmitting}>
          メールアドレスを変更
        </button>
      )}
    </form>
  )
}

export default AccountRecoveryForm
