export const MAX_BATCH_EDITS = 200

export function isBatchOverLimit(pendingCount: number, maxBatch = MAX_BATCH_EDITS): boolean {
  return pendingCount > maxBatch
}
