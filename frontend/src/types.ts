export type AggregatedState =
  | 'idle'
  | 'running'
  | 'waiting'
  | 'error'
  | 'fault'
  | 'gray';

export const STATE_PRIORITY: Record<AggregatedState, number> = {
  error: 5,
  fault: 4,
  waiting: 3,
  running: 2,
  idle: 1,
  gray: 0,
};

export const STATE_LABEL: Record<AggregatedState, string> = {
  idle: '绿灯常亮 · 空闲',
  running: '黄灯慢闪 · 运行中',
  waiting: '黄灯快闪 · 等待你授权',
  error: '红灯常亮 · 异常',
  fault: '红灯快闪 · API 故障',
  gray: '灰灯 · 全部禁用',
};

export type AgentName =
  | 'claude'
  | 'codex'
  | 'cursor'
  | 'cline'
  | 'aider'
  | 'continue'
  | 'windsurf';

export const AGENT_DISPLAY: Record<AgentName, string> = {
  claude: 'Claude Code',
  codex: 'Codex',
  cursor: 'Cursor',
  cline: 'Cline',
  aider: 'Aider',
  continue: 'Continue',
  windsurf: 'Windsurf',
};

export interface AgentStatus {
  name: AgentName;
  state: AggregatedState;
  enabled: boolean;
  message?: string;
  lastUpdate: number;
}
