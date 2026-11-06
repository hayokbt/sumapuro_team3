import { useState, type SubmitEvent } from 'react'
import SettingForm from '../ui/SettingForm'

const STORAGE_KEY = 'settings'

function loadSettings() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
    return {
      bgmVolume: typeof saved.bgmVolume === 'number' ? saved.bgmVolume : 30,
      seVolume: typeof saved.seVolume === 'number' ? saved.seVolume : 60,
    }
  } catch {
    return { bgmVolume: 30, seVolume: 60 }
  }
}

function SettingController() {
  const [bgmVolume, setBgmVolume] = useState(() => loadSettings().bgmVolume)
  const [seVolume, setSeVolume] = useState(() => loadSettings().seVolume)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsSubmitting(true)
    setErrorMessage('')
    try {
      // TODO: サーバーに保存するなら ../api の関数に置き換える
      // 例: await requestSaveSettings({ bgmVolume, seVolume })
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ bgmVolume, seVolume }))
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '設定を保存できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  function handleAccount() {
    window.location.assign('/account')
  }

  function handleLogout() {
    // TODO: ログアウトAPIを作ったらここで呼ぶ
    window.location.assign('/login')
  }

  function handleBack() {
    window.location.assign('/home')
  }

  return (
    <SettingForm
      bgmVolume={bgmVolume}
      seVolume={seVolume}
      onBgmVolumeChange={setBgmVolume}
      onSeVolumeChange={setSeVolume}
      onAccount={handleAccount}
      onLogout={handleLogout}
      onBack={handleBack}
      onSubmit={handleSubmit}
      isSubmitting={isSubmitting}
      errorMessage={errorMessage}
    />
  )
}

export default SettingController