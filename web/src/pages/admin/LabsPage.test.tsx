import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { mskToday } from '../../lib/labs'
import { mockApi, renderWithProviders } from '../../test/render'
import { LabsPage } from './LabsPage'

afterEach(() => vi.unstubAllGlobals())

function renderLabs() {
  return renderWithProviders(
    <Routes>
      <Route path="/admin/labs" element={<LabsPage />} />
    </Routes>,
    { path: '/admin/labs' },
  )
}

const today = mskToday()

const busyRoom = {
  room: '502/1',
  floor: 5,
  busy: true,
  lesson: {
    number: 3,
    title: 'Базы данных и информационные системы',
    kind: 'лаб.',
    teacher: 'Шадрин Лев Сергеевич',
    groups: '141-ПМо',
    startsAt: '2026-10-16T12:40:00+03:00',
    endsAt: '2026-10-16T14:10:00+03:00',
  },
  fetchedAt: '2026-10-16T11:00:00+03:00',
  stale: false,
  failed: false,
}

const freeRoom = { room: '503/1', floor: 5, busy: false, lesson: null, fetchedAt: '2026-10-16T11:00:00+03:00', stale: false, failed: false }

const occupancy = {
  date: '2026-10-16T00:00:00+03:00',
  pair: 3,
  startsAt: '2026-10-16T12:40:00+03:00',
  endsAt: '2026-10-16T14:10:00+03:00',
  fetchedAt: '2026-10-16T11:00:00+03:00',
  rooms: [busyRoom, freeRoom],
}

describe('LabsPage', () => {
  it('asks only after both the floor and the pair are picked', async () => {
    const fetchMock = mockApi({
      [`GET /api/admin/labs?floor=5&date=${today}&pair=now`]: () => ({ status: 200, body: occupancy }),
    })
    renderLabs()

    await userEvent.click(screen.getByRole('button', { name: '5 этаж' }))
    expect(fetchMock).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: 'Сейчас' }))
    expect(await screen.findByText('Шадрин Лев Сергеевич · 141-ПМо')).toBeInTheDocument()
    expect(screen.getByText('Занята')).toBeInTheDocument()
    expect(screen.getByText('Свободна')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(`/api/admin/labs?floor=5&date=${today}&pair=now`, expect.anything())
  })

  it('loads another pair of the same floor', async () => {
    const fetchMock = mockApi({
      [`GET /api/admin/labs?floor=2&date=${today}&pair=4`]: () => ({ status: 200, body: { ...occupancy, pair: 4, rooms: [] } }),
    })
    renderLabs()

    await userEvent.click(screen.getByRole('button', { name: '2 этаж' }))
    await userEvent.click(screen.getByRole('button', { name: '4' }))

    expect(await screen.findByText(/4-я пара/)).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(`/api/admin/labs?floor=2&date=${today}&pair=4`, expect.anything())
  })

  it('refreshes one room from the campus and keeps the rest', async () => {
    const updated = { ...freeRoom, busy: true, lesson: busyRoom.lesson }
    mockApi({
      [`GET /api/admin/labs?floor=5&date=${today}&pair=now`]: () => ({ status: 200, body: occupancy }),
      'POST /api/admin/labs/refresh': (init) => {
        expect(JSON.parse(String(init.body))).toEqual({ room: '503/1', date: today, pair: 3 })
        return { status: 200, body: updated }
      },
    })
    renderLabs()

    await userEvent.click(screen.getByRole('button', { name: '5 этаж' }))
    await userEvent.click(screen.getByRole('button', { name: 'Сейчас' }))
    const buttons = await screen.findAllByRole('button', { name: 'Запросить с кампуса' })
    expect(buttons).toHaveLength(2)

    await userEvent.click(buttons[1]!)

    // The refreshed room is busy now; the other one is untouched.
    expect(await screen.findAllByText('Занята')).toHaveLength(2)
  })

  it('shows the server error and lets retry', async () => {
    let attempts = 0
    mockApi({
      [`GET /api/admin/labs?floor=all&date=${today}&pair=1`]: () => {
        attempts++
        return attempts === 1 ? { status: 500, body: { error: 'Что-то пошло не так' } } : { status: 200, body: occupancy }
      },
    })
    renderLabs()

    await userEvent.click(screen.getByRole('button', { name: 'Все' }))
    await userEvent.click(screen.getByRole('button', { name: '1' }))

    expect(await screen.findByText('Что-то пошло не так')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: /ещё раз/i }))
    expect(await screen.findByText('Шадрин Лев Сергеевич · 141-ПМо')).toBeInTheDocument()
  })
})
