import { AGENT_BADGE, STATE_PRIORITY, type AgentName, type AggregatedState } from '../types';

interface AgentLike {
  name?: string;
  state?: AggregatedState;
  enabled?: boolean;
}

/**
 * 计算当前最该被关注的 agent 名字 (大写紧凑形式) 用作物理灯铭牌.
 * 规则: 在所有 enabled agent 中取 state 优先级最高且非 idle/gray 的; 平局取 lastUpdate 最近;
 * 若全部 idle/gray, 返回 'AGENTLAMP'.
 */
export function dominantAgentBadge(agents: Record<string, AgentLike> | AgentLike[] | undefined): string {
  if (!agents) return 'AGENTLAMP';
  const list = Array.isArray(agents) ? agents : Object.values(agents);

  let best: AgentLike | null = null;
  let bestPriority = -1;
  for (const a of list) {
    if (!a || !a.enabled) continue;
    const state = (a.state ?? 'gray') as AggregatedState;
    if (state === 'idle' || state === 'gray') continue;
    const p = STATE_PRIORITY[state] ?? 0;
    if (p > bestPriority) {
      bestPriority = p;
      best = a;
    }
  }

  if (!best || !best.name) return 'AGENTLAMP';
  const badge = AGENT_BADGE[best.name as AgentName];
  return badge ?? best.name.toUpperCase();
}
