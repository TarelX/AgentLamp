import type * as React from 'react';
import { type AggregatedState, STATE_LABEL } from '../types';
import './TrafficLight.css';

interface TrafficLightProps {
  state: AggregatedState;
  showLabel?: boolean;
  /** 灯架底部铭牌; 不传则显示默认 'AGENTLAMP' */
  badgeText?: string;
}

export function TrafficLight({ state, showLabel = true, badgeText }: TrafficLightProps) {
  const badgeStyle = badgeText
    ? ({ ['--lamp-label' as string]: `'${badgeText}'` } as React.CSSProperties)
    : undefined;

  const redClasses =
    state === 'error'
      ? 'bulb on red-on'
      : state === 'fault'
      ? 'bulb on red-on fault-blink'
      : 'bulb';

  const yellowClasses =
    state === 'running'
      ? 'bulb on yellow-on running'
      : state === 'waiting'
      ? 'bulb on yellow-on waiting'
      : 'bulb';

  const greenClasses = state === 'idle' ? 'bulb on green-on' : 'bulb';

  return (
    <div className="agent-lamp-stage">
      <div className="traffic-light" style={badgeStyle}>
        <div className={redClasses} />
        <div className={yellowClasses} />
        <div className={greenClasses} />
      </div>
      {showLabel && <div className="lamp-label">{STATE_LABEL[state]}</div>}
    </div>
  );
}
