import { createElement, FormEvent, useEffect, useMemo, useState } from 'react'
import {
  Activity,
  CalendarDays,
  CheckCircle2,
  CircleAlert,
  Code2,
  ExternalLink,
  FileText,
  Github,
  MessageSquare,
  Newspaper,
  Play,
  RefreshCw,
  Rss,
  Star,
} from 'lucide-react'
import { getDashboardConfig, getPanel, runTask, startRun } from './api'
import type { DashboardConfig, Page, PanelConfig, PanelRenderer as PanelRendererConfig, PanelResponse, RunRecord } from './types'

type PanelState = {
  data?: PanelResponse
  loading: boolean
  error?: string
}

const iconByType = {
  calendar: CalendarDays,
  feed: Rss,
  'hacker-news-top': Newspaper,
  'article-list': Newspaper,
  'github-list': Github,
  'agent-runs': Activity,
  'markdown-view': FileText,
  'video-card': Play,
  'social-post': Code2,
  'html-template': Code2,
}

export function App() {
  const [config, setConfig] = useState<DashboardConfig | null>(null)
  const [panelState, setPanelState] = useState<Record<string, PanelState>>({})
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('printf "# Agent note\\n\\nCreated by Daymine.\\n" > notes/daily/agent-note.md')
  const [provider, setProvider] = useState('local-command')
  const [running, setRunning] = useState(false)
  const [taskRunning, setTaskRunning] = useState<Record<string, boolean>>({})

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

  async function handleRunTask(taskId: string, panel: PanelConfig) {
    setTaskRunning((state) => ({ ...state, [taskId]: true }))
    try {
      await runTask(taskId)
      const refreshIds = Array.from(new Set([panel.id, 'agent-runs']))
      await Promise.all(refreshIds.map((id) => loadPanel(id)))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Task run failed')
      await loadPanel('agent-runs')
    } finally {
      setTaskRunning((state) => ({ ...state, [taskId]: false }))
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
                return (
                  <PanelCard
                    key={id}
                    panel={panel}
                    state={panelState[id]}
                    taskRunning={taskRunning}
                    onRefresh={() => loadPanel(id)}
                    onRunTask={(taskId) => handleRunTask(taskId, panel)}
                  />
                )
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

function PanelCard({
  panel,
  state,
  taskRunning,
  onRefresh,
  onRunTask,
}: {
  panel: PanelConfig
  state?: PanelState
  taskRunning: Record<string, boolean>
  onRefresh: () => void
  onRunTask: (taskId: string) => void
}) {
  const Icon = iconByType[panel.type as keyof typeof iconByType] ?? FileText
  const configuredTaskId = panel.config?.task_id
  const taskId = typeof configuredTaskId === 'string' ? configuredTaskId : ''
  return (
    <article className="panel">
      <header className="panel-header">
        <div className="panel-title">
          <Icon size={15} />
          <span>{panel.title}</span>
        </div>
        <div className="panel-actions">
          {taskId && (
            <button className="task-button" onClick={() => onRunTask(taskId)} disabled={taskRunning[taskId]}>
              {taskRunning[taskId] ? <RefreshCw className="spin" size={13} /> : <Play size={13} />}
              {taskRunning[taskId] ? 'Running' : 'Run'}
            </button>
          )}
          <button className="icon-button" onClick={onRefresh} aria-label={`Refresh ${panel.title}`}>
            <RefreshCw size={14} className={state?.loading ? 'spin' : ''} />
          </button>
        </div>
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
  if (response.renderer?.type === 'html-template' || panel.renderer?.type === 'html-template') {
    return <HtmlTemplatePanel panelId={panel.id} renderer={response.renderer ?? panel.renderer} data={response.data} />
  }

  switch (panel.type) {
    case 'calendar':
      return <CalendarPanel data={response.data} />
    case 'agent-runs':
      return <RunsPanel data={response.data} />
    case 'hacker-news-top':
      return <HackerNewsPanel data={response.data} />
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

function HtmlTemplatePanel({
  panelId,
  renderer,
  data,
}: {
  panelId: string
  renderer?: PanelRendererConfig
  data: Record<string, unknown>
}) {
  if (!renderer?.template) {
    return <EmptyState text="Template is empty" />
  }
  return (
    <div className="html-template-panel" data-panel-id={panelId}>
      {renderer.style ? <style>{scopePanelCSS(renderer.style, panelId)}</style> : null}
      {renderTemplate(renderer.template, data)}
    </div>
  )
}

function HackerNewsPanel({ data }: { data: Record<string, unknown> }) {
  const items = asArray(data.items)
  if (items.length === 0) return <EmptyState text={String(data.message ?? 'No Hacker News digest yet')} />
  return (
    <div className="hn-panel">
      <div className="hn-meta">
        <span>{String(data.generated_at ?? '')}</span>
        <span>{String(data.window_hours ?? 24)}h window</span>
      </div>
      <div className="stack">
        {items.map((item, index) => (
          <div className="hn-item" key={String(item.id ?? index)}>
            <div className="hn-rank">{String(item.rank ?? index + 1).padStart(2, '0')}</div>
            <div className="hn-content">
              <a className="hn-title" href={String(item.url || item.hn_url || '#')} target="_blank" rel="noreferrer">
                {String(item.title ?? 'Untitled')}
                <ExternalLink size={12} />
              </a>
              <div className="hn-summary">{String(item.summary ?? '')}</div>
              <div className="hn-stats">
                <span>
                  <Star size={12} />
                  {String(item.score ?? 0)}
                </span>
                <span>
                  <MessageSquare size={12} />
                  {String(item.comments ?? 0)}
                </span>
                {item.by ? <span>{String(item.by)}</span> : null}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
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
            <span className="muted">{run.task_id ?? run.provider}</span>
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

function renderTemplate(template: string, context: Record<string, unknown>): React.ReactNode {
  const document = new DOMParser().parseFromString(`<template>${template}</template>`, 'text/html')
  const root = document.querySelector('template')
  if (!root) return null

  return Array.from(root.content.childNodes).map((node, index) => renderNode(node, context, index))
}

function renderNode(node: Node, context: Record<string, unknown>, key: React.Key): React.ReactNode {
  if (node.nodeType === Node.TEXT_NODE) {
    return interpolate(node.textContent ?? '', context)
  }

  if (node.nodeType !== Node.ELEMENT_NODE) {
    return null
  }

  const element = node as Element
  const tag = element.tagName.toLowerCase()
  if (blockedTags.has(tag)) {
    return null
  }

  const condition = element.getAttribute('data-if')
  if (condition && !resolvePath(context, condition)) {
    return null
  }

  const loop = element.getAttribute('data-for')
  if (loop) {
    const match = loop.match(/^\s*([A-Za-z_$][\w$]*)\s+in\s+([\w.$]+)\s*$/)
    if (!match) return null
    const [, itemName, listPath] = match
    const list = resolvePath(context, listPath)
    if (!Array.isArray(list)) return null
    return list.map((item, index) => {
      const loopContext = { ...context, [itemName]: item, index }
      const clone = element.cloneNode(true) as Element
      clone.removeAttribute('data-for')
      return renderNode(clone, loopContext, `${String(key)}-${index}`)
    })
  }

  const children = Array.from(element.childNodes).map((child, index) => renderNode(child, context, index))
  const props = safeProps(element, context)

  switch (tag) {
    case 'dm-list':
      return <div key={key} className="stack">{children}</div>
    case 'dm-item':
    case 'dm-card':
      return <div key={key} className="item template-item">{children}</div>
    case 'dm-title':
      return <div key={key} className="item-title">{children}</div>
    case 'dm-text':
      return <div key={key} className={element.getAttribute('tone') === 'muted' ? 'item-detail muted' : 'item-detail'}>{children}</div>
    case 'dm-meta':
      return <div key={key} className="muted">{children}</div>
    case 'dm-badge':
      return <span key={key} className="pill">{children}</span>
    case 'dm-link':
      return <a key={key} className="item-title template-link" {...props}>{children}</a>
    case 'dm-markdown':
      return <pre key={key} className="markdown">{children}</pre>
    default:
      if (!allowedTags.has(tag)) return <span key={key}>{children}</span>
      return createElement(tag, { key, ...props }, children)
  }
}

function safeProps(element: Element, context: Record<string, unknown>): Record<string, unknown> {
  const props: Record<string, unknown> = {}
  for (const attr of Array.from(element.attributes)) {
    const name = attr.name.toLowerCase()
    if (name.startsWith('on') || name === 'style' || name === 'data-for' || name === 'data-if') continue
    if (name === 'class') {
      props.className = interpolate(attr.value, context)
      continue
    }
    if (name === 'href') {
      const href = interpolate(attr.value, context)
      if (isSafeURL(href)) {
        props.href = href
        props.target = '_blank'
        props.rel = 'noreferrer'
      }
      continue
    }
    if (name === 'src') {
      const src = interpolate(attr.value, context)
      if (isSafeURL(src)) props.src = src
      continue
    }
    if (name.startsWith('aria-') || name === 'title' || name === 'alt') {
      props[name] = interpolate(attr.value, context)
    }
  }
  return props
}

function interpolate(value: string, context: Record<string, unknown>): string {
  return value.replace(/\{\{\s*([\w.$]+)\s*\}\}/g, (_, path: string) => {
    const resolved = resolvePath(context, path)
    if (resolved == null) return ''
    if (typeof resolved === 'object') return JSON.stringify(resolved)
    return String(resolved)
  })
}

function resolvePath(context: Record<string, unknown>, path: string): unknown {
  const clean = path.replace(/^\$\./, '')
  return clean.split('.').reduce<unknown>((value, part) => {
    if (value == null || typeof value !== 'object') return undefined
    return (value as Record<string, unknown>)[part]
  }, context)
}

function isSafeURL(value: string): boolean {
  if (!value) return false
  try {
    const parsed = new URL(value, window.location.origin)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' || parsed.protocol === 'mailto:'
  } catch {
    return false
  }
}

function scopePanelCSS(css: string, panelId: string): string {
  const scope = `.html-template-panel[data-panel-id="${CSS.escape(panelId)}"]`
  return css
    .split('}')
    .map((rule) => {
      const [selector, body] = rule.split('{')
      if (!selector || !body || selector.includes('@') || selector.includes('html') || selector.includes('body')) return ''
      const scopedSelector = selector
        .split(',')
        .map((part) => `${scope} ${translateTemplateSelector(part.trim())}`)
        .join(', ')
      return `${scopedSelector}{${body}}`
    })
    .join('\n')
}

function translateTemplateSelector(selector: string): string {
  return selector
    .replace(/\bdm-item\b/g, '.template-item')
    .replace(/\bdm-card\b/g, '.template-item')
    .replace(/\bdm-link\b/g, '.template-link')
    .replace(/\bdm-text\b/g, '.item-detail')
    .replace(/\bdm-meta\b/g, '.muted')
    .replace(/\bdm-title\b/g, '.item-title')
    .replace(/\bdm-badge\b/g, '.pill')
}

const blockedTags = new Set(['script', 'iframe', 'object', 'embed', 'link', 'meta', 'style'])
const allowedTags = new Set(['div', 'section', 'header', 'footer', 'ul', 'ol', 'li', 'p', 'span', 'a', 'img', 'strong', 'em', 'small', 'code', 'pre'])
