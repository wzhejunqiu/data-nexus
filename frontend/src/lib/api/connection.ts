import {
  CloseConnection,
  CreateConnection,
  GetRestoreOpenOnStartup,
  ListConnections,
  OpenConnection,
  OpenConnectionFromFile,
  RemoveConnection,
  RenameConnection,
  SetRestoreOpenOnStartup,
  UpdateConnectionReadOnly,
} from '../../../wailsjs/go/wails/ConnectionService'
import type { ConnectRequest, Connection, ConnectionListView, SavedConnection } from '../types'
import { mapWailsError } from './errors'

async function wrap<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    throw mapWailsError(err)
  }
}

export const connectionApi = {
  list: () => wrap(() => ListConnections() as Promise<ConnectionListView>),
  create: (req: ConnectRequest) => wrap(() => CreateConnection(req) as Promise<SavedConnection>),
  open: (id: string) => wrap(() => OpenConnection(id) as Promise<Connection>),
  openFromFile: (req: ConnectRequest) =>
    wrap(() => OpenConnectionFromFile(req) as Promise<Connection>),
  close: (id: string) => wrap(() => CloseConnection(id)),
  remove: (id: string) => wrap(() => RemoveConnection(id)),
  rename: (id: string, name: string) =>
    wrap(() => RenameConnection(id, name) as Promise<SavedConnection>),
  updateReadOnly: (id: string, readOnly: boolean) =>
    wrap(() => UpdateConnectionReadOnly(id, readOnly) as Promise<SavedConnection>),
  getRestoreOpenOnStartup: () => wrap(() => GetRestoreOpenOnStartup() as Promise<boolean>),
  setRestoreOpenOnStartup: (enabled: boolean) => wrap(() => SetRestoreOpenOnStartup(enabled)),
}
