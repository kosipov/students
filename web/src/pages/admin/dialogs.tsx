import { useId, useState, type FormEvent, type ReactNode } from 'react'
import { ApiError } from '../../api/client'
import {
  useCategories,
  useCreateGroup,
  useCreateSubject,
  useCreateTask,
  useUpdateGroup,
  useUpdateSubject,
  useUpdateTask,
} from '../../api/queries'
import type { AdminGroup, AdminSubject, AdminTask } from '../../api/types'
import { addCategories, CategoriesField } from '../../components/CategoriesField'
import { Modal } from '../../components/Modal'
import { TextField } from '../../components/TextField'
import { useToast } from '../../components/Toast'
import { generatePassword } from '../../lib/password'

interface FormDialogProps {
  title: string
  pending: boolean
  error: unknown
  onSubmit: () => void
  onClose: () => void
  children: ReactNode
}

function FormDialog({ title, pending, error, onSubmit, onClose, children }: FormDialogProps) {
  const submit = (event: FormEvent) => {
    event.preventDefault()
    onSubmit()
  }
  const message = error instanceof ApiError ? error.message : error ? 'Что-то пошло не так, попробуйте ещё раз' : ''

  return (
    <Modal title={title} onClose={onClose}>
      <form onSubmit={submit} noValidate>
        {message && (
          <p className="form-error" role="alert">
            {message}
          </p>
        )}
        {children}
        <div className="dialog-actions">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" className="btn btn-primary" disabled={pending}>
            {pending ? 'Сохраняем…' : 'Сохранить'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function fieldErrors(error: unknown): Record<string, string> {
  return error instanceof ApiError ? error.fields : {}
}

function PasswordField({ value, onChange, error }: { value: string; onChange: (value: string) => void; error?: string }) {
  const id = useId()
  return (
    <div className="field form-field">
      <label htmlFor={id}>Пароль для студентов</label>
      <div className="input-row">
        {/* Plain text on purpose: the admin needs to see the password to tell it to students. */}
        <input
          id={id}
          className="input"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete="off"
          spellCheck={false}
          maxLength={100}
          aria-invalid={error ? true : undefined}
          aria-describedby={`${id}-hint`}
        />
        <button type="button" className="btn btn-secondary" onClick={() => onChange(generatePassword())}>
          Сгенерировать
        </button>
      </div>
      <div id={`${id}-hint`} className="field-hint">
        Студенты введут его, чтобы открыть задание и ссылку на файл. Пустое поле — без пароля.
      </div>
      {error && <div className="field-error">{error}</div>}
    </div>
  )
}

function VisibleCheckbox({ checked, onChange }: { checked: boolean; onChange: (checked: boolean) => void }) {
  return (
    <label className="checkbox-row">
      <input type="checkbox" checked={checked} onChange={(event) => onChange(event.target.checked)} />
      Показывать студентам
    </label>
  )
}

export function GroupDialog({ group, onClose }: { group?: AdminGroup; onClose: () => void }) {
  const toast = useToast()
  const create = useCreateGroup()
  const update = useUpdateGroup()
  const mutation = group ? update : create
  const [name, setName] = useState(group?.name ?? '')
  const [visible, setVisible] = useState(group ? !group.hidden : true)

  const save = () => {
    const input = { name, hidden: !visible }
    const onSuccess = () => {
      toast('Сохранено')
      onClose()
    }
    if (group) update.mutate({ id: group.id, ...input }, { onSuccess })
    else create.mutate(input, { onSuccess })
  }

  return (
    <FormDialog
      title={group ? 'Изменить группу' : 'Новая группа'}
      pending={mutation.isPending}
      error={mutation.error}
      onSubmit={save}
      onClose={onClose}
    >
      <TextField label="Название" value={name} onChange={setName} error={fieldErrors(mutation.error).name} maxLength={255} />
      <VisibleCheckbox checked={visible} onChange={setVisible} />
    </FormDialog>
  )
}

export function SubjectDialog({ groupId, subject, onClose }: { groupId: number; subject?: AdminSubject; onClose: () => void }) {
  const toast = useToast()
  const create = useCreateSubject()
  const update = useUpdateSubject()
  const mutation = subject ? update : create
  const [name, setName] = useState(subject?.name ?? '')

  const save = () => {
    const onSuccess = () => {
      toast('Сохранено')
      onClose()
    }
    if (subject) update.mutate({ id: subject.id, name }, { onSuccess })
    else create.mutate({ groupId, name }, { onSuccess })
  }

  return (
    <FormDialog
      title={subject ? 'Изменить предмет' : 'Новый предмет'}
      pending={mutation.isPending}
      error={mutation.error}
      onSubmit={save}
      onClose={onClose}
    >
      <TextField label="Название" value={name} onChange={setName} error={fieldErrors(mutation.error).name} maxLength={255} />
    </FormDialog>
  )
}

export function TaskDialog({ subjectId, task, onClose }: { subjectId: number; task?: AdminTask; onClose: () => void }) {
  const toast = useToast()
  const create = useCreateTask()
  const update = useUpdateTask()
  const mutation = task ? update : create
  const [name, setName] = useState(task?.name ?? '')
  const [href, setHref] = useState(task?.href ?? '')
  const [comment, setComment] = useState(task?.comment ?? '')
  const [visible, setVisible] = useState(task ? !task.hidden : true)
  const [categories, setCategories] = useState<string[]>(task?.categories ?? [])
  const [categoryDraft, setCategoryDraft] = useState('')
  const [password, setPassword] = useState(task?.password ?? '')
  const suggestions = useCategories()
  const errors = fieldErrors(mutation.error)

  const save = () => {
    // A category typed without pressing Enter is saved too.
    const allCategories = addCategories(categories, categoryDraft)
    setCategories(allCategories)
    setCategoryDraft('')
    const input = { name, href, comment, hidden: !visible, categories: allCategories, password }
    const onSuccess = () => {
      toast('Сохранено')
      onClose()
    }
    if (task) update.mutate({ id: task.id, ...input }, { onSuccess })
    else create.mutate({ subjectId, ...input }, { onSuccess })
  }

  return (
    <FormDialog
      title={task ? 'Изменить задание' : 'Новое задание'}
      pending={mutation.isPending}
      error={mutation.error}
      onSubmit={save}
      onClose={onClose}
    >
      <TextField label="Название задания" value={name} onChange={setName} error={errors.name} maxLength={255} />
      <TextField
        label="Ссылка на файл"
        type="url"
        inputMode="url"
        placeholder="https://1drv.ms/…"
        value={href}
        onChange={setHref}
        error={errors.href}
        maxLength={255}
      />
      <TextField
        label="Короткое пояснение для студентов"
        value={comment}
        onChange={setComment}
        error={errors.comment}
        maxLength={255}
      />
      <CategoriesField
        label="Категории"
        value={categories}
        onChange={setCategories}
        draft={categoryDraft}
        onDraftChange={setCategoryDraft}
        suggestions={suggestions.data ?? []}
        error={errors.categories}
      />
      <PasswordField value={password} onChange={setPassword} error={errors.password} />
      <VisibleCheckbox checked={visible} onChange={setVisible} />
    </FormDialog>
  )
}
