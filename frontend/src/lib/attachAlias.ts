const MAX_IDENTIFIER_LEN = 128
const RESERVED_ALIASES = new Set(['main', 'temp'])

export type AttachAliasError = 'required' | 'invalid' | 'reserved' | 'duplicate'

export function isSafeQuotedIdentifier(name: string): boolean {
  if (name === '' || name.length > MAX_IDENTIFIER_LEN) return false
  if (name.trim() === '') return false
  for (let i = 0; i < name.length; i++) {
    const code = name.charCodeAt(i)
    if (code === 0 || code === 34) return false
  }
  return true
}

export function validateAttachAlias(
  alias: string,
  existingAliases: string[],
): AttachAliasError | null {
  const trimmed = alias.trim()
  if (!trimmed) return 'required'
  if (!isSafeQuotedIdentifier(trimmed)) return 'invalid'
  if (RESERVED_ALIASES.has(trimmed.toLowerCase())) return 'reserved'
  const lower = trimmed.toLowerCase()
  if (existingAliases.some((a) => a.toLowerCase() === lower)) return 'duplicate'
  return null
}
