import type { ApiResponse, PaginatedResponse, Task, Result, Worker, StatsResponse, CreateScanRequest } from './types'

const BASE = '/api/v1'

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  const json = await res.json()
  if (!res.ok) {
    throw new Error(json.error || `HTTP ${res.status}`)
  }
  return json as T
}

// Scans
export function createScan(req: CreateScanRequest) {
  return request<ApiResponse<Task>>(`${BASE}/scans/`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export function listScans(page = 1, perPage = 20, status?: string) {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) })
  if (status) params.set('status', status)
  return request<PaginatedResponse<Task>>(`${BASE}/scans/?${params}`)
}

export function getScan(id: string) {
  return request<ApiResponse<Task>>(`${BASE}/scans/${id}`)
}

export function cancelScan(id: string) {
  return request<ApiResponse<Task>>(`${BASE}/scans/${id}`, { method: 'DELETE' })
}

export function retryScan(id: string) {
  return request<ApiResponse<Task>>(`${BASE}/scans/${id}/retry`, { method: 'POST' })
}

export function listScanResults(taskId: string, page = 1, perPage = 20, severity?: string) {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) })
  if (severity) params.set('severity', severity)
  return request<PaginatedResponse<Result>>(`${BASE}/scans/${taskId}/results?${params}`)
}

// Results
export function listResults(page = 1, perPage = 20, filters?: { severity?: string; host?: string; template_id?: string }) {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) })
  if (filters?.severity) params.set('severity', filters.severity)
  if (filters?.host) params.set('host', filters.host)
  if (filters?.template_id) params.set('template_id', filters.template_id)
  return request<PaginatedResponse<Result>>(`${BASE}/results/?${params}`)
}

export function getStats() {
  return request<ApiResponse<StatsResponse>>(`${BASE}/results/stats`)
}

// Workers
export function listWorkers() {
  return request<ApiResponse<Worker[]>>(`${BASE}/workers/`)
}

export function getWorker(id: string) {
  return request<ApiResponse<Worker>>(`${BASE}/workers/${id}`)
}

export function disableWorker(id: string) {
  return request<ApiResponse<Worker>>(`${BASE}/workers/${id}/disable`, { method: 'POST' })
}

export function enableWorker(id: string) {
  return request<ApiResponse<Worker>>(`${BASE}/workers/${id}/enable`, { method: 'POST' })
}
