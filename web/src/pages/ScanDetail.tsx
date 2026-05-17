import { useState } from 'react'
import { useParams } from 'react-router'
import { Card, Descriptions, Table, Button, Space, Select, Popconfirm, message, Spin } from 'antd'
import { useTranslation } from 'react-i18next'
import { useScan, useScanResults, useCancelScan, useRetryScan } from '../api/hooks'
import StatusTag from '../components/StatusTag'
import SeverityTag from '../components/SeverityTag'

export default function ScanDetail() {
  const { id } = useParams<{ id: string }>()
  const { t } = useTranslation()
  const { data: scan, loading: scanLoading, reload: reloadScan } = useScan(id!)
  const [page, setPage] = useState(1)
  const [severity, setSeverity] = useState<string | undefined>()
  const { data: results, loading: resultsLoading } = useScanResults(id!, page, 20, severity)
  const { mutate: cancel } = useCancelScan()
  const { mutate: retry } = useRetryScan()

  const task = scan?.data
  const canCancel = task && (task.status === 'queued' || task.status === 'running')
  const canRetry = task && (task.status === 'failed' || task.status === 'cancelled')

  const handleCancel = async () => {
    await cancel(id!)
    message.success(t('scan.cancelSuccess'))
    reloadScan()
  }

  const handleRetry = async () => {
    await retry(id!)
    message.success(t('scan.retrySuccess'))
    reloadScan()
  }

  return (
    <Spin spinning={scanLoading}>
      <Card
        title={task?.name || t('scan.detail')}
        extra={
          <Space>
            {canCancel && (
              <Popconfirm title={t('scan.confirmCancel')} onConfirm={handleCancel}>
                <Button danger>{t('scan.cancel')}</Button>
              </Popconfirm>
            )}
            {canRetry && (
              <Popconfirm title={t('scan.confirmRetry')} onConfirm={handleRetry}>
                <Button type="primary">{t('scan.retry')}</Button>
              </Popconfirm>
            )}
          </Space>
        }
      >
        {task && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID">{task.id}</Descriptions.Item>
            <Descriptions.Item label={t('scan.status')}><StatusTag status={task.status} /></Descriptions.Item>
            <Descriptions.Item label={t('scan.resultCount')}>{task.result_count}</Descriptions.Item>
            <Descriptions.Item label={t('scan.assignedWorker')}>{task.assigned_worker_id || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('scan.createdAt')}>{new Date(task.created_at).toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label={t('scan.startedAt')}>{task.started_at ? new Date(task.started_at).toLocaleString() : '-'}</Descriptions.Item>
            <Descriptions.Item label={t('scan.completedAt')}>{task.completed_at ? new Date(task.completed_at).toLocaleString() : '-'}</Descriptions.Item>
            <Descriptions.Item label={t('scan.errorMessage')} span={2}>{task.error_message || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('scan.targets')} span={2}>
              <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{task.targets}</pre>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      <Card title={t('scan.scanResults')} style={{ marginTop: 16 }}>
        <div style={{ marginBottom: 16 }}>
          <Select
            allowClear
            placeholder={t('result.severity')}
            style={{ width: 160 }}
            value={severity}
            onChange={(v) => { setSeverity(v); setPage(1) }}
            options={[
              { value: 'critical', label: t('severity.critical') },
              { value: 'high', label: t('severity.high') },
              { value: 'medium', label: t('severity.medium') },
              { value: 'low', label: t('severity.low') },
              { value: 'info', label: t('severity.info') },
            ]}
          />
        </div>
        <Table
          dataSource={results?.data || []}
          rowKey="id"
          loading={resultsLoading}
          pagination={{
            current: page,
            pageSize: 20,
            total: results?.total || 0,
            onChange: setPage,
            showTotal: (total) => t('common.total', { count: total }),
          }}
          columns={[
            { title: t('result.severity'), dataIndex: 'severity', key: 'severity', width: 100, render: (s: string) => <SeverityTag severity={s} /> },
            { title: t('result.host'), dataIndex: 'host', key: 'host', ellipsis: true },
            { title: t('result.templateName'), dataIndex: 'template_name', key: 'template_name', ellipsis: true },
            { title: t('result.type'), dataIndex: 'type', key: 'type', width: 100 },
            { title: t('result.url'), dataIndex: 'url', key: 'url', ellipsis: true },
            { title: t('result.cveId'), dataIndex: 'cve_id', key: 'cve_id', width: 130 },
            {
              title: t('result.cvssScore'),
              dataIndex: 'cvss_score',
              key: 'cvss_score',
              width: 100,
              render: (v: number) => v > 0 ? v.toFixed(1) : '-',
            },
          ]}
        />
      </Card>
    </Spin>
  )
}
