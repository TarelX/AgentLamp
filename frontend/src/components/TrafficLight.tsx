import { type AggregatedState, STATE_LABEL } from '../types';
import './TrafficLight.css';

interface TrafficLightProps {
  state: AggregatedState;
  showLabel?: boolean;
}

export function TrafficLight({ state, showLabel = true }: TrafficLightProps) {
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
      <div className="traffic-light">
        <div className={redClasses} />
        <div className={yellowClasses} />
        <div className={greenClasses} />
      </div>
      {showLabel && <div className="lamp-label">{STATE_LABEL[state]}</div>}
    </div>
  );
}
