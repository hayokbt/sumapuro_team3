import type { SubmitEvent } from 'react'

type LoginFormProps = {
  accountName: string
  password: string
  onAccountNameChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
}

function LoginForm({
  accountName,
  password,
  onAccountNameChange,
  onPasswordChange,
  onSubmit,
}: LoginFormProps) {
  return (
    <div className="form-container login-form">
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
            autoComplete="current-password"
            value={password}
            onChange={(event) => onPasswordChange(event.target.value)}
            required
          />
        </label>
        <button type="submit">ログイン</button>
        <p>
          <a href="/account/create">アカウントがない方</a>
        </p>
        <p>
          <a href="/account/recovery">アカウント名またはパスワードを忘れた方</a>
        </p>
      </form>
    </div>
  )
}

export default LoginForm
