import { useEffect, useRef } from 'react';
import { TrafficLight } from './components/TrafficLight';
import { useStatusStore } from './stores/useStatusStore';
import { setGlobalVolume, unlockAudio } from './audio/synth';
import { findPreset, DEFAULT_PRESET_BINDINGS } from './audio/presets';
import type { AggregatedState } from './types';

const DEMO_STATES: AggregatedState[] = ['idle', 'running', 'waiting', 'error', 'fault', 'gray'];

const STATE_BTN_LABEL: Record<AggregatedState, string> = {
  idle: '空闲',
  running: '运行中',
  waiting: '等待',
  error: '错误',
  fault: '故障',
  gray: '禁用',
};

function App() {
  const mainState = useStatusStore((s) => s.mainState);
  const setMainStateDemo = useStatusStore((s) => s.setMainStateDemo);
  const globalVolume = useStatusStore((s) => s.globalVolume);
  const setVolume = useStatusStore((s) => s.setVolume);
  const muted = useStatusStore((s) => s.muted);
  const toggleMuted = useStatusStore((s) => s.toggleMuted);

  const prevState = useRef<AggregatedState>(mainState);

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

  return (
    <div className="app-root" onClick={unlockAudio}>
      <header className="app-header">
        <h1>AgentLamp</h1>
        <p className="tagline">程序员的过街信号</p>
      </header>

      <TrafficLight state={mainState} />

      <section className="controls">
        <span className="controls-label">模拟状态</span>
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
        <button
          type="button"
          className={`btn ${muted ? 'active' : ''}`}
          onClick={toggleMuted}
        >
          {muted ? '已静音' : '静音'}
        </button>
      </section>

      <footer className="app-footer">
        <span>v0.1 · MIT · </span>
        <a href="https://github.com/TarelX/AgentLamp" target="_blank" rel="noreferrer">
          github.com/TarelX/AgentLamp
        </a>
      </footer>
    </div>
  );
}

export default App;
