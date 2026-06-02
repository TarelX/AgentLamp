import { useEffect, useRef, useState } from 'react';
import { Events } from '@wailsio/runtime';
import { TrafficLight } from './components/TrafficLight';
import { useStatusStore } from './stores/useStatusStore';
import { setGlobalVolume, unlockAudio } from './audio/synth';
import { findPreset, DEFAULT_PRESET_BINDINGS } from './audio/presets';
import { type AggregatedState, type AgentName, AGENT_DISPLAY, STATE_LABEL } from './types';
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

function App() {
  const mainState = useStatusStore((s) => s.mainState);
  const setMainStateDemo = useStatusStore((s) => s.setMainStateDemo);
  const globalVolume = useStatusStore((s) => s.globalVolume);
  const setVolume = useStatusStore((s) => s.setVolume);
  const muted = useStatusStore((s) => s.muted);
  const toggleMuted = useStatusStore((s) => s.toggleMuted);

  const prevState = useRef<AggregatedState>(mainState);
  const [runtimeAgents, setRuntimeAgents] = useState<RuntimeAgent[]>([]);
  const [backendConnected, setBackendConnected] = useState(false);
  const [installState, setInstallState] = useState<Record<'claude' | 'cursor', InstallSlot>>({
    claude: { installed: null, path: '', busy: false, msg: '' },
    cursor: { installed: null, path: '', busy: false, msg: '' },
  });

  const refreshInstallStatus = async () => {
    const [claudeSt, cursorSt] = await Promise.allSettled([
      InstallService.ClaudeStatus(),
      InstallService.CursorStatus(),
    ]);
    setInstallState((prev) => ({
      claude: applySlot(prev.claude, claudeSt),
      cursor: applySlot(prev.cursor, cursorSt),
    }));
  };

  const toggleInstall = async (agent: 'claude' | 'cursor') => {
    setInstallState((prev) => ({ ...prev, [agent]: { ...prev[agent], busy: true, msg: '' } }));
    const cur = installState[agent];
    const fn = installFn(agent, cur.installed === true);
    const restartHint =
      agent === 'claude' ? '重启 Claude Code 让 hook 生效' : '重启 Cursor 让 hook 生效';
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

  useEffect(() => {
    setGlobalVolume(globalVolume);
  }, [globalVolume]);

  // 仅在状态变化的边沿播放, 防止持续 waiting / error 期间重复响铃
  useEffect(() => {
    if (muted) return;
    if (prevState.current !== mainState) {
      if (mainState === 'waiting' && DEFAULT_PRESET_BINDINGS.waiting) {
        findPreset(DEFAULT_PRESET_BINDINGS.waiting)?.play();
      } else if (
        (mainState === 'error' || mainState === 'fault') &&
        DEFAULT_PRESET_BINDINGS.error
      ) {
        findPreset(DEFAULT_PRESET_BINDINGS.error)?.play();
      }
      prevState.current = mainState;
    }
  }, [mainState, muted]);

  // 启动时拉一次快照, 并订阅后续 status:update 事件; 失败不阻塞 UI demo 模式
  useEffect(() => {
    let cancelled = false;
    StatusService.GetSnapshot()
      .then((snap) => {
        if (cancelled) return;
        setBackendConnected(true);
        applySnapshot(snap, setMainStateDemo, setRuntimeAgents);
        void refreshInstallStatus();
      })
      .catch(() => {
        // 浏览器开发模式下 (无 Wails runtime) 静默忽略, 走纯 demo 模式
      });
    const off = Events.On('status:update', (evt: { data?: unknown }) => {
      if (cancelled) return;
      setBackendConnected(true);
      applySnapshot(evt.data, setMainStateDemo, setRuntimeAgents);
    });
    return () => {
      cancelled = true;
      if (typeof off === 'function') off();
    };
  }, [setMainStateDemo]);

  return (
    <div className="app-root" onClick={unlockAudio}>
      <header className="app-header">
        <h1>AgentLamp</h1>
        <p className="tagline">程序员的过街信号 · {STATE_LABEL[mainState]}</p>
      </header>

      <TrafficLight state={mainState} showLabel={false} />

      {runtimeAgents.length > 0 && (
        <section className="agents-row">
          {runtimeAgents.map((a) => (
            <span key={a.name} className={`agent-chip ${a.enabled ? '' : 'disabled'}`}>
              <span
                className="agent-dot"
                style={{ background: STATE_DOT_COLOR[a.state] }}
              />
              {AGENT_DISPLAY[a.name]} · {STATE_BTN_LABEL[a.state]}
            </span>
          ))}
        </section>
      )}

      <section className="controls">
        <span className="controls-label">{backendConnected ? '手动测试' : 'Demo 模式'}</span>
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
      </section>

      <section className="audio-row">
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
      </section>

      {(['claude', 'cursor'] as const).map((agent) => {
        const slot = installState[agent];
        if (slot.installed === null) return null;
        const display = agent === 'claude' ? 'Claude hook' : 'Cursor hook';
        return (
          <section key={agent} className="hooks-row">
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
          </section>
        );
      })}

      <footer className="app-footer">
        <span>v0.1 · MIT · </span>
        <a href="https://github.com/TarelX/AgentLamp" target="_blank" rel="noreferrer">
          github.com/TarelX/AgentLamp
        </a>
        <span> · webhook 127.0.0.1:19840</span>
      </footer>
    </div>
  );
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

function installFn(agent: 'claude' | 'cursor', isInstalled: boolean) {
  if (agent === 'claude') {
    return isInstalled ? InstallService.ClaudeUninstall : InstallService.ClaudeInstall;
  }
  return isInstalled ? InstallService.CursorUninstall : InstallService.CursorInstall;
}

function applySnapshot(
  raw: unknown,
  setMain: (s: AggregatedState) => void,
  setAgents: (xs: RuntimeAgent[]) => void
) {
  if (!raw || typeof raw !== 'object') return;
  const snap = raw as {
    mainState?: AggregatedState;
    agents?: Record<string, { name?: string; state?: AggregatedState; enabled?: boolean; lastUpdate?: number }>;
  };
  if (snap.mainState) {
    setMain(snap.mainState);
  }
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
}

export default App;
