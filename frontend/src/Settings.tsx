import { useEffect, useState } from 'react';
import { Events } from '@wailsio/runtime';
import { useStatusStore } from './stores/useStatusStore';
import { findPreset } from './audio/presets';
import { setGlobalVolume, unlockAudio } from './audio/synth';
import {
  type AggregatedState,
  type AgentName,
  AGENT_DISPLAY,
  STATE_LABEL,
} from './types';
import * as StatusService from '../bindings/github.com/TarelX/AgentLamp/backend/service/statusservice';
import * as InstallService from '../bindings/github.com/TarelX/AgentLamp/backend/service/installservice';

const DEMO_STATES: AggregatedState[] = ['idle', 'running', 'waiting', 'error', 'fault', 'gray'];

const STATE_BTN_LABEL: Record<AggregatedState, string> = {
  idle: '空闲',
  running: '运行中',
  waiting: '等待',
  error: '错误',
  fault: '故障',
  gray: '禁用',
};

const STATE_DOT_COLOR: Record<AggregatedState, string> = {
  idle: '#22c55e',
  running: '#f59e0b',
  waiting: '#f59e0b',
  error: '#ef4444',
  fault: '#ef4444',
  gray: '#6b7280',
};

interface RuntimeAgent {
  name: AgentName;
  state: AggregatedState;
  enabled: boolean;
  lastUpdate: number;
}

interface InstallSlot {
  installed: boolean | null;
  path: string;
  busy: boolean;
  msg: string;
}

interface RawInstallStatus {
  installed?: boolean;
  settingsPath?: string;
}

const SUPPORTED_INSTALL_AGENTS = ['claude', 'cursor'] as const;
type InstallAgent = (typeof SUPPORTED_INSTALL_AGENTS)[number];

function Settings() {
  const mainState = useStatusStore((s) => s.mainState);
  const setMainStateDemo = useStatusStore((s) => s.setMainStateDemo);
  const globalVolume = useStatusStore((s) => s.globalVolume);
  const setVolume = useStatusStore((s) => s.setVolume);
  const muted = useStatusStore((s) => s.muted);
  const toggleMuted = useStatusStore((s) => s.toggleMuted);

  const [agents, setAgents] = useState<RuntimeAgent[]>([]);
  const [backendConnected, setBackendConnected] = useState(false);
  const [installState, setInstallState] = useState<Record<InstallAgent, InstallSlot>>({
    claude: { installed: null, path: '', busy: false, msg: '' },
    cursor: { installed: null, path: '', busy: false, msg: '' },
  });

  useEffect(() => {
    setGlobalVolume(globalVolume);
  }, [globalVolume]);

  useEffect(() => {
    let cancelled = false;
    const apply = (raw: unknown) => {
      if (!raw || typeof raw !== 'object') return;
      const snap = raw as {
        mainState?: AggregatedState;
        agents?: Record<string, { name?: string; state?: AggregatedState; enabled?: boolean; lastUpdate?: number }>;
      };
      if (snap.mainState) setMainStateDemo(snap.mainState);
      if (snap.agents) {
        const list: RuntimeAgent[] = [];
        for (const key of Object.keys(snap.agents)) {
          const a = snap.agents[key];
          list.push({
            name: (a?.name as AgentName) ?? (key as AgentName),
            state: (a?.state as AggregatedState) ?? 'gray',
            enabled: !!a?.enabled,
            lastUpdate: a?.lastUpdate ?? 0,
          });
        }
        setAgents(list);
      }
    };

    StatusService.GetSnapshot()
      .then((snap) => {
        if (cancelled) return;
        setBackendConnected(true);
        apply(snap);
        void refreshInstall();
      })
      .catch(() => undefined);
    const off = Events.On('status:update', (evt: { data?: unknown }) => {
      if (cancelled) return;
      setBackendConnected(true);
      apply(evt.data);
    });
    return () => {
      cancelled = true;
      if (typeof off === 'function') off();
    };
  }, [setMainStateDemo]);

  const refreshInstall = async () => {
    const [c, k] = await Promise.allSettled([
      InstallService.ClaudeStatus(),
      InstallService.CursorStatus(),
    ]);
    setInstallState((prev) => ({
      claude: applySlot(prev.claude, c),
      cursor: applySlot(prev.cursor, k),
    }));
  };

  const toggleInstall = async (agent: InstallAgent) => {
    setInstallState((prev) => ({ ...prev, [agent]: { ...prev[agent], busy: true, msg: '' } }));
    const cur = installState[agent];
    const fn = installFn(agent, cur.installed === true);
    const restartHint = agent === 'claude' ? '重启 Claude Code 让 hook 生效' : '重启 Cursor 让 hook 生效';
    try {
      const st = await fn();
      setInstallState((prev) => ({
        ...prev,
        [agent]: {
          installed: !!st.installed,
          path: st.settingsPath ?? '',
          busy: false,
          msg: st.installed ? `已写入, ${restartHint}` : '已卸载',
        },
      }));
    } catch (e) {
      setInstallState((prev) => ({
        ...prev,
        [agent]: {
          ...prev[agent],
          busy: false,
          msg: `失败: ${e instanceof Error ? e.message : String(e)}`,
        },
      }));
    }
  };

  return (
    <div className="settings-window" onClick={unlockAudio}>
      <header className="settings-header">
        <h1>AgentLamp · 设置</h1>
        <p className="tagline">
          当前主灯: <strong>{STATE_LABEL[mainState]}</strong>
          {!backendConnected && <span className="warn"> · 后端未连接 (Demo 模式)</span>}
        </p>
      </header>

      <section className="settings-section">
        <h2>Agent 状态</h2>
        {agents.length === 0 ? (
          <p className="hint">暂无后端推送, 启动 webhook 服务并安装对应 hook 后会自动出现.</p>
        ) : (
          <div className="agents-row">
            {agents.map((a) => (
              <span key={a.name} className={`agent-chip ${a.enabled ? '' : 'disabled'}`}>
                <span className="agent-dot" style={{ background: STATE_DOT_COLOR[a.state] }} />
                {AGENT_DISPLAY[a.name]} · {STATE_BTN_LABEL[a.state]}
              </span>
            ))}
          </div>
        )}
      </section>

      <section className="settings-section">
        <h2>Hook 安装</h2>
        {SUPPORTED_INSTALL_AGENTS.map((agent) => {
          const slot = installState[agent];
          const display = agent === 'claude' ? 'Claude Code' : 'Cursor';
          if (slot.installed === null) {
            return (
              <div key={agent} className="hooks-row">
                <span className="hooks-label">
                  {display}: <strong>检测中…</strong>
                </span>
              </div>
            );
          }
          return (
            <div key={agent} className="hooks-row">
              <span className="hooks-label">
                {display}: <strong>{slot.installed ? '已安装' : '未安装'}</strong>
              </span>
              <button
                type="button"
                className={`btn ${slot.installed ? '' : 'active'}`}
                onClick={() => void toggleInstall(agent)}
                disabled={slot.busy}
              >
                {slot.busy ? '处理中…' : slot.installed ? '卸载' : '一键安装'}
              </button>
              {slot.msg && <span className="hooks-msg">{slot.msg}</span>}
              {slot.path && (
                <span className="hooks-path" title={slot.path}>
                  {slot.path}
                </span>
              )}
            </div>
          );
        })}
      </section>

      <section className="settings-section">
        <h2>音效</h2>
        <div className="audio-row">
          <span className="audio-label">音量 {Math.round(globalVolume * 100)}%</span>
          <input
            type="range"
            min={0}
            max={100}
            value={Math.round(globalVolume * 100)}
            onChange={(e) => setVolume(Number(e.target.value) / 100)}
          />
          <button type="button" className="btn" onClick={() => findPreset('W1')?.play()}>
            试听 W1
          </button>
          <button type="button" className="btn" onClick={() => findPreset('E1')?.play()}>
            试听 E1
          </button>
          <button type="button" className={`btn ${muted ? 'active' : ''}`} onClick={toggleMuted}>
            {muted ? '已静音' : '静音'}
          </button>
        </div>
      </section>

      <section className="settings-section">
        <h2>调试</h2>
        <div className="controls">
          <span className="controls-label">手动模拟主灯状态</span>
          {DEMO_STATES.map((s) => (
            <button
              key={s}
              type="button"
              className={`btn ${mainState === s ? 'active' : ''}`}
              onClick={() => setMainStateDemo(s)}
            >
              {STATE_BTN_LABEL[s]}
            </button>
          ))}
        </div>
      </section>

      <footer className="settings-footer">
        <span>AgentLamp v0.1 · MIT · </span>
        <a href="https://github.com/TarelX/AgentLamp" target="_blank" rel="noreferrer">
          github.com/TarelX/AgentLamp
        </a>
        <span> · webhook 127.0.0.1:19840</span>
      </footer>
    </div>
  );
}

function applySlot(prev: InstallSlot, settled: PromiseSettledResult<RawInstallStatus>): InstallSlot {
  if (settled.status !== 'fulfilled') {
    return { ...prev, installed: null };
  }
  return {
    ...prev,
    installed: !!settled.value.installed,
    path: settled.value.settingsPath ?? '',
  };
}

function installFn(agent: InstallAgent, isInstalled: boolean) {
  if (agent === 'claude') {
    return isInstalled ? InstallService.ClaudeUninstall : InstallService.ClaudeInstall;
  }
  return isInstalled ? InstallService.CursorUninstall : InstallService.CursorInstall;
}

export default Settings;
