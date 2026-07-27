import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App'
import { AuthProvider } from './auth/AuthContext'
import { ToastProvider } from './components/toast/ToastContext'
import { createLogger, currentLevel } from './lib/logger'

const log = createLogger('app')

// First line of any session: which API the browser will actually talk to. A wrong
// VITE_API_BASE_URL makes every request fail, and this is what says so.
log.info('SoundFlow booting', {
  api_base_url: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:3000/api/v1 (default)',
  log_level: currentLevel(),
})

const root = document.getElementById('root')
if (!root) throw new Error('Root element #root not found')

createRoot(root).render(
  <StrictMode>
    <ToastProvider>
      <AuthProvider>
        <App />
      </AuthProvider>
    </ToastProvider>
  </StrictMode>,
)
