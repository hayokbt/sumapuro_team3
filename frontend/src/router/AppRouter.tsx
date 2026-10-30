import AccountCreateController from '../controller/AccountCreateController'
import AccountRecoveryController from '../controller/AccountRecoveryController'
import AccountResetController from '../controller/AccountResetController'
import LoginController from '../controller/LoginController'
import GameForm from '../ui/GameForm'
import HomeForm from '../ui/HomeForm'

function AppRouter() {
  switch (window.location.pathname) {
    case '/home':
      return <HomeForm />
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
    default:
      return <p>ページが見つかりません。</p>
  }
}

export default AppRouter
