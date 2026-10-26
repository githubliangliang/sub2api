import { describe, expect, it } from 'vitest'
import { formatUsageOutputRate } from '../latencyHealth'

const textRow = { output_tokens: 100, duration_ms: 2000, image_count: 0, image_output_tokens: 0, billing_mode: 'token', request_type: 'stream' } as const

describe('usage output TPS', () => {
  it('uses total duration, supports all text transports and rounds tiny positive rates', () => {
    expect(formatUsageOutputRate(textRow)).toBe('50.0 tok/s')
    for (const request_type of ['sync', 'stream', 'ws_v2', 'cyber'] as const) {
      expect(formatUsageOutputRate({ ...textRow, request_type })).toBe('50.0 tok/s')
    }
    expect(formatUsageOutputRate({ ...textRow, output_tokens: 1, duration_ms: 100000 })).toBe('0.0 tok/s')
    expect(formatUsageOutputRate({ ...textRow, request_type: undefined })).toBe('50.0 tok/s')
  })

  it.each([0, -1, NaN, Infinity, undefined, null])('does not display invalid output or duration: %s', (value) => {
    expect(formatUsageOutputRate({ ...textRow, output_tokens: value as number })).toBe('—')
    expect(formatUsageOutputRate({ ...textRow, duration_ms: value as number })).toBe('—')
  })

  it('excludes image output, image billing and unsupported request types', () => {
    for (const fields of [{ image_count: 1 }, { image_output_tokens: 1 }, { billing_mode: 'image' }]) {
      expect(formatUsageOutputRate({ ...textRow, ...fields })).toBe('—')
    }
    for (const request_type of ['live', 'unknown'] as const) {
      expect(formatUsageOutputRate({ ...textRow, request_type })).toBe('—')
    }
  })
})
