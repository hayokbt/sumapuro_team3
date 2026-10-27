import AccountCreateController from '../controller/AccountCreateController'
import AccountRecoveryController from '../controller/AccountRecoveryController'
import AccountResetController from '../controller/AccountResetController'
import LoginController from '../controller/LoginController'
import HomeController from '../controller/HomeController'

function AppRouter() {
  switch (window.location.pathname) {
    case '/home':
      return <HomeController />
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
