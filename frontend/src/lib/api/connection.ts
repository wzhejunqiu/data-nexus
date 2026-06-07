import {
  Attach,
  CloseConnection,
  CreateConnection,
  CreateRemoteConnection,
  Detach,
  GetRestoreOpenOnStartup,
  ListAttached,
  ListConnections,
  OpenConnection,
  OpenConnectionFromFile,
  RemoveConnection,
  RenameConnection,
  SetRestoreOpenOnStartup,
  TestConnection,
  UpdateConnectionMySQLSettings,
  UpdateConnectionPostgresSettings,
  UpdateConnectionSQLiteSettings,
} from '../../../wailsjs/go/wails/ConnectionService'
import { model } from '../../../wailsjs/go/models'
import type {
  AttachedDatabase,
  ConnectRequest,
  Connection,
  ConnectionListView,
  MySQLSettingsUpdate,
  PostgresSettingsUpdate,
  RemoteConnectRequest,
  SavedConnection,
  SQLiteSettingsUpdate,
  TestConnectionRequest,
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
  createRemote: (req: RemoteConnectRequest) =>
    wrap(() => CreateRemoteConnection(buildRemoteRequest(req)) as Promise<SavedConnection>),
  test: (req: TestConnectionRequest) => wrap(() => TestConnection(buildTestRequest(req))),
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
  updatePostgresSettings: (id: string, settings: PostgresSettingsUpdate) =>
    wrap(
      () =>
        UpdateConnectionPostgresSettings(
          id,
          model.PostgresSettingsUpdate.createFrom(settings),
        ) as Promise<SavedConnection>,
    ),
  updateMySQLSettings: (id: string, settings: MySQLSettingsUpdate) =>
    wrap(
      () =>
        UpdateConnectionMySQLSettings(
          id,
          model.MySQLSettingsUpdate.createFrom(settings),
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

function buildRemoteRequest(req: RemoteConnectRequest): model.RemoteConnectRequest {
  const out = new model.RemoteConnectRequest({
    type: req.type,
    name: req.name ?? '',
    password: req.password,
    open: req.open ?? false,
  })
  if (req.postgres) {
    out.postgres = model.PostgresConfig.createFrom(req.postgres)
  }
  if (req.mysql) {
    out.mysql = model.MySQLConfig.createFrom(req.mysql)
  }
  return out
}

function buildTestRequest(req: TestConnectionRequest): model.TestConnectionRequest {
  const out = new model.TestConnectionRequest({
    type: req.type,
    password: req.password,
  })
  if (req.postgres) {
    out.postgres = model.PostgresConfig.createFrom(req.postgres)
  }
  if (req.mysql) {
    out.mysql = model.MySQLConfig.createFrom(req.mysql)
  }
  return out
}
