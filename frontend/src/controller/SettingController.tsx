import { useRef, useState, type SubmitEvent } from 'react'
import SettingForm from '../ui/SettingForm'
import {
  applyBgmVolume,
  applySeVolume,
  getVolumes,
  playSe,
  saveVolumes,
  startBgm,
} from '../audio/audioManager.ts'

function SettingController() {
  const [bgmVolume, setBgmVolume] = useState(() => getVolumes().bgmVolume)
  const [seVolume, setSeVolume] = useState(() => getVolumes().seVolume)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')
  const previewTimer = useRef<number | undefined>(undefined)

  function handleBgmVolumeChange(value: number) {
    setBgmVolume(value)
    applyBgmVolume(value)
    startBgm() // スライダー操作は「ユーザー操作」なので、ここでBGMを鳴らせる
  }

  function handleSeVolumeChange(value: number) {
    setSeVolume(value)
    applySeVolume(value)
    // 動かし終わったあたりで、試しに効果音を1回鳴らす
    window.clearTimeout(previewTimer.current)
    previewTimer.current = window.setTimeout(playSe, 200)
  }

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsSubmitting(true)
    setErrorMessage('')
    try {
      // TODO: サーバーに保存するなら ../api の関数に置き換える
      saveVolumes()
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '設定を保存できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  function handleAccount() {
    // TODO: アカウント画面ができたら移動先を設定する(今は /account が存在しない)
    window.location.assign('/account')
  }

  async function handleLogout() {
    try {
      await fetch('/api/logout', {
        method: 'POST',
        headers: { 'X-Requested-With': 'gomoku-web' },
      })
    } finally {
      window.location.assign('/login')
    }
  }

  function handleBack() {
    window.location.assign('/home')
  }

  return (
    <SettingForm
      bgmVolume={bgmVolume}
      seVolume={seVolume}
      onBgmVolumeChange={handleBgmVolumeChange}
      onSeVolumeChange={handleSeVolumeChange}
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