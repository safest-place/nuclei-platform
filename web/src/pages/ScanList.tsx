import { useState } from 'react'
import { Table, Button, Space, Select, Popconfirm, message } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router'
import { useScans, useCancelScan, useRetryScan } from '../api/hooks'
import StatusTag from '../components/StatusTag'

export default function ScanList() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<string | undefined>()
  const { data, loading, reload } = useScans(page, 20, status)
  const { mutate: cancel } = useCancelScan()
  const { mutate: retry } = useRetryScan()

  const handleCancel = async (id: string) => {
    await cancel(id)
    message.success(t('scan.cancelSuccess'))
    reload()
  }

  const handleRetry = async (id: string) => {
    await retry(id)
    message.success(t('scan.retrySuccess'))
    reload()
  }

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Select
            allowClear
            placeholder={t('scan.status')}
            style={{ width: 160 }}
            value={status}
            onChange={(v) => { setStatus(v); setPage(1) }}
            options={[
              { value: 'created', label: t('status.created') },
              { value: 'queued', label: t('status.queued') },
              { value: 'running', label: t('status.running') },
              { value: 'completed', label: t('status.completed') },
              { value: 'failed', label: t('status.failed') },
              { value: 'cancelled', label: t('status.cancelled') },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={reload}>{t('common.refresh')}</Button>
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/scans/create')}>
          {t('scan.create')}
        </Button>
      </div>

      <Table
        dataSource={data?.data || []}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize: 20,
          total: data?.total || 0,
          onChange: setPage,
          showTotal: (total) => t('common.total', { count: total }),
        }}
        columns={[
          { title: t('scan.name'), dataIndex: 'name', key: 'name', ellipsis: true },
          {
            title: t('scan.status'),
            dataIndex: 'status',
            key: 'status',
            width: 120,
            render: (s: string) => <StatusTag status={s} />,
          },
          { title: t('scan.resultCount'), dataIndex: 'result_count', key: 'result_count', width: 100 },
          {
            title: t('scan.createdAt'),
            dataIndex: 'created_at',
            key: 'created_at',
            width: 180,
            render: (v: string) => new Date(v).toLocaleString(),
          },
          {
            title: t('common.actions'),
            key: 'actions',
            width: 200,
            render: (_: unknown, record: { id: string; status: string }) => (
              <Space>
                <Button size="small" onClick={() => navigate(`/scans/${record.id}`)}>
                  {t('common.view')}
                </Button>
                {(record.status === 'queued' || record.status === 'running') && (
                  <Popconfirm title={t('scan.confirmCancel')} onConfirm={() => handleCancel(record.id)}>
                    <Button size="small" danger>{t('scan.cancel')}</Button>
                  </Popconfirm>
                )}
                {(record.status === 'failed' || record.status === 'cancelled') && (
                  <Popconfirm title={t('scan.confirmRetry')} onConfirm={() => handleRetry(record.id)}>
                    <Button size="small" type="primary">{t('scan.retry')}</Button>
                  </Popconfirm>
                )}
              </Space>
            ),
          },
        ]}
      />
    </>
  )
}
