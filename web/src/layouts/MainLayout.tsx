import { Layout, Menu } from 'antd'
import {
  DashboardOutlined,
  ScanOutlined,
  FileSearchOutlined,
  CloudServerOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router'
import { useTranslation } from 'react-i18next'
import LanguageSwitch from '../components/LanguageSwitch'

const { Header, Sider, Content } = Layout

export default function MainLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: t('nav.dashboard') },
    { key: '/scans', icon: <ScanOutlined />, label: t('nav.scans') },
    { key: '/results', icon: <FileSearchOutlined />, label: t('nav.results') },
    { key: '/workers', icon: <CloudServerOutlined />, label: t('nav.workers') },
  ]

  const selectedKey = '/' + (location.pathname.split('/')[1] || '')

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible>
        <div style={{ height: 32, margin: 16, color: '#fff', fontWeight: 700, fontSize: 16, textAlign: 'center', whiteSpace: 'nowrap', overflow: 'hidden' }}>
          Nuclei Platform
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
          <LanguageSwitch />
        </Header>
        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
