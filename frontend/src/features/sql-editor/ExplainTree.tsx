import type { QueryResponse } from '@/lib/types'

interface ExplainNode {
  id: number
  parent: number
  detail: string
  children: ExplainNode[]
}

function buildExplainTree(rows: Record<string, unknown>[]): ExplainNode[] {
  const nodes = rows.map((row, i) => ({
    id: Number(row.id ?? row.ID ?? i),
    parent: Number(row.parent ?? row.parentid ?? row.parentId ?? -1),
    detail: String(row.detail ?? row.DETAIL ?? ''),
    children: [] as ExplainNode[],
  }))
  const byId = new Map<number, ExplainNode>()
  for (const n of nodes) byId.set(n.id, n)
  const roots: ExplainNode[] = []
  for (const n of nodes) {
    if (n.parent < 0 || !byId.has(n.parent)) roots.push(n)
    else byId.get(n.parent)!.children.push(n)
  }
  return roots.length > 0 ? roots : nodes
}

function ExplainNodeView({ node, depth }: { node: ExplainNode; depth: number }) {
  return (
    <li className="font-mono text-xs">
      <span style={{ paddingLeft: depth * 12 }}>{node.detail}</span>
      {node.children.length > 0 && (
        <ul className="mt-1">
          {node.children.map((c) => (
            <ExplainNodeView key={c.id} node={c} depth={depth + 1} />
          ))}
        </ul>
      )}
    </li>
  )
}

export function ExplainTree({ result }: { result: QueryResponse }) {
  if (result.kind !== 'result' || !result.rows?.length) return null
  const roots = buildExplainTree(result.rows)
  return (
    <div className="rounded border border-border bg-muted/20 p-3">
      <ul>
        {roots.map((n) => (
          <ExplainNodeView key={n.id} node={n} depth={0} />
        ))}
      </ul>
    </div>
  )
}

export function isExplainResult(sql: string, result: QueryResponse): boolean {
  return /^\s*EXPLAIN\b/i.test(sql) && result.kind === 'result'
}
