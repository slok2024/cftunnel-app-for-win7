// 内联 SVG 图标组件（Lucide 风格，零依赖）
const Icon = ({ d, size = 18 }: { d: string; size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d={d} />
  </svg>
)

// 仪表盘 - LayoutDashboard
export const IconDashboard = ({ size }: { size?: number }) => (
  <svg width={size||18} height={size||18} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="3" width="7" height="9" rx="1" />
    <rect x="14" y="3" width="7" height="5" rx="1" />
    <rect x="14" y="12" width="7" height="9" rx="1" />
    <rect x="3" y="16" width="7" height="5" rx="1" />
  </svg>
)

// 闪电 - Zap
export const IconZap = ({ size }: { size?: number }) => (
  <Icon d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" size={size} />
)

// 路由 - GitBranch
export const IconRoute = ({ size }: { size?: number }) => (
  <svg width={size||18} height={size||18} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="6" y1="3" x2="6" y2="15" />
    <circle cx="18" cy="6" r="3" />
    <circle cx="6" cy="18" r="3" />
    <path d="M18 9a9 9 0 0 1-9 9" />
  </svg>
)

// 终端 - Terminal
export const IconTerminal = ({ size }: { size?: number }) => (
  <svg width={size||18} height={size||18} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="4 17 10 11 4 5" />
    <line x1="12" y1="19" x2="20" y2="19" />
  </svg>
)

// 警告 - AlertTriangle
export const IconAlert = ({ size }: { size?: number }) => (
  <svg width={size||40} height={size||40} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
    <line x1="12" y1="9" x2="12" y2="13" />
    <line x1="12" y1="17" x2="12.01" y2="17" />
  </svg>
)

// 播放 - Play
export const IconPlay = ({ size }: { size?: number }) => (
  <svg width={size||16} height={size||16} viewBox="0 0 24 24" fill="currentColor" stroke="none">
    <polygon points="5 3 19 12 5 21 5 3" />
  </svg>
)

// 停止 - Square
export const IconStop = ({ size }: { size?: number }) => (
  <svg width={size||16} height={size||16} viewBox="0 0 24 24" fill="currentColor" stroke="none">
    <rect x="4" y="4" width="16" height="16" rx="2" />
  </svg>
)

// 刷新 - RefreshCw
export const IconRefresh = ({ size }: { size?: number }) => (
  <svg width={size||16} height={size||16} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="23 4 23 10 17 10" />
    <polyline points="1 20 1 14 7 14" />
    <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
  </svg>
)

// 加号 - Plus
export const IconPlus = ({ size }: { size?: number }) => (
  <svg width={size||16} height={size||16} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="12" y1="5" x2="12" y2="19" />
    <line x1="5" y1="12" x2="19" y2="12" />
  </svg>
)

// 删除 - Trash2
export const IconTrash = ({ size }: { size?: number }) => (
  <svg width={size||14} height={size||14} viewBox="0 0 24 24" fill="none" stroke="currentColor"
    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="3 6 5 6 21 6" />
    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
  </svg>
)

// 发送 - Send
export const IconSend = ({ size }: { size?: number }) => (
  <Icon d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" size={size || 16} />
)
