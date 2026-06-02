import { useEffect, useRef, useState } from 'react';
import { Events } from '@wailsio/runtime';
import { TrafficLight } from './components/TrafficLight';
import { findPreset, DEFAULT_PRESET_BINDINGS } from './audio/presets';
import { unlockAudio } from './audio/synth';
import { dominantAgentBadge } from './hooks/useDominantAgent';
import type { AggregatedState } from './types';
import * as StatusService from '../bindings/github.com/TarelX/AgentLamp/backend/service/statusservice';
import * as WindowService from '../bindings/github.com/TarelX/AgentLamp/backend/service/windowservice';

/** 悬浮模式: 透明置顶可拖动小灯, 双击切回完整窗口 */
function Lamp() {
  const [state, setState] = useState<AggregatedState>('idle');
  const [badge, setBadge] = useState<string>('AGENTLAMP');
  const prev = useRef<AggregatedState>('idle');

  useEffect(() => {
    let cancelled = false;
    const apply = (raw: unknown) => {
      if (!raw || typeof raw !== 'object') return;
      const snap = raw as { mainState?: AggregatedState; agents?: unknown };
      if (snap.mainState) setState(snap.mainState);
      setBadge(dominantAgentBadge(snap.agents as never));
    };

    StatusService.GetSnapshot()
      .then((snap) => {
        if (cancelled) return;
        apply(snap);
      })
      .catch(() => undefined);
    const off = Events.On('status:update', (evt: { data?: unknown }) => {
      if (cancelled) return;
      apply(evt.data);
    });
    return () => {
      cancelled = true;
      if (typeof off === 'function') off();
    };
  }, []);

  useEffect(() => {
    if (prev.current !== state) {
      if (state === 'waiting' && DEFAULT_PRESET_BINDINGS.waiting) {
        findPreset(DEFAULT_PRESET_BINDINGS.waiting)?.play();
      } else if (
        (state === 'error' || state === 'fault') &&
        DEFAULT_PRESET_BINDINGS.error
      ) {
        findPreset(DEFAULT_PRESET_BINDINGS.error)?.play();
      }
      prev.current = state;
    }
  }, [state]);

  return (
    <div
      className="lamp-window"
      onClick={unlockAudio}
      onDoubleClick={() => void WindowService.SwitchToFull()}
      title="双击切回主窗口"
    >
      <TrafficLight state={state} showLabel={false} badgeText={badge} />
    </div>
  );
}

export default Lamp;
