import type { SubmitEvent } from 'react'

type AccountCreateFormProps = {
  email: string
  accountName: string
  password: string
  onEmailChange: (value: string) => void
  onAccountNameChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
}

function AccountCreateForm({
  email,
  accountName,
  password,
  onEmailChange,
  onAccountNameChange,
  onPasswordChange,
  onSubmit,
}: AccountCreateFormProps) {
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
        />
      </label>
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
      <button type="submit">アカウント作成</button>
    </form>
  )
}

export default AccountCreateForm
