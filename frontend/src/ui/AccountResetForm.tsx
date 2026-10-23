import type { SubmitEvent } from 'react'

type AccountResetFormProps = {
  accountName: string
  password: string
  onAccountNameChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
  isSaving: boolean
  isCompleted: boolean
  statusMessage: string
  errorMessage: string
}

function AccountResetForm({
  accountName,
  password,
  onAccountNameChange,
  onPasswordChange,
  onSubmit,
  isSaving,
  isCompleted,
  statusMessage,
  errorMessage,
}: AccountResetFormProps) {
  return (
    <form onSubmit={onSubmit}>
      {statusMessage && <p role="status">{statusMessage}</p>}
      {errorMessage && <p role="alert">{errorMessage}</p>}
      <label>
        アカウント名
        <input
          type="text"
          name="accountName"
          autoComplete="username"
          value={accountName}
          onChange={(event) => onAccountNameChange(event.target.value)}
          required
          maxLength={50}
          disabled={isSaving || isCompleted}
        />
      </label>
      <label>
        パスワード
        <input
          type="password"
          name="password"
          autoComplete="new-password"
          value={password}
          onChange={(event) => onPasswordChange(event.target.value)}
          required
          minLength={8}
          maxLength={72}
          disabled={isSaving || isCompleted}
        />
      </label>
      {isCompleted
        ? <a href="/login">ログイン画面へ</a>
        : (
          <button type="submit" disabled={isSaving}>
            {isSaving ? '更新中...' : 'アカウント情報を更新'}
          </button>
        )}
    </form>
  )
}

export default AccountResetForm
