import { useId, useRef, type KeyboardEvent } from 'react'
import { categoryKey } from '../lib/catalog'

export const MAX_CATEGORIES = 10
export const MAX_CATEGORY_LENGTH = 50

interface CategoriesFieldProps {
  label: string
  value: string[]
  onChange: (value: string[]) => void
  /** The text typed but not yet added; the form adds it on save. */
  draft: string
  onDraftChange: (draft: string) => void
  suggestions: string[]
  error?: string
}

/** Adds names to the list, skipping empty ones and repeats in any letter case. */
export function addCategories(current: string[], text: string): string[] {
  const result = [...current]
  for (const part of text.split(',')) {
    const name = part.trim().replace(/\s+/g, ' ')
    if (name && !result.some((existing) => categoryKey(existing) === categoryKey(name))) {
      result.push(name)
    }
  }
  return result
}

/** A list of free-form labels: type a name and press Enter or a comma to add it. */
export function CategoriesField({ label, value, onChange, draft, onDraftChange, suggestions, error }: CategoriesFieldProps) {
  const id = useId()
  const hintId = `${id}-hint`
  const errorId = `${id}-error`
  const listId = `${id}-suggestions`
  const inputRef = useRef<HTMLInputElement>(null)

  const commit = (text: string) => {
    onChange(addCategories(value, text))
    onDraftChange('')
  }

  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === ',') {
      if (draft.trim()) {
        event.preventDefault()
        commit(draft)
      } else if (event.key === ',') {
        event.preventDefault()
      }
      // Enter on an empty field submits the form.
    } else if (event.key === 'Backspace' && draft === '' && value.length > 0) {
      onChange(value.slice(0, -1))
    }
  }

  const remove = (name: string) => {
    onChange(value.filter((category) => category !== name))
    inputRef.current?.focus()
  }

  const available = suggestions.filter((name) => !value.some((category) => categoryKey(category) === categoryKey(name)))

  return (
    <div className="field form-field">
      <label htmlFor={id}>{label}</label>
      {value.length > 0 && (
        <ul className="chips" aria-label="Выбранные категории">
          {value.map((name) => (
            <li key={name} className="chip">
              {name}
              <button type="button" className="chip-remove" onClick={() => remove(name)} aria-label={`Убрать категорию «${name}»`}>
                ×
              </button>
            </li>
          ))}
        </ul>
      )}
      <input
        ref={inputRef}
        id={id}
        className="input"
        list={listId}
        value={draft}
        placeholder={value.length >= MAX_CATEGORIES ? 'Больше категорий добавить нельзя' : 'Например, Курсовые'}
        disabled={value.length >= MAX_CATEGORIES}
        maxLength={MAX_CATEGORY_LENGTH}
        onChange={(event) => {
          const text = event.target.value
          // A pasted "Курсовые, Методички" or a picked suggestion ending with a comma is added at once.
          if (text.includes(',')) commit(text)
          else onDraftChange(text)
        }}
        onKeyDown={onKeyDown}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${hintId} ${errorId}` : hintId}
      />
      <datalist id={listId}>
        {available.map((name) => (
          <option key={name} value={name} />
        ))}
      </datalist>
      <div id={hintId} className="field-hint">
        Enter или запятая добавляют категорию. По категориям студенты фильтруют задания предмета.
      </div>
      {error && (
        <div id={errorId} className="field-error">
          {error}
        </div>
      )}
    </div>
  )
}
