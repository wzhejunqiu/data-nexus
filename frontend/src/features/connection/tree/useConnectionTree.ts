import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { schemaApi } from '@/lib/api/schema'

export function useConnectionTree(connectionId: string, enabled: boolean) {
  const [expandedNs, setExpandedNs] = useState<Record<string, boolean>>({})
  const [expandedSchema, setExpandedSchema] = useState<Record<string, boolean>>({})

  const { data: namespaces } = useQuery({
    queryKey: ['namespaces', connectionId],
    queryFn: () => schemaApi.listNamespaces(connectionId),
    enabled,
  })

  return {
    nsItems: namespaces?.items ?? [],
    expandedNs,
    setExpandedNs,
    expandedSchema,
    setExpandedSchema,
  }
}
