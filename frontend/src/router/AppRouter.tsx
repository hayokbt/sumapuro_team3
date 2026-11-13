import { Fragment, useEffect, useSyncExternalStore } from 'react'
import AccountController from '../controller/AccountController'
import AccountCreateController from '../controller/AccountCreateController'
import AccountRecoveryController from '../controller/AccountRecoveryController'
import AccountResetController from '../controller/AccountResetController'
import HomeController from '../controller/HomeController'
import LoginController from '../controller/LoginController'
import RankingController from '../controller/RankingController'
import GameForm from '../ui/GameForm'
import SettingController from '../controller/SettingController'
import { installAudio } from '../audio/audioManager'
import { getRoute, installNavigation, subscribeToNavigation } from './navigation'

function renderPage(pathname: string) {
  switch (pathname) {
    case '/home':
      return <HomeController />
    case '/ranking':
      return <RankingController />
    case '/game':
      return <GameForm />
    case '/':
    case '/login':
      return <LoginController />
    case '/account':
      return <AccountController />
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
  const route = useSyncExternalStore(subscribeToNavigation, getRoute)
  const pathname = route.split('?')[0]

  useEffect(() => installNavigation(), [])

  useEffect(() => {
    return installAudio()
  }, [])

  // 対局中のページ遷移で接続を切らないよう、設定への入口はホームに限定する。
  const showSettingsButton = pathname === '/home'

  return (
    <>
      <Fragment key={route}>{renderPage(pathname)}</Fragment>
      {showSettingsButton && (
        <a href="/settings" className="settings-link" aria-label="設定">
          ⚙
        </a>
      )}
    </>
  )
}

export default AppRouter
