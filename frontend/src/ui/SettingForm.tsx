import type { SubmitEvent } from 'react'

type SettingFormProps = {
  bgmVolume: number
  seVolume: number
  onBgmVolumeChange: (value: number) => void
  onSeVolumeChange: (value: number) => void
  onAccount: () => void
  onLogout: () => void
  onBack: () => void
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
  isSubmitting: boolean
  errorMessage: string
}

function SettingForm({
  bgmVolume,
  seVolume,
  onBgmVolumeChange,
  onSeVolumeChange,
  onAccount,
  onLogout,
  onBack,
  onSubmit,
  isSubmitting,
  errorMessage,
}: SettingFormProps) {
  return (
    <main className="settings-page">
      <div className="form-container settings-form">
        <form onSubmit={onSubmit}>
          <div className="settings-row settings-top">
            <button type="button" onClick={onAccount} disabled={isSubmitting}>
              アカウント
            </button>
            <button type="button" onClick={onLogout} disabled={isSubmitting}>
              ログアウト
            </button>
          </div>

          <div className="settings-row">
            <label htmlFor="bgmVolume" className="settings-label">
              BGM
            </label>
            <input
              id="bgmVolume"
              type="range"
              name="bgmVolume"
              min={0}
              max={100}
              value={bgmVolume}
              onChange={(event) => onBgmVolumeChange(Number(event.target.value))}
              disabled={isSubmitting}
            />
          </div>

          <div className="settings-row">
            <label htmlFor="seVolume" className="settings-label">
              効果音
            </label>
            <input
              id="seVolume"
              type="range"
              name="seVolume"
              min={0}
              max={100}
              value={seVolume}
              onChange={(event) => onSeVolumeChange(Number(event.target.value))}
              disabled={isSubmitting}
            />
          </div>

          {errorMessage && <p role="alert">{errorMessage}</p>}

          <div className="settings-row settings-bottom">
            <button
              type="button"
              className="settings-back"
              aria-label="戻る"
              onClick={onBack}
              disabled={isSubmitting}
            >
              ↩
            </button>
            <button type="submit" className="settings-save" disabled={isSubmitting}>
              {isSubmitting ? '保存中…' : '保存'}
            </button>
          </div>
        </form>
      </div>
    </main>
  )
}

export default SettingForm