import type { SubmitEvent } from 'react'

type AccountResetFormProps = {
  accountName: string
  password: string
  onAccountNameChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
}

function AccountResetForm({
  accountName,
  password,
  onAccountNameChange,
  onPasswordChange,
  onSubmit,
}: AccountResetFormProps) {
  return (
    <form onSubmit={onSubmit}>
      <label>
        アカウント名
        <input
          type="text"
          name="accountName"
          autoComplete="username"
          value={accountName}
          onChange={(event) => onAccountNameChange(event.target.value)}
          required
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
        />
      </label>
      <button type="submit">パスワード再設定</button>
    </form>
  )
}

export default AccountResetForm
