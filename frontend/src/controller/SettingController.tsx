import { useEffect, useRef, useState, type SubmitEvent } from 'react'
import { getSettings, saveSettings } from '../api/account'
import { logout } from '../api/authentication'
import SettingForm from '../ui/SettingForm'
import {
  applyBgmVolume,
  applySeVolume,
  getVolumes,
  playSe,
  saveVolumes,
} from '../audio/audioManager.ts'

function SettingController() {
  const [bgmVolume, setBgmVolume] = useState(() => getVolumes().bgmVolume)
  const [seVolume, setSeVolume] = useState(() => getVolumes().seVolume)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')
  const lastPreviewTime = useRef(-Infinity)
  const [loaded, setLoaded] = useState(false)
  useEffect(() => {
    const controller = new AbortController()
    getSettings(controller.signal).then(settings => {
      setBgmVolume(settings.bgmVolume); setSeVolume(settings.seVolume)
      applyBgmVolume(settings.bgmVolume); applySeVolume(settings.seVolume); saveVolumes()
      setLoaded(true)
    }).catch(error => {
      if (!controller.signal.aborted) setErrorMessage(error instanceof Error ? error.message : '設定を取得できませんでした。')
    })
    return () => controller.abort()
  }, [])

  function handleBgmVolumeChange(value: number) {
    setBgmVolume(value)
    applyBgmVolume(value)
  }

  function handleSeVolumeChange(value: number) {
    setSeVolume(value)
    applySeVolume(value)
    // Preview throughout dragging, with a short interval to avoid overlapping clicks.
    const now = performance.now()
    if (now - lastPreviewTime.current >= 80 || value === 0 || value === 100) {
      lastPreviewTime.current = now
      playSe()
    }
  }

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!loaded || isSubmitting) return
    setIsSubmitting(true)
    setErrorMessage('')
    try {
      await saveSettings({ bgmVolume, seVolume })
      saveVolumes()
      window.location.assign('/home') // 保存できたら、ホームへ戻る
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '設定を保存できませんでした。')
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleLogout() {
    try {
      await logout()
      window.location.assign('/login')
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'ログアウトできませんでした。')
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
      onLogout={handleLogout}
      onBack={handleBack}
      onSubmit={handleSubmit}
      isLoading={!loaded}
      isSubmitting={isSubmitting}
      errorMessage={errorMessage}
    />
  )
}

export default SettingController
