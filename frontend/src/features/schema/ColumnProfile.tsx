import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { schemaApi } from '@/lib/api/schema'
import { formatError } from '@/lib/api/errors'

export function ColumnProfile({
  connectionId,
  tableName,
}: {
  connectionId: string
  tableName: string
}) {
  const { t } = useTranslation()
  const { data, isLoading, error } = useQuery({
    queryKey: ['profile', connectionId, tableName],
    queryFn: () => schemaApi.getTableProfile(connectionId, tableName),
  })

  if (isLoading) return <p className="text-sm text-muted">{t('common.loading')}</p>
  if (error) return <p className="text-sm text-red-500">{formatError(t, error)}</p>
  if (!data) return null

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <h4 className="text-sm font-semibold">{t('profile.title')}</h4>
        {data.isSampled && (
          <span className="rounded bg-muted/40 px-2 py-0.5 text-xs text-muted">
            {t('profile.sampled', { count: data.sampledRows })}
          </span>
        )}
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-left text-muted">
            <th className="py-2">{t('schema.column')}</th>
            <th>{t('profile.distinct')}</th>
            <th>{t('profile.nullPercent')}</th>
            <th>{t('profile.min')}</th>
            <th>{t('profile.max')}</th>
            <th>{t('profile.topValues')}</th>
          </tr>
        </thead>
        <tbody>
          {data.columns.map((col) => (
            <tr key={col.name} className="border-b border-border/50 align-top">
              <td className="py-2 font-mono">{col.name}</td>
              <td>{col.distinctCount ?? '—'}</td>
              <td>{col.nullPercent != null ? `${col.nullPercent.toFixed(1)}%` : '—'}</td>
              <td className="font-mono text-xs">{col.minValue ?? '—'}</td>
              <td className="font-mono text-xs">{col.maxValue ?? '—'}</td>
              <td>
                {col.topValues && col.topValues.length > 0 ? (
                  <ul className="space-y-0.5 text-xs">
                    {col.topValues.map((tv) => (
                      <li key={tv.value} className="flex justify-between gap-2">
                        <span className="truncate font-mono">{tv.value}</span>
                        <span className="text-muted">{tv.count}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  '—'
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
