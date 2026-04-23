import type { DashboardConfig, PanelResponse, RunRecord, TaskRecord } from './types'

const jsonHeaders = { 'Content-Type': 'application/json' }

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init)
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    const message = typeof payload.error === 'string' ? payload.error : `Request failed: ${response.status}`
    throw new Error(message)
  }
  return payload as T
}

export function getDashboardConfig(): Promise<DashboardConfig> {
  return request<DashboardConfig>('/api/v1/dashboard/config')
}

export function getPanel(id: string): Promise<PanelResponse> {
  return request<PanelResponse>(`/api/v1/panels/${encodeURIComponent(id)}`)
}

export async function startRun(query: string, provider = 'local-command'): Promise<RunRecord> {
  const payload = await request<{ run: RunRecord }>('/api/v1/agent/runs', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ query, provider }),
  })
  return payload.run
}

export async function getTasks(): Promise<TaskRecord[]> {
  const payload = await request<{ tasks: TaskRecord[] }>('/api/v1/tasks')
  return payload.tasks
}

export async function runTask(taskId: string): Promise<RunRecord> {
  const payload = await request<{ run: RunRecord }>(`/api/v1/tasks/${encodeURIComponent(taskId)}/runs`, {
    method: 'POST',
    headers: jsonHeaders,
  })
  return payload.run
}
