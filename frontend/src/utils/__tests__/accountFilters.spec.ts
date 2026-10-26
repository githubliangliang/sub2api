import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { loadAccountFilters, saveAccountFilters } from '../accountFilters'

describe('account platform and status preferences', () => {
  beforeEach(() => localStorage.clear())
  afterEach(() => vi.restoreAllMocks())

  it('restores only supported fields and persists clearing', () => {
    saveAccountFilters({ platform: 'openai', status: 'error' })
    expect(loadAccountFilters()).toEqual({ platform: 'openai', status: 'error' })
    expect(JSON.parse(localStorage.getItem('account-filter-values')!)).toEqual({ platform: 'openai', status: 'error' })
    saveAccountFilters({ platform: '', status: 'error' })
    expect(loadAccountFilters()).toEqual({ platform: '', status: 'error' })
  })

  it.each(['null', '[]', '{', 'false', '123'])('ignores malformed structures %s', (value) => {
    localStorage.setItem('account-filter-values', value)
    expect(loadAccountFilters()).toEqual({ platform: '', status: '' })
  })

  it('validates fields independently against local supported values', () => {
    localStorage.setItem('account-filter-values', JSON.stringify({ platform: 'cline', status: 'active', search: 'private' }))
    expect(loadAccountFilters()).toEqual({ platform: '', status: 'active' })
    localStorage.setItem('account-filter-values', JSON.stringify({ platform: 'grok', status: 7 }))
    expect(loadAccountFilters()).toEqual({ platform: 'grok', status: '' })
  })

  it('tolerates storage read and write failures', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('disabled') })
    expect(loadAccountFilters()).toEqual({ platform: '', status: '' })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('full') })
    expect(() => saveAccountFilters({ platform: 'openai', status: 'error' })).not.toThrow()
  })
})
