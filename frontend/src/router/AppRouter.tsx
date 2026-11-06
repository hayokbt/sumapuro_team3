import { useEffect } from 'react'
import AccountCreateController from '../controller/AccountCreateController'
import AccountRecoveryController from '../controller/AccountRecoveryController'
import AccountResetController from '../controller/AccountResetController'
import HomeController from '../controller/HomeController'
import LoginController from '../controller/LoginController'
import GameForm from '../ui/GameForm'
import SettingController from '../controller/SettingController'
import { startBgm } from '../audio/audioManager'

// 設定ボタンを出さない画面(ログイン・アカウント作成と、設定画面そのもの)
const SETTINGS_BUTTON_HIDDEN_PATHS = ['/', '/login', '/account/create', '/settings']

function renderPage(pathname: string) {
  switch (pathname) {
    case '/home':
      return <HomeController />
    case '/game':
      return <GameForm />
    case '/':
    case '/login':
      return <LoginController />
    case '/account/create':
      return <AccountCreateController />
    case '/account/recovery':
      return <AccountRecoveryController />
    case '/account/reset':
      return <AccountResetController />
    case '/settings':
      return <SettingController />
    default:
      return <p>ページが見つかりません。</p>
  }
}

function AppRouter() {
  const pathname = window.location.pathname

  useEffect(() => {
    startBgm()
  }, [])

  const showSettingsButton = !SETTINGS_BUTTON_HIDDEN_PATHS.includes(pathname)

  return (
    <>
      {renderPage(pathname)}
      {showSettingsButton && (
        <a href="/settings" className="settings-link" aria-label="設定">
          ⚙
        </a>
      )}
    </>
  )
}

export default AppRouter
