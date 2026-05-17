import { useState, useEffect, useCallback } from 'react'
import * as api from './client'
import type { Task, Result, Worker, StatsResponse, PaginatedResponse, CreateScanRequest } from './types'

function useAsync<T>(fn: () => Promise<T>, deps: unknown[] = []) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const reload = useCallback(() => {
    setLoading(true)
    setError(null)
    fn()
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }, deps) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => { reload() }, [reload])

  return { data, loading, error, reload }
}

export function useScans(page: number, perPage: number, status?: string) {
  return useAsync(() => api.listScans(page, perPage, status), [page, perPage, status])
}

export function useScan(id: string) {
  return useAsync(() => api.getScan(id), [id])
}

export function useScanResults(taskId: string, page: number, perPage: number, severity?: string) {
  return useAsync(() => api.listScanResults(taskId, page, perPage, severity), [taskId, page, perPage, severity])
}

export function useResults(page: number, perPage: number, filters?: { severity?: string; host?: string; template_id?: string }) {
  return useAsync(() => api.listResults(page, perPage, filters), [page, perPage, filters])
}

export function useStats() {
  return useAsync(() => api.getStats(), [])
}

export function useWorkers() {
  return useAsync(() => api.listWorkers(), [])
}

export function useCreateScan() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const mutate = useCallback(async (req: CreateScanRequest) => {
    setLoading(true)
    setError(null)
    try {
      const res = await api.createScan(req)
      return res.data
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(msg)
      throw e
    } finally {
      setLoading(false)
    }
  }, [])

  return { mutate, loading, error }
}

export function useCancelScan() {
  const [loading, setLoading] = useState(false)
  const mutate = useCallback(async (id: string) => {
    setLoading(true)
    try {
      const res = await api.cancelScan(id)
      return res.data
    } finally {
      setLoading(false)
    }
  }, [])
  return { mutate, loading }
}

export function useRetryScan() {
  const [loading, setLoading] = useState(false)
  const mutate = useCallback(async (id: string) => {
    setLoading(true)
    try {
      const res = await api.retryScan(id)
      return res.data
    } finally {
      setLoading(false)
    }
  }, [])
  return { mutate, loading }
}

export function useToggleWorker() {
  const [loading, setLoading] = useState(false)
  const mutate = useCallback(async (id: string, disable: boolean) => {
    setLoading(true)
    try {
      const res = disable ? await api.disableWorker(id) : await api.enableWorker(id)
      return res.data
    } finally {
      setLoading(false)
    }
  }, [])
  return { mutate, loading }
}

export type { Task, Result, Worker, StatsResponse, PaginatedResponse }
