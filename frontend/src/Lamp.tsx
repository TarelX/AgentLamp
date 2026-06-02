import { useEffect, useRef, useState } from 'react';
import { Events } from '@wailsio/runtime';
import { TrafficLight } from './components/TrafficLight';
import { findPreset, DEFAULT_PRESET_BINDINGS } from './audio/presets';
import { unlockAudio } from './audio/synth';
import type { AggregatedState } from './types';
import * as StatusService from '../bindings/github.com/TarelX/AgentLamp/backend/service/statusservice';
import * as WindowService from '../bindings/github.com/TarelX/AgentLamp/backend/service/windowservice';

/** 悬浮模式: 透明置顶可拖动小灯, 双击切回完整窗口 */
function Lamp() {
  const [state, setState] = useState<AggregatedState>('idle');
  const prev = useRef<AggregatedState>('idle');

  useEffect(() => {
    let cancelled = false;
    StatusService.GetSnapshot()
      .then((snap) => {
        if (cancelled) return;
        if (snap.mainState) setState(snap.mainState as AggregatedState);
      })
      .catch(() => undefined);
    const off = Events.On('status:update', (evt: { data?: unknown }) => {
      if (cancelled) return;
      const raw = evt.data as { mainState?: AggregatedState } | undefined;
      if (raw?.mainState) setState(raw.mainState);
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
      <TrafficLight state={state} showLabel={false} />
    </div>
  );
}

export default Lamp;
