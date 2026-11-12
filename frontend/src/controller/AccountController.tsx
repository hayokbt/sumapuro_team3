import { useEffect, useState, type SubmitEvent } from 'react'
import { getProfile, updateProfile, type Profile } from '../api/account'

export default function AccountController() {
 const [profile, setProfile] = useState<Profile | null>(null)
 const [accountName, setAccountName] = useState('')
 const [currentPassword, setCurrentPassword] = useState('')
 const [password, setPassword] = useState('')
 const [busy, setBusy] = useState(false)
 const [error, setError] = useState('')
 const [message, setMessage] = useState('')
 useEffect(() => {
  const controller = new AbortController()
  getProfile(controller.signal).then(p => { setProfile(p); setAccountName(p.accountName) }).catch(e => {
   if (!controller.signal.aborted) setError(e instanceof Error ? e.message : '取得できませんでした。')
  })
  return () => controller.abort()
 }, [])
 async function submit(event: SubmitEvent<HTMLFormElement>) {
  event.preventDefault(); setBusy(true); setError(''); setMessage('')
  try {
   await updateProfile({ accountName, currentPassword, password })
   if (password) { window.location.assign('/login'); return }
   setCurrentPassword(''); setMessage('アカウント名を更新しました。')
  } catch (e) { setError(e instanceof Error ? e.message : '保存できませんでした。') }
  finally { setBusy(false) }
 }
 return <main className="account-flow-page"><section className="form-container account-flow-form">
  <h1>アカウント</h1>
  {error && <p role="alert">{error}</p>}
  {message && <p role="status">{message}</p>}
  {!profile && !error && <p role="status">読み込み中…</p>}
  {profile && <>
   <p>メールアドレス：{profile.email}</p>
   <p>レート：{profile.rating.toFixed(2)} ／ {profile.wins}勝 {profile.losses}敗 {profile.draws}分</p>
   <form onSubmit={submit}>
    <label>アカウント名<input required maxLength={50} autoComplete="username" value={accountName} onChange={e => setAccountName(e.target.value)} disabled={busy} /></label>
    <label>現在のパスワード<input type="password" required minLength={8} maxLength={72} autoComplete="current-password" value={currentPassword} onChange={e => setCurrentPassword(e.target.value)} disabled={busy} /></label>
    <label>新しいパスワード（変更する場合のみ）<input type="password" minLength={8} maxLength={72} autoComplete="new-password" value={password} onChange={e => setPassword(e.target.value)} disabled={busy} /></label>
    <p>パスワード変更後は全端末で再ログインが必要です。</p>
    <button disabled={busy} type="submit">{busy ? '保存中…' : '保存'}</button>
   </form>
  </>}
  <a href="/settings">設定へ戻る</a><a href="/login">ログイン画面へ</a>
 </section></main>
}
