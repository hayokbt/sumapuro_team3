import AccountCreateController from '../controller/AccountCreateController'
import AccountRecoveryController from '../controller/AccountRecoveryController'
import AccountResetController from '../controller/AccountResetController'
import HomeController from '../controller/HomeController'
import LoginController from '../controller/LoginController'
import GameForm from '../ui/GameForm'
import SettingController from '../controller/SettingController'

function AppRouter() {
  switch (window.location.pathname) {
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

export default AppRouter
