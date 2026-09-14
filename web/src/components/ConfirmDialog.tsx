import { useState } from 'react'
import { ApiError } from '../api/client'
import { Modal } from './Modal'

interface ConfirmDialogProps {
  title: string
  message: string
  confirmLabel: string
  onConfirm: () => Promise<unknown>
  onClose: () => void
}

export function ConfirmDialog({ title, message, confirmLabel, onConfirm, onClose }: ConfirmDialogProps) {
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  const confirm = async () => {
    setPending(true)
    setError('')
    try {
      await onConfirm()
      onClose()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Что-то пошло не так, попробуйте ещё раз')
      setPending(false)
    }
  }

  return (
    <Modal title={title} onClose={onClose}>
      <p className="dialog-body">{message}</p>
      {error && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}
      <div className="dialog-actions">
        <button type="button" className="btn btn-secondary" onClick={onClose}>
          Отмена
        </button>
        <button type="button" className="btn btn-primary" onClick={confirm} disabled={pending}>
          {pending ? 'Удаляем…' : confirmLabel}
        </button>
      </div>
    </Modal>
  )
}
