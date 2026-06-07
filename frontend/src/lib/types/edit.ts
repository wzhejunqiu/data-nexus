export interface CellChange {
  columnName: string
  primaryKey: Record<string, unknown>
  newValue: unknown
}

export interface UpdateCellsBatchRequest {
  connectionId: string
  tableName: string
  changes: CellChange[]
}

export interface UpdateCellsBatchResult {
  updatedCount: number
}

export interface PendingEdit {
  columnName: string
  primaryKey: Record<string, unknown>
  originalValue: unknown
  newValue: unknown
}
