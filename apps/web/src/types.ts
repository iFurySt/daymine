export interface DashboardConfig {
  pages: Page[]
}

export interface Page {
  name: string
  title: string
  column_widths: number[]
  layout_by_column: string[][]
  panels: PanelConfig[]
}

export interface PanelConfig {
  id: string
  type: PanelType
  title: string
  refresh?: string
  source?: string
  config?: Record<string, unknown>
}

export type PanelType =
  | 'calendar'
  | 'feed'
  | 'article-list'
  | 'video-card'
  | 'github-list'
  | 'social-post'
  | 'agent-runs'
  | 'markdown-view'
  | string

export interface PanelResponse {
  id: string
  type: PanelType
  title: string
  updated_at: string
  data: Record<string, unknown>
}

export interface RunRecord {
  id: string
  provider: string
  query: string
  status: string
  started_at: string
  completed_at?: string
  exit_code?: number
  output?: string
  error?: string
  artifacts?: string[]
}
