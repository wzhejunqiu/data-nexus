import { useTranslation } from 'react-i18next'
import type { ConnectionListItem } from '@/lib/types'
import { NamespaceNode } from './NamespaceNode'
import { useConnectionTree } from './useConnectionTree'

export function ConnectionSchemaTree({
  item,
  onSelectTable,
  searchQuery = '',
}: {
  item: ConnectionListItem
  onSelectTable: (table: string, ctx?: { database?: string; schema?: string }) => void
  searchQuery?: string
}) {
  const { t } = useTranslation()
  const { nsItems, expandedNs, setExpandedNs, expandedSchema, setExpandedSchema } =
    useConnectionTree(item.id, item.status === 'open')

  const q = searchQuery.trim().toLowerCase()

  return (
    <div className="ml-3 mt-1 border-l border-border/50 pl-2">
      {nsItems.map((ns) => {
        const nsKey = ns.name
        const autoExpand = Boolean(q)
        const nsOpen = autoExpand || (expandedNs[nsKey] ?? false)
        return (
          <NamespaceNode
            key={nsKey}
            ns={ns}
            nsOpen={nsOpen}
            connectionId={item.id}
            driverType={item.type}
            searchQuery={q}
            expandedSchema={expandedSchema}
            setExpandedSchema={setExpandedSchema}
            onToggleExpand={() => setExpandedNs((s) => ({ ...s, [nsKey]: !nsOpen }))}
            onExpand={() => setExpandedNs((s) => ({ ...s, [nsKey]: true }))}
            onSelectTable={onSelectTable}
          />
        )
      })}
      {nsItems.length === 0 && <p className="text-xs text-muted">{t('common.loading')}</p>}
    </div>
  )
}
