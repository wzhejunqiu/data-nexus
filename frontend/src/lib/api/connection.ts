import {
  Attach,
  CloseConnection,
  CreateConnection,
  Detach,
  GetRestoreOpenOnStartup,
  ListAttached,
  ListConnections,
  OpenConnection,
  OpenConnectionFromFile,
  RemoveConnection,
  RenameConnection,
  SetRestoreOpenOnStartup,
  UpdateConnectionSQLiteSettings,
} from '../../../wailsjs/go/wails/ConnectionService'
import { model } from '../../../wailsjs/go/models'
import type {
  AttachedDatabase,
  ConnectRequest,
  Connection,
  ConnectionListView,
  SavedConnection,
  SQLiteSettingsUpdate,
} from '../types'
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
  create: (req: ConnectRequest) =>
    wrap(() => CreateConnection(model.ConnectRequest.createFrom(req)) as Promise<SavedConnection>),
  open: (id: string) => wrap(() => OpenConnection(id) as Promise<Connection>),
  openFromFile: (req: ConnectRequest) =>
    wrap(() => OpenConnectionFromFile(model.ConnectRequest.createFrom(req)) as Promise<Connection>),
  close: (id: string) => wrap(() => CloseConnection(id)),
  remove: (id: string) => wrap(() => RemoveConnection(id)),
  rename: (id: string, name: string) =>
    wrap(() => RenameConnection(id, name) as Promise<SavedConnection>),
  updateSQLiteSettings: (id: string, settings: SQLiteSettingsUpdate) =>
    wrap(
      () =>
        UpdateConnectionSQLiteSettings(
          id,
          model.SQLiteSettingsUpdate.createFrom(settings),
        ) as Promise<SavedConnection>,
    ),
  getRestoreOpenOnStartup: () => wrap(() => GetRestoreOpenOnStartup() as Promise<boolean>),
  setRestoreOpenOnStartup: (enabled: boolean) => wrap(() => SetRestoreOpenOnStartup(enabled)),
  attach: (connectionId: string, filePath: string, alias: string) =>
    wrap(() => Attach(connectionId, filePath, alias)),
  detach: (connectionId: string, alias: string) => wrap(() => Detach(connectionId, alias)),
  listAttached: (connectionId: string) =>
    wrap(() => ListAttached(connectionId) as Promise<AttachedDatabase[]>),
}
