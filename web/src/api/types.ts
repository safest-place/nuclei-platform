export type TaskStatus = 'created' | 'queued' | 'running' | 'completed' | 'failed' | 'cancelled'

export interface Task {
  id: string
  name: string
  targets?: string
  template_filters?: string
  concurrency?: string
  rate_limit?: number
  headers?: string
  status: TaskStatus
  assigned_worker_id?: string
  error_message?: string
  result_count: number
  created_at: string
  started_at?: string
  completed_at?: string
  updated_at: string
}

export interface Result {
  id: string
  task_id: string
  template_id: string
  template_name: string
  host: string
  matched_at: string
  severity: string
  ip: string
  port: string
  scheme: string
  url: string
  request?: string
  response?: string
  curl_command?: string
  extracted_results?: string
  matcher_name?: string
  type: string
  cve_id?: string
  cvss_score?: number
  timestamp: string
  created_at: string
}

export interface Worker {
  id: string
  hostname: string
  ip: string
  status: string
  capabilities?: string
  last_heartbeat: string
  disabled: boolean
  created_at: string
  updated_at: string
}

export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: string
}

export interface PaginatedResponse<T> {
  success: boolean
  data: T[]
  total: number
  page: number
  per_page: number
}

export interface StatsResponse {
  by_severity: { severity: string; count: number }[]
  by_host: { host: string; count: number }[]
}

export interface CreateScanRequest {
  name: string
  targets: string[]
  template_filters?: Record<string, unknown>
  concurrency?: Record<string, unknown>
  rate_limit?: number
  headers?: string[]
}
