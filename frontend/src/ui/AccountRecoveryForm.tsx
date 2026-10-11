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
    </form>
  )
}

export default AccountRecoveryForm
