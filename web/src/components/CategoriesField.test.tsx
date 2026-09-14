import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { addCategories, CategoriesField } from './CategoriesField'

function Harness({ initial = [] as string[] }) {
  const [value, setValue] = useState(initial)
  const [draft, setDraft] = useState('')
  return (
    <>
      <CategoriesField label="Категории" value={value} onChange={setValue} draft={draft} onDraftChange={setDraft} suggestions={['Курсовые', 'Методички']} />
      <output data-testid="value">{value.join('|')}</output>
    </>
  )
}

describe('addCategories', () => {
  it('splits by comma, trims and skips repeats in any case', () => {
    expect(addCategories(['Курсовые'], ' курсовые ,  Методички   к  лабам, ,')).toEqual(['Курсовые', 'Методички к лабам'])
  })
})

describe('CategoriesField', () => {
  it('adds a category on Enter and on comma', async () => {
    render(<Harness />)
    const input = screen.getByLabelText('Категории')

    await userEvent.type(input, 'Курсовые{Enter}')
    await userEvent.type(input, 'Методички,')

    expect(screen.getByTestId('value')).toHaveTextContent('Курсовые|Методички')
    expect(input).toHaveValue('')
  })

  it('removes a category with its button and with Backspace in the empty field', async () => {
    render(<Harness initial={['Курсовые', 'Методички', 'Лекции']} />)

    await userEvent.click(screen.getByRole('button', { name: 'Убрать категорию «Методички»' }))
    expect(screen.getByTestId('value')).toHaveTextContent('Курсовые|Лекции')
    expect(screen.getByLabelText('Категории')).toHaveFocus()

    await userEvent.keyboard('{Backspace}')
    expect(screen.getByTestId('value')).toHaveTextContent('Курсовые')
  })

  it('suggests only categories that are not chosen yet', () => {
    const { container } = render(<Harness initial={['курсовые']} />)
    const options = Array.from(container.querySelectorAll('datalist option')).map((o) => o.getAttribute('value'))
    expect(options).toEqual(['Методички'])
  })
})
