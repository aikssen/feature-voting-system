import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '../components/Modal'
import { Logo } from '../components/Logo'
import { Button, TextField } from '../components/ui'
import { ApiError } from '../api'
import { useAuth } from './useAuth'

type Fields = Record<string, string>

export function AuthModal() {
  const { authOpen, authMode, closeAuth, openAuth, login, signup } = useAuth()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Fields>({})
  const [formError, setFormError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const isSignup = authMode === 'signup'

  // Reset transient state whenever the dialog opens or flips mode.
  useEffect(() => {
    if (authOpen) {
      setFieldErrors({})
      setFormError('')
    }
  }, [authOpen, authMode])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setFieldErrors({})
    setFormError('')
    try {
      if (isSignup) {
        await signup(name, email, password)
      } else {
        await login(email, password)
      }
      // Success closes the modal; clear the form for next time.
      setName('')
      setEmail('')
      setPassword('')
    } catch (err) {
      if (err instanceof ApiError && err.details.length > 0) {
        const mapped: Fields = {}
        for (const d of err.details) mapped[d.field] = d.issue
        setFieldErrors(mapped)
        setFormError('Please fix the highlighted fields.')
      } else if (err instanceof ApiError) {
        setFormError(err.message)
      } else {
        setFormError('Something went wrong. Please try again.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal open={authOpen} onClose={closeAuth} labelledBy="auth-title">
      <div className="mb-5 flex flex-col gap-3">
        <Logo />
        <div>
          <h2 id="auth-title" className="text-xl font-bold tracking-tight">
            {isSignup ? 'Create your account' : 'Welcome back'}
          </h2>
          <p className="mt-1 text-sm text-muted">
            {isSignup ? 'Join the community shaping SoundFlow.' : 'Log in to submit ideas and vote.'}
          </p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
        {isSignup && (
          <TextField
            id="name"
            label="Name"
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            error={fieldErrors.name}
            placeholder="Ada Lovelace"
          />
        )}
        <TextField
          id="email"
          label="Email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={fieldErrors.email}
          placeholder="you@example.com"
        />
        <TextField
          id="password"
          label="Password"
          type="password"
          autoComplete={isSignup ? 'new-password' : 'current-password'}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={fieldErrors.password}
          hint={isSignup ? '4–12 characters, at least 1 special character.' : undefined}
          placeholder="••••••"
        />

        {formError && (
          <p role="alert" className="rounded-lg border border-danger/40 bg-danger/10 px-3 py-2 text-sm text-danger">
            {formError}
          </p>
        )}

        <Button type="submit" loading={submitting} className="w-full">
          {isSignup ? 'Create account' : 'Log in'}
        </Button>
      </form>

      <p className="mt-5 text-center text-sm text-muted">
        {isSignup ? 'Already have an account?' : 'New to SoundFlow?'}{' '}
        <button
          type="button"
          className="font-semibold text-accent hover:underline"
          onClick={() => openAuth(isSignup ? 'login' : 'signup')}
        >
          {isSignup ? 'Log in' : 'Create one'}
        </button>
      </p>
    </Modal>
  )
}
