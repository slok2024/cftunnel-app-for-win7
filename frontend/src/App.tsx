import { useState, useEffect, useCallback } from 'react'
import './style.css'
import { CheckInstall, GetStatus, GetRoutes, TunnelUp, TunnelDown, RunCommand } from '../wailsjs/go/main/App'

type Route = { name: string; hostname: string; service: string }
type Page = 'dashboard' | 'quick' | 'routes' | 'terminal'

function App() {
  const [page, setPage] = useState<Page>('dashboard')
  const [version, setVersion] = useState('')
  const [installed, setInstalled] = useState(false)
  const [status, setStatus] = useState('')
  const [routes, setRoutes] = useState<Route[]>([])
  const [loading, setLoading] = useState(false)

  const refresh = useCallback(async () => {
    const info = await CheckInstall()
    setInstalled(info.installed)
    setVersion(info.version || '')
    if (info.installed) {
      setStatus(await GetStatus())
      setRoutes(await GetRoutes() || [])
    }
  }, [])

  useEffect(() => { refresh() }, [refresh])

  const isRunning = status.includes('运行中')

  return (
    <>
      <div className="titlebar">cftunnel</div>
      <div className="app">
        <Sidebar page={page} setPage={setPage} version={version} />
        <div className="main">
          {!installed ? <NotInstalled /> :
            page === 'dashboard' ? <Dashboard status={status} isRunning={isRunning} routes={routes} loading={loading} setLoading={setLoading} refresh={refresh} /> :
            page === 'quick' ? <QuickMode /> :
            page === 'routes' ? <Routes routes={routes} refresh={refresh} /> :
            <Terminal />}
        </div>
      </div>
    </>
  )
}

function Sidebar({ page, setPage, version }: { page: Page; setPage: (p: Page) => void; version: string }) {
  const items: { id: Page; icon: string; label: string }[] = [
    { id: 'dashboard', icon: '📊', label: '仪表盘' },
    { id: 'quick', icon: '⚡', label: '免域名模式' },
    { id: 'routes', icon: '🔗', label: '路由管理' },
    { id: 'terminal', icon: '💻', label: '终端' },
  ]
  return (
    <div className="sidebar">
      <div className="sidebar-header">cf<span>tunnel</span></div>
      <div className="sidebar-nav">
        {items.map(i => (
          <button key={i.id} className={`nav-item${page === i.id ? ' active' : ''}`} onClick={() => setPage(i.id)}>
            {i.icon} {i.label}
          </button>
        ))}
      </div>
      <div className="sidebar-footer">{version || 'cftunnel'}</div>
    </div>
  )
}

function NotInstalled() {
  return (
    <div className="empty">
      <div className="empty-icon">⚠️</div>
      <p>未检测到 cftunnel CLI</p>
      <p style={{ marginTop: 8, fontSize: 13, color: 'var(--text2)' }}>
        请先安装: curl -fsSL https://raw.githubusercontent.com/qingchencloud/cftunnel/main/install.sh | bash
      </p>
    </div>
  )
}

function Dashboard({ status, isRunning, routes, loading, setLoading, refresh }: {
  status: string; isRunning: boolean; routes: Route[]; loading: boolean; setLoading: (b: boolean) => void; refresh: () => Promise<void>
}) {
  const handleUp = async () => { setLoading(true); await TunnelUp(); await refresh(); setLoading(false) }
  const handleDown = async () => { setLoading(true); await TunnelDown(); await refresh(); setLoading(false) }

  return (
    <>
      <div className="page-title">仪表盘</div>
      <div className="card">
        <div className="card-title">隧道状态</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
          <span className={`status-dot ${isRunning ? 'running' : 'stopped'}`} />
          <span>{isRunning ? '运行中' : '已停止'}</span>
        </div>
        <div className="terminal" style={{ marginBottom: 16 }}>{status}</div>
        <div className="btn-group">
          <button className="btn btn-primary" onClick={handleUp} disabled={loading || isRunning}>
            {loading ? <span className="spinner" /> : null} 启动
          </button>
          <button className="btn btn-danger" onClick={handleDown} disabled={loading || !isRunning}>
            停止
          </button>
          <button className="btn btn-outline" onClick={refresh}>刷新</button>
        </div>
      </div>
      <div className="card">
        <div className="card-title">路由 ({routes.length})</div>
        {routes.length === 0 ? <div style={{ color: 'var(--text2)', fontSize: 14 }}>暂无路由</div> : (
          <table className="route-table">
            <thead><tr><th>名称</th><th>域名</th><th>服务</th></tr></thead>
            <tbody>{routes.map(r => <tr key={r.name}><td>{r.name}</td><td>{r.hostname}</td><td>{r.service}</td></tr>)}</tbody>
          </table>
        )}
      </div>
    </>
  )
}

function QuickMode() {
  const [port, setPort] = useState('3000')
  const [output, setOutput] = useState('')
  const [running, setRunning] = useState(false)

  const start = async () => {
    setRunning(true)
    setOutput('正在启动免域名隧道...\n')
    const result = await RunCommand(`quick ${port}`)
    setOutput(result)
    setRunning(false)
  }

  return (
    <>
      <div className="page-title">免域名模式</div>
      <div className="card">
        <div className="card-title">快速启动</div>
        <p style={{ fontSize: 14, color: 'var(--text2)', marginBottom: 16 }}>
          零配置生成 *.trycloudflare.com 临时公网地址
        </p>
        <div className="input-row" style={{ marginBottom: 16 }}>
          <input className="input" style={{ width: 120 }} value={port} onChange={e => setPort(e.target.value)} placeholder="端口" />
          <button className="btn btn-primary" onClick={start} disabled={running}>
            {running ? <span className="spinner" /> : null} 启动隧道
          </button>
        </div>
        {output && <div className="terminal">{output}</div>}
      </div>
    </>
  )
}

function Routes({ routes, refresh }: { routes: Route[]; refresh: () => Promise<void> }) {
  const [name, setName] = useState('')
  const [port, setPort] = useState('')
  const [domain, setDomain] = useState('')
  const [output, setOutput] = useState('')

  const addRoute = async () => {
    if (!name || !port || !domain) return
    const result = await RunCommand(`add ${name} ${port} --domain ${domain}`)
    setOutput(result)
    await refresh()
    setName(''); setPort(''); setDomain('')
  }

  const removeRoute = async (n: string) => {
    const result = await RunCommand(`remove ${n}`)
    setOutput(result)
    await refresh()
  }

  return (
    <>
      <div className="page-title">路由管理</div>
      <div className="card">
        <div className="card-title">添加路由</div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 12 }}>
          <input className="input" style={{ width: 120 }} value={name} onChange={e => setName(e.target.value)} placeholder="名称" />
          <input className="input" style={{ width: 80 }} value={port} onChange={e => setPort(e.target.value)} placeholder="端口" />
          <input className="input" style={{ flex: 1, minWidth: 200 }} value={domain} onChange={e => setDomain(e.target.value)} placeholder="域名 (如 app.example.com)" />
          <button className="btn btn-primary" onClick={addRoute}>添加</button>
        </div>
        {output && <div className="terminal">{output}</div>}
      </div>
      <div className="card">
        <div className="card-title">当前路由 ({routes.length})</div>
        {routes.length === 0 ? <div style={{ color: 'var(--text2)', fontSize: 14 }}>暂无路由</div> : (
          <table className="route-table">
            <thead><tr><th>名称</th><th>域名</th><th>服务</th><th>操作</th></tr></thead>
            <tbody>{routes.map(r => (
              <tr key={r.name}>
                <td>{r.name}</td><td>{r.hostname}</td><td>{r.service}</td>
                <td><button className="btn btn-danger" style={{ padding: '4px 12px', fontSize: 12 }} onClick={() => removeRoute(r.name)}>删除</button></td>
              </tr>
            ))}</tbody>
          </table>
        )}
      </div>
    </>
  )
}

function Terminal() {
  const [cmd, setCmd] = useState('')
  const [output, setOutput] = useState('欢迎使用 cftunnel 终端\n输入命令（不需要 cftunnel 前缀）\n')

  const run = async () => {
    if (!cmd.trim()) return
    setOutput(prev => prev + `\n$ cftunnel ${cmd}\n`)
    const result = await RunCommand(cmd)
    setOutput(prev => prev + result + '\n')
    setCmd('')
  }

  return (
    <>
      <div className="page-title">终端</div>
      <div className="card">
        <div className="terminal" style={{ maxHeight: 400, marginBottom: 12 }}>{output}</div>
        <div className="input-row">
          <span style={{ color: 'var(--green)', fontFamily: 'monospace' }}>$</span>
          <input className="input" value={cmd} onChange={e => setCmd(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && run()} placeholder="status / list / up / down ..." />
          <button className="btn btn-primary" onClick={run}>执行</button>
        </div>
      </div>
    </>
  )
}

export default App
