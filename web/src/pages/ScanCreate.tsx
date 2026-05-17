import { Card, Form, Input, Button, InputNumber, message } from 'antd'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router'
import { useCreateScan } from '../api/hooks'

export default function ScanCreate() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { mutate: createScan, loading } = useCreateScan()
  const [form] = Form.useForm()

  const onFinish = async (values: {
    name: string
    targets: string
    template_filters?: string
    concurrency?: string
    rate_limit?: number
    headers?: string
  }) => {
    const targets = values.targets.split('\n').map((s) => s.trim()).filter(Boolean)
    if (targets.length === 0) {
      message.error(t('scan.targetsRequired'))
      return
    }

    const req: Record<string, unknown> = {
      name: values.name,
      targets,
    }

    if (values.template_filters) {
      try { req.template_filters = JSON.parse(values.template_filters) } catch { /* ignore */ }
    }
    if (values.concurrency) {
      try { req.concurrency = JSON.parse(values.concurrency) } catch { /* ignore */ }
    }
    if (values.rate_limit) req.rate_limit = values.rate_limit
    if (values.headers) {
      req.headers = values.headers.split('\n').map((s) => s.trim()).filter(Boolean)
    }

    await createScan(req as never)
    message.success(t('scan.createSuccess'))
    navigate('/scans')
  }

  return (
    <Card title={t('scan.create')}>
      <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 640 }}>
        <Form.Item name="name" label={t('scan.name')} rules={[{ required: true, message: t('scan.nameRequired') }]}>
          <Input placeholder={t('scan.namePlaceholder')} />
        </Form.Item>

        <Form.Item name="targets" label={t('scan.targets')} rules={[{ required: true, message: t('scan.targetsRequired') }]}>
          <Input.TextArea rows={6} placeholder={t('scan.targetsPlaceholder')} />
        </Form.Item>

        <Form.Item name="template_filters" label={t('scan.templateFilters')}>
          <Input.TextArea rows={3} placeholder={t('scan.templateFiltersPlaceholder')} />
        </Form.Item>

        <Form.Item name="concurrency" label={t('scan.concurrency')}>
          <Input.TextArea rows={2} placeholder={t('scan.concurrencyPlaceholder')} />
        </Form.Item>

        <Form.Item name="rate_limit" label={t('scan.rateLimit')}>
          <InputNumber min={0} placeholder={t('scan.rateLimitPlaceholder')} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item name="headers" label={t('scan.headers')}>
          <Input.TextArea rows={3} placeholder={t('scan.headersPlaceholder')} />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>{t('scan.submit')}</Button>
        </Form.Item>
      </Form>
    </Card>
  )
}
