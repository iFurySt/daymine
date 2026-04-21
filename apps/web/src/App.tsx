import { FormEvent, useEffect, useMemo, useState } from 'react'
import {
  Activity,
  CalendarDays,
  CheckCircle2,
  CircleAlert,
  Code2,
  FileText,
  Github,
  Newspaper,
  Play,
  RefreshCw,
  Rss,
} from 'lucide-react'
import { getDashboardConfig, getPanel, startRun } from './api'
import type { DashboardConfig, Page, PanelConfig, PanelResponse, RunRecord } from './types'

type PanelState = {
  data?: PanelResponse
  loading: boolean
  error?: string
}

const iconByType = {
  calendar: CalendarDays,
  feed: Rss,
  'article-list': Newspaper,
  'github-list': Github,
  'agent-runs': Activity,
  'markdown-view': FileText,
  'video-card': Play,
  'social-post': Code2,
}

export function App() {
  const [config, setConfig] = useState<DashboardConfig | null>(null)
  const [panelState, setPanelState] = useState<Record<string, PanelState>>({})
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('printf "# Agent note\\n\\nCreated by Daymine.\\n" > notes/daily/agent-note.md')
  const [provider, setProvider] = useState('local-command')
  const [running, setRunning] = useState(false)

  useEffect(() => {
    getDashboardConfig()
      .then(setConfig)
      .catch((err: Error) => setError(err.message))
  }, [])

  const page = config?.pages[0]

  useEffect(() => {
    if (!page) return
    for (const panel of page.panels) {
      loadPanel(panel.id)
    }
  }, [page?.name])

  const panelsById = useMemo(() => {
    const map = new Map<string, PanelConfig>()
    page?.panels.forEach((panel) => map.set(panel.id, panel))
    return map
  }, [page])

  async function loadPanel(id: string) {
    setPanelState((state) => ({ ...state, [id]: { ...state[id], loading: true, error: undefined } }))
    try {
      const data = await getPanel(id)
      setPanelState((state) => ({ ...state, [id]: { data, loading: false } }))
    } catch (err) {
      setPanelState((state) => ({
        ...state,
        [id]: { ...state[id], loading: false, error: err instanceof Error ? err.message : 'Failed to load panel' },
      }))
    }
  }

  async function handleRun(event: FormEvent) {
    event.preventDefault()
    setRunning(true)
    try {
      await startRun(query, provider)
      await Promise.all(['agent-runs', 'markdown-view'].map((id) => loadPanel(id)))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Agent run failed')
      await loadPanel('agent-runs')
    } finally {
      setRunning(false)
    }
  }

  if (error && !config) {
    return <FullState icon={<CircleAlert />} title="Could not load Daymine" detail={error} />
  }

  if (!page) {
    return <FullState icon={<RefreshCw className="spin" />} title="Loading workspace" detail="Reading local panel manifest..." />
  }

  return (
    <div className="shell">
      <Header page={page} />
      <main className="layout">
        <section className="agent-bar">
          <div>
            <div className="eyebrow">Agent Control</div>
            <h2>Run a local workspace task</h2>
          </div>
          <form onSubmit={handleRun} className="agent-form">
            <select value={provider} onChange={(event) => setProvider(event.target.value)} aria-label="Provider">
              <option value="local-command">local-command</option>
              <option value="codex-cli">codex-cli</option>
            </select>
            <input value={query} onChange={(event) => setQuery(event.target.value)} aria-label="Query" />
            <button type="submit" disabled={running}>
              {running ? <RefreshCw className="spin" size={15} /> : <Play size={15} />}
              Run
            </button>
          </form>
        </section>

        <section
          className="grid"
          style={{ gridTemplateColumns: page.column_widths.map((weight) => `${weight}fr`).join(' ') }}
        >
          {page.layout_by_column.map((column, index) => (
            <div className="column" key={index}>
              {column.map((id) => {
                const panel = panelsById.get(id)
                if (!panel) return null
                return <PanelCard key={id} panel={panel} state={panelState[id]} onRefresh={() => loadPanel(id)} />
              })}
            </div>
          ))}
        </section>
      </main>
    </div>
  )
}

function Header({ page }: { page: Page }) {
  return (
    <header className="header">
      <div>
        <div className="brand">Daymine</div>
        <div className="subtle">{page.title} / local self-hosted workspace</div>
      </div>
      <div className="status">
        <CheckCircle2 size={15} />
        No auth
      </div>
    </header>
  )
}

function PanelCard({ panel, state, onRefresh }: { panel: PanelConfig; state?: PanelState; onRefresh: () => void }) {
  const Icon = iconByType[panel.type as keyof typeof iconByType] ?? FileText
  return (
    <article className="panel">
      <header className="panel-header">
        <div className="panel-title">
          <Icon size={15} />
          <span>{panel.title}</span>
        </div>
        <button className="icon-button" onClick={onRefresh} aria-label={`Refresh ${panel.title}`}>
          <RefreshCw size={14} className={state?.loading ? 'spin' : ''} />
        </button>
      </header>
      <div className="panel-body">
        {state?.error && <EmptyState text={state.error} />}
        {!state?.error && state?.loading && !state.data && <EmptyState text="Loading..." />}
        {!state?.error && state?.data && <PanelRenderer panel={panel} response={state.data} />}
      </div>
    </article>
  )
}

function PanelRenderer({ panel, response }: { panel: PanelConfig; response: PanelResponse }) {
  switch (panel.type) {
    case 'calendar':
      return <CalendarPanel data={response.data} />
    case 'agent-runs':
      return <RunsPanel data={response.data} />
    case 'markdown-view':
      return <MarkdownPanel data={response.data} />
    case 'github-list':
      return <ListPanel data={response.data} titleKey="full_name" detailKey="description" metaKey="language" />
    case 'article-list':
      return <ListPanel data={response.data} titleKey="title" detailKey="path" metaKey="status" />
    default:
      return <ListPanel data={response.data} titleKey="title" detailKey="summary" metaKey="source" />
  }
}

function CalendarPanel({ data }: { data: Record<string, unknown> }) {
  const events = asArray(data.events)
  return (
    <div className="calendar-panel">
      <div className="date-number">{String(data.today ?? '')}</div>
      {events.map((event, index) => (
        <div className="row" key={index}>
          <span>{String(event.title ?? 'Event')}</span>
          <span className="muted">{String(event.time ?? event.date ?? '')}</span>
        </div>
      ))}
    </div>
  )
}

function RunsPanel({ data }: { data: Record<string, unknown> }) {
  const runs = asArray(data.runs) as unknown as RunRecord[]
  if (runs.length === 0) return <EmptyState text="No Agent runs yet" />
  return (
    <div className="stack">
      {runs.map((run) => (
        <div className="run" key={run.id}>
          <div className="run-top">
            <span className={`pill ${run.status}`}>{run.status}</span>
            <span className="muted">{run.provider}</span>
          </div>
          <div className="mono">{run.query}</div>
          {run.artifacts?.length ? <div className="muted">{run.artifacts.join(', ')}</div> : null}
        </div>
      ))}
    </div>
  )
}

function MarkdownPanel({ data }: { data: Record<string, unknown> }) {
  return (
    <pre className="markdown">
      {String(data.markdown ?? '')}
    </pre>
  )
}

function ListPanel({
  data,
  titleKey,
  detailKey,
  metaKey,
}: {
  data: Record<string, unknown>
  titleKey: string
  detailKey: string
  metaKey: string
}) {
  const items = asArray(data.items)
  if (items.length === 0) return <EmptyState text="No items" />
  return (
    <div className="stack">
      {items.map((item, index) => (
        <div className="item" key={index}>
          <div className="item-title">{String(item[titleKey] ?? item.name ?? 'Untitled')}</div>
          <div className="item-detail">{String(item[detailKey] ?? '')}</div>
          <div className="muted">{String(item[metaKey] ?? '')}</div>
        </div>
      ))}
    </div>
  )
}

function EmptyState({ text }: { text: string }) {
  return <div className="empty">{text}</div>
}

function FullState({ icon, title, detail }: { icon: React.ReactNode; title: string; detail: string }) {
  return (
    <div className="full-state">
      {icon}
      <h1>{title}</h1>
      <p>{detail}</p>
    </div>
  )
}

function asArray(value: unknown): Record<string, unknown>[] {
  return Array.isArray(value) ? (value as Record<string, unknown>[]) : []
}
