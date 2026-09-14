import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from './client'

function mockFetch(response: Response) {
  const fetchMock = vi.fn().mockResolvedValue(response)
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('api client', () => {
  it('sends JSON with the CSRF header', async () => {
    const fetchMock = mockFetch(new Response(JSON.stringify({ id: 1 }), { status: 201 }))

    await expect(api.post('/admin/groups', { name: 'ИСП-301' })).resolves.toEqual({ id: 1 })

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toBe('/api/admin/groups')
    expect(init.method).toBe('POST')
    expect(init.body).toBe('{"name":"ИСП-301"}')
    expect(init.headers).toMatchObject({ 'X-Requested-With': 'fetch', 'Content-Type': 'application/json' })
  })

  it('returns nothing for 204', async () => {
    mockFetch(new Response(null, { status: 204 }))
    await expect(api.delete('/admin/groups/1')).resolves.toBeUndefined()
  })

  it('turns an error response into ApiError with field messages', async () => {
    mockFetch(new Response(JSON.stringify({ error: 'Проверьте введённые данные', fields: { name: 'Укажите название' } }), { status: 422 }))

    const error = await api.patch('/admin/groups/1', { name: '' }).catch((e: unknown) => e)
    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 422, message: 'Проверьте введённые данные', fields: { name: 'Укажите название' } })
  })

  it('reports a response without JSON', async () => {
    mockFetch(new Response('Bad Gateway', { status: 502 }))
    await expect(api.get('/catalog')).rejects.toMatchObject({ status: 502, message: 'Что-то пошло не так, попробуйте ещё раз' })
  })

  it('reports a network failure', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
    await expect(api.get('/catalog')).rejects.toMatchObject({ status: 0 })
  })
})
