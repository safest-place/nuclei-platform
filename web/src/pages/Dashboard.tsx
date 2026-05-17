import { Card, Col, Row, Statistic, Table, Spin } from 'antd'
import {
  BugOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useStats, useScans } from '../api/hooks'
import StatusTag from '../components/StatusTag'

export default function Dashboard() {
  const { t } = useTranslation()
  const { data: stats, loading: statsLoading } = useStats()
  const { data: scans, loading: scansLoading } = useScans(1, 10)

  const severityData = stats?.data.by_severity || []
  const hostData = stats?.data.by_host || []
  const recentScans = scans?.data || []

  const totalScans = scans?.total ?? 0
  const runningCount = recentScans.filter((s) => s.status === 'running').length
  const completedCount = recentScans.filter((s) => s.status === 'completed').length
  const failedCount = recentScans.filter((s) => s.status === 'failed').length

  return (
    <Spin spinning={statsLoading && scansLoading}>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic title={t('dashboard.totalScans')} value={totalScans} prefix={<BugOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={t('dashboard.running')} value={runningCount} prefix={<SyncOutlined spin />} valueStyle={{ color: '#1677ff' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={t('dashboard.completed')} value={completedCount} prefix={<CheckCircleOutlined />} valueStyle={{ color: '#389e0d' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={t('dashboard.failed')} value={failedCount} prefix={<CloseCircleOutlined />} valueStyle={{ color: '#cf1322' }} />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={12}>
          <Card title={t('dashboard.severityDistribution')}>
            <Table
              dataSource={severityData}
              rowKey="severity"
              pagination={false}
              size="small"
              columns={[
                { title: t('result.severity'), dataIndex: 'severity', key: 'severity' },
                { title: 'Count', dataIndex: 'count', key: 'count' },
              ]}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={t('dashboard.topHosts')}>
            <Table
              dataSource={hostData}
              rowKey="host"
              pagination={false}
              size="small"
              columns={[
                { title: t('result.host'), dataIndex: 'host', key: 'host', ellipsis: true },
                { title: 'Count', dataIndex: 'count', key: 'count' },
              ]}
            />
          </Card>
        </Col>
      </Row>

      <Card title={t('dashboard.recentScans')}>
        <Table
          dataSource={recentScans}
          rowKey="id"
          loading={scansLoading}
          pagination={false}
          size="small"
          columns={[
            { title: t('scan.name'), dataIndex: 'name', key: 'name' },
            {
              title: t('scan.status'),
              dataIndex: 'status',
              key: 'status',
              render: (s: string) => <StatusTag status={s} />,
            },
            { title: t('scan.resultCount'), dataIndex: 'result_count', key: 'result_count' },
            {
              title: t('scan.createdAt'),
              dataIndex: 'created_at',
              key: 'created_at',
              render: (v: string) => new Date(v).toLocaleString(),
            },
          ]}
        />
      </Card>
    </Spin>
  )
}
