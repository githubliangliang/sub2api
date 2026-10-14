import { describe, expect, it } from 'vitest'
import { extractApiErrorMessage } from '../apiError'

describe('API error display text', () => {
  it.each([
    [{ message: 'Request failed with status code 403', response: { data: { detail: 'Access denied' } } }, 'Access denied'],
    [{ message: 'Request failed with status code 403' }, 'fallback'],
    [{ message: { secret: 'value' }, error: { message: 'nested' } }, 'fallback'],
    [{ response: { data: { detail: { reason: 'structured' }, message: 'Readable reason' } } }, 'Readable reason'],
    [{ status: 400, message: 'Normalized reason' }, 'Normalized reason'],
    [new Error('Network unavailable'), 'Network unavailable'],
    [{ message: '  ' }, 'fallback'],
  ])('extracts safe text from %j', (error, expected) => {
    expect(extractApiErrorMessage(error, 'fallback')).toBe(expected)
  })

  it('retains explicit error-code translations', () => {
    expect(extractApiErrorMessage({ code: 'DENIED', message: 'Raw reason' }, 'fallback', { DENIED: 'Translated' })).toBe('Translated')
  })
})
