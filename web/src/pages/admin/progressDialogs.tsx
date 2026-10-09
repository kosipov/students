import { useState, type FormEvent } from 'react'
import { ApiError } from '../../api/client'
import { useAddStudent, useImportStudents, useRenameStudent } from '../../api/queries'
import type { ProgressStudent } from '../../api/types'
import { Modal } from '../../components/Modal'
import { TextField } from '../../components/TextField'
import { useToast } from '../../components/Toast'

function fieldError(error: unknown, field: string): string | undefined {
  return error instanceof ApiError ? error.fields[field] : undefined
}

/** Adds one student or renames an existing one. */
export function StudentDialog({
  groupId,
  student,
  onClose,
}: {
  groupId: number
  student?: ProgressStudent
  onClose: () => void
}) {
  const toast = useToast()
  const add = useAddStudent(groupId)
  const rename = useRenameStudent(groupId)
  const mutation = student ? rename : add
  const [name, setName] = useState(student?.name ?? '')

  const save = (event: FormEvent) => {
    event.preventDefault()
    const onSuccess = () => {
      toast('Сохранено')
      onClose()
    }
    if (student) rename.mutate({ id: student.id, name }, { onSuccess })
    else add.mutate(name, { onSuccess })
  }

  return (
    <Modal title={student ? 'Переименовать студента' : 'Новый студент'} onClose={onClose}>
      <form onSubmit={save} noValidate>
        {mutation.isError && (
          <p className="form-error" role="alert">
            {mutation.error instanceof ApiError ? mutation.error.message : 'Что-то пошло не так, попробуйте ещё раз'}
          </p>
        )}
        <TextField
          label="Фамилия и имя"
          value={name}
          onChange={setName}
          error={fieldError(mutation.error, 'name')}
          maxLength={120}
        />
        <div className="dialog-actions">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
            {mutation.isPending ? 'Сохраняем…' : 'Сохранить'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

/** Imports the roster pasted as one name per line. */
export function ImportStudentsDialog({ groupId, onClose }: { groupId: number; onClose: () => void }) {
  const toast = useToast()
  const importStudents = useImportStudents(groupId)
  const [names, setNames] = useState('')

  const save = (event: FormEvent) => {
    event.preventDefault()
    importStudents.mutate(names, {
      onSuccess: (result) => {
        toast(
          result.skipped > 0
            ? `Добавлено: ${result.added}, уже были в списке: ${result.skipped}`
            : `Добавлено студентов: ${result.added}`,
        )
        onClose()
      },
    })
  }

  return (
    <Modal title="Импорт списка группы" onClose={onClose}>
      <form onSubmit={save} noValidate>
        {importStudents.isError && (
          <p className="form-error" role="alert">
            {importStudents.error instanceof ApiError
              ? importStudents.error.message
              : 'Что-то пошло не так, попробуйте ещё раз'}
          </p>
        )}
        <div className="field form-field">
          <label htmlFor="import-names">Список студентов</label>
          <textarea
            id="import-names"
            className="input"
            rows={10}
            placeholder={'Иванов Иван\nПетрова Мария\nСидоров Алексей'}
            value={names}
            onChange={(event) => setNames(event.target.value)}
            aria-invalid={fieldError(importStudents.error, 'names') ? true : undefined}
          />
          <div className="field-hint">По одному студенту на строку. Те, кто уже есть в списке, пропускаются.</div>
          {fieldError(importStudents.error, 'names') && (
            <div className="field-error">{fieldError(importStudents.error, 'names')}</div>
          )}
        </div>
        <div className="dialog-actions">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" className="btn btn-primary" disabled={importStudents.isPending}>
            {importStudents.isPending ? 'Добавляем…' : 'Добавить'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
