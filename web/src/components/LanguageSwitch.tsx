import { Button } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

export default function LanguageSwitch() {
  const { i18n } = useTranslation()

  const toggle = () => {
    const next = i18n.language === 'zh' ? 'en' : 'zh'
    i18n.changeLanguage(next)
    localStorage.setItem('lang', next)
  }

  return (
    <Button type="text" icon={<GlobalOutlined />} onClick={toggle}>
      {i18n.language === 'zh' ? '中文' : 'EN'}
    </Button>
  )
}
