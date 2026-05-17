import { Tag } from 'antd'
import { useTranslation } from 'react-i18next'

const colorMap: Record<string, string> = {
  created: 'default',
  queued: 'blue',
  running: 'processing',
  completed: 'success',
  failed: 'error',
  cancelled: 'warning',
  online: 'success',
  offline: 'error',
  busy: 'processing',
}

export default function StatusTag({ status }: { status: string }) {
  const { t } = useTranslation()
  const label = t(`status.${status}`, status)
  const color = colorMap[status] || 'default'
  return <Tag color={color}>{label}</Tag>
}
