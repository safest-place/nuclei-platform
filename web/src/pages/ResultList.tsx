import { useState } from 'react'
import { Table, Card, Space, Select, Input, Button } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useResults } from '../api/hooks'
import SeverityTag from '../components/SeverityTag'

export default function ResultList() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [severity, setSeverity] = useState<string | undefined>()
  const [host, setHost] = useState<string | undefined>()
  const [templateId, setTemplateId] = useState<string | undefined>()
  const filters = { severity, host, template_id: templateId }
  const { data, loading, reload } = useResults(page, 20, filters)

  return (
    <Card title={t('result.title')}>
      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          allowClear
          placeholder={t('result.severity')}
          style={{ width: 140 }}
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
        <Input
          allowClear
          placeholder={t('result.host')}
          style={{ width: 200 }}
          value={host}
          onChange={(e) => { setHost(e.target.value || undefined); setPage(1) }}
        />
        <Input
          allowClear
          placeholder={t('result.templateId')}
          style={{ width: 200 }}
          value={templateId}
          onChange={(e) => { setTemplateId(e.target.value || undefined); setPage(1) }}
        />
        <Button icon={<ReloadOutlined />} onClick={reload}>{t('common.refresh')}</Button>
      </Space>

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
          { title: t('result.ip'), dataIndex: 'ip', key: 'ip', width: 130 },
          { title: t('result.port'), dataIndex: 'port', key: 'port', width: 80 },
        ]}
      />
    </Card>
  )
}
