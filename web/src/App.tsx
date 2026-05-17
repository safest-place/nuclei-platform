import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import MainLayout from './layouts/MainLayout'
import Dashboard from './pages/Dashboard'
import ScanList from './pages/ScanList'
import ScanCreate from './pages/ScanCreate'
import ScanDetail from './pages/ScanDetail'
import ResultList from './pages/ResultList'
import WorkerList from './pages/WorkerList'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<MainLayout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/scans" element={<ScanList />} />
          <Route path="/scans/create" element={<ScanCreate />} />
          <Route path="/scans/:id" element={<ScanDetail />} />
          <Route path="/results" element={<ResultList />} />
          <Route path="/workers" element={<WorkerList />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
