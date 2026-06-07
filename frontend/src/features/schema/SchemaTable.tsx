import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { schemaApi } from '@/lib/api/schema'
import { formatError } from '@/lib/api/errors'
import { IndexList } from './IndexList'

export function SchemaTable({
  connectionId,
  tableName,
}: {
  connectionId: string
  tableName: string
}) {
  const { t } = useTranslation()
  const { data, isLoading, error } = useQuery({
    queryKey: ['schema', connectionId, tableName],
    queryFn: () => schemaApi.getTableSchema(connectionId, tableName),
  })

  if (isLoading) return <p className="p-4 text-sm text-muted">{t('common.loading')}</p>
  if (error) return <p className="p-4 text-sm text-red-500">{formatError(t, error)}</p>
  if (!data) return null

  const copy = (text: string) => navigator.clipboard.writeText(text)

  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center gap-2">
        <h3 className="font-mono text-sm">{data.name}</h3>
        <button className="text-xs text-accent" onClick={() => copy(data.name)}>
          {t('schema.copyTable')}
        </button>
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-left text-muted">
            <th className="py-2">{t('schema.column')}</th>
            <th>{t('schema.type')}</th>
            <th>{t('schema.nullable')}</th>
            <th>{t('schema.pk')}</th>
            <th>{t('schema.default')}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {data.columns.map((col) => (
            <tr key={col.name} className="border-b border-border/50">
              <td className="py-2 font-mono">{col.name}</td>
              <td>{col.dataType}</td>
              <td>{col.nullable ? t('common.yes') : t('common.no')}</td>
              <td>{col.primaryKey ? '✓' : ''}</td>
              <td className="font-mono text-xs">{col.defaultValue ?? ''}</td>
              <td>
                <button className="text-xs text-accent" onClick={() => copy(col.name)}>
                  {t('schema.copyColumn')}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <IndexList indexes={data.indexes} />
    </div>
  )
}
