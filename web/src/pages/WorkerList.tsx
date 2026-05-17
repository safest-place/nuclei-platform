import { Table, Card, Button, Space, Popconfirm, message } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useWorkers, useToggleWorker } from '../api/hooks'
import StatusTag from '../components/StatusTag'

export default function WorkerList() {
  const { t } = useTranslation()
  const { data, loading, reload } = useWorkers()
  const { mutate: toggle } = useToggleWorker()

  const handleToggle = async (id: string, disable: boolean) => {
    await toggle(id, disable)
    message.success(disable ? t('worker.disabled') : t('worker.enabled'))
    reload()
  }

  return (
    <Card
      title={t('worker.title')}
      extra={<Button icon={<ReloadOutlined />} onClick={reload}>{t('common.refresh')}</Button>}
    >
      <Table
        dataSource={data?.data || []}
        rowKey="id"
        loading={loading}
        pagination={false}
        columns={[
          { title: t('worker.hostname'), dataIndex: 'hostname', key: 'hostname' },
          { title: t('worker.ip'), dataIndex: 'ip', key: 'ip', width: 150 },
          {
            title: t('worker.status'),
            dataIndex: 'status',
            key: 'status',
            width: 120,
            render: (s: string) => <StatusTag status={s} />,
          },
          {
            title: t('worker.lastHeartbeat'),
            dataIndex: 'last_heartbeat',
            key: 'last_heartbeat',
            width: 200,
            render: (v: string) => new Date(v).toLocaleString(),
          },
          {
            title: t('worker.disabled'),
            dataIndex: 'disabled',
            key: 'disabled',
            width: 100,
            render: (v: boolean) => v ? t('worker.disabled') : t('worker.enabled'),
          },
          {
            title: t('worker.actions'),
            key: 'actions',
            width: 120,
            render: (_: unknown, record: { id: string; disabled: boolean }) => (
              <Space>
                {record.disabled ? (
                  <Popconfirm title={t('worker.enable') + '?'} onConfirm={() => handleToggle(record.id, false)}>
                    <Button size="small" type="primary">{t('worker.enable')}</Button>
                  </Popconfirm>
                ) : (
                  <Popconfirm title={t('worker.disable') + '?'} onConfirm={() => handleToggle(record.id, true)}>
                    <Button size="small" danger>{t('worker.disable')}</Button>
                  </Popconfirm>
                )}
              </Space>
            ),
          },
        ]}
      />
    </Card>
  )
}
