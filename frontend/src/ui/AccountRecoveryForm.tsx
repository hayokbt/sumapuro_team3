import type { SubmitEvent } from 'react'

type AccountRecoveryFormProps = {
  email: string
  onEmailChange: (value: string) => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
}

function AccountRecoveryForm({
  email,
  onEmailChange,
  onSubmit,
}: AccountRecoveryFormProps) {
  return (
    <div className="form-container account-recovery-form">
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
        <button type="submit">アカウント復旧</button>
        <button type="button" onClick={() => window.location.assign('/login')}>
          戻る
        </button>
      </form>
    </div>
  )
}

export default AccountRecoveryForm
