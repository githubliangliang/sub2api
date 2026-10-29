const { createHash } = require('node:crypto')

// Independently sample JavaScript strings; no Go helper or generated expected values.
const samples = [
  ['empty', ''],
  ['short_ascii', 'hi'],
  ['ascii', 'abcdefghijklmnopqrstuvwxyz'],
  ['chinese', '这是一个用于验证中文指纹计算的测试消息，请保持一致。'],
  ['short_chinese', '你好世界呀'],
  ['mixed', 'abcd中fg文ijklmnopqrst界'],
  ['emoji_before_indices', '😀abcdefghijklmnopqrstuvwxyz'],
  ['isolated_high_surrogate', 'abcd😀fghijklmnopqrstuvwxyz'],
  ['isolated_low_surrogate', 'abc😀efghijklmnopqrstuvwxyz'],
  ['sampled_surrogates_form_pair', 'abcd😀😀efghijklmnopqrstuvwxyz'],
  ['nul_is_not_missing', 'abcd\0fg\0hijklmnopqrs\0'],
  ['long_prompt', 'abcdefghijklmnopqrstuvwxyz' + '😀中文'.repeat(32768)],
]

for (const version of ['2.1.283', '2.1.258', '2.9.0']) {
  for (const [name, text] of samples) {
    const joined = [4, 7, 20].map(i => text[i] || '0').join('')
    const fingerprint = createHash('sha256')
      .update('59cf53e54c78' + joined + version)
      .digest('hex').slice(0, 3)
    console.log(JSON.stringify({ name, version, joined, fingerprint }))
  }
}
