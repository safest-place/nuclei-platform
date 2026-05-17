import { Tag } from 'antd'
import { useTranslation } from 'react-i18next'

const colorMap: Record<string, string> = {
  critical: '#a61d24',
  high: '#cf1322',
  medium: '#d46b08',
  low: '#389e0d',
  info: '#1677ff',
  unknown: '#8c8c8c',
}

export default function SeverityTag({ severity }: { severity: string }) {
  const { t } = useTranslation()
  const key = `severity.${severity}`
  const label = t(key) === key ? severity : t(key)
  return <Tag color={colorMap[severity] || '#8c8c8c'}>{label.toUpperCase()}</Tag>
}
