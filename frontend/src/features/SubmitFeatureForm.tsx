import { useEffect, useRef, useState, type FormEvent } from 'react'
import { api, ApiError } from '../api'
import { useAuth } from '../auth/useAuth'
import { useToast } from '../components/toast/useToast'
import { Button, TextArea, TextField } from '../components/ui'

const TITLE_MAX = 100
const DESC_MAX = 200

interface SubmitFeatureFormProps {
  open: boolean
  onClose: () => void
  onCreated: () => void
}

export function SubmitFeatureForm({ open, onClose, onCreated }: SubmitFeatureFormProps) {
  const { token } = useAuth()
  const toast = useToast()
  const titleRef = useRef<HTMLInputElement>(null)

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) titleRef.current?.focus()
  }, [open])

  if (!open) return null

  function reset() {
    setTitle('')
    setDescription('')
    setFieldErrors({})
  }

  function clientValidate(): Record<string, string> {
    const errs: Record<string, string> = {}
    const t = title.trim()
    const d = description.trim()
    if (t.length < 2 || t.length > TITLE_MAX) errs.title = 'Title must be 2–100 characters.'
    if (d.length < 2 || d.length > DESC_MAX) errs.description = 'Description must be 2–200 characters.'
    return errs
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!token) return

    const errs = clientValidate()
    if (Object.keys(errs).length > 0) {
      setFieldErrors(errs)
      return
    }

    setSubmitting(true)
    setFieldErrors({})
    try {
      await api.createFeature({ title: title.trim(), description: description.trim() }, token)
      toast.success('Request submitted')
      reset()
      onClose()
      onCreated()
    } catch (err) {
      if (err instanceof ApiError && err.code === 'DUPLICATE_FEATURE') {
        setFieldErrors({ title: 'A request with this title already exists.' })
      } else if (err instanceof ApiError && err.details.length > 0) {
        const mapped: Record<string, string> = {}
        for (const fe of err.details) mapped[fe.field] = fe.issue
        setFieldErrors(mapped)
      } else if (err instanceof ApiError) {
        toast.error(err.message)
      } else {
        toast.error('Could not submit your request. Please try again.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form
      onSubmit={onSubmit}
      noValidate
      className="animate-rise flex flex-col gap-4 rounded-2xl border border-accent/30 bg-surface/80 p-5 shadow-[0_0_40px_-20px_rgba(6,182,212,0.5)]"
    >
      <div className="flex items-center justify-between">
        <h2 className="font-semibold text-text">New feature request</h2>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close form"
          className="rounded-lg px-2 py-1 text-sm text-muted hover:text-text"
        >
          Cancel
        </button>
      </div>

      <TextField
        ref={titleRef}
        id="feature-title"
        label="Title"
        value={title}
        maxLength={TITLE_MAX}
        onChange={(e) => setTitle(e.target.value)}
        error={fieldErrors.title}
        hint={fieldErrors.title ? undefined : `${title.length}/${TITLE_MAX}`}
        placeholder="Offline downloads"
      />
      <TextArea
        id="feature-description"
        label="Description"
        rows={3}
        value={description}
        maxLength={DESC_MAX}
        onChange={(e) => setDescription(e.target.value)}
        error={fieldErrors.description}
        hint={fieldErrors.description ? undefined : `${description.length}/${DESC_MAX}`}
        placeholder="Download playlists to listen offline on flights and trips."
      />

      <div className="flex justify-end gap-2">
        <Button type="button" variant="ghost" onClick={onClose}>
          Cancel
        </Button>
        <Button type="submit" loading={submitting}>
          Create Request
        </Button>
      </div>
    </form>
  )
}
