import { create } from 'zustand';
import { type AgentName, type AgentStatus, type AggregatedState, STATE_PRIORITY } from '../types';

const ALL_AGENTS: readonly AgentName[] = [
  'claude',
  'codex',
  'cursor',
  'cline',
  'aider',
  'continue',
  'windsurf',
] as const;

const V0_1_DEFAULT_ENABLED: readonly AgentName[] = ['claude', 'cursor'] as const;

function buildInitialAgents(): Record<AgentName, AgentStatus> {
  const now = Date.now();
  return ALL_AGENTS.reduce<Record<AgentName, AgentStatus>>((acc, name) => {
    const enabled = V0_1_DEFAULT_ENABLED.includes(name);
    acc[name] = {
      name,
      state: enabled ? 'idle' : 'gray',
      enabled,
      lastUpdate: now,
    };
    return acc;
  }, {} as Record<AgentName, AgentStatus>);
}

function computeMainState(agents: Record<AgentName, AgentStatus>): AggregatedState {
  const enabledList = Object.values(agents).filter((a) => a.enabled);
  if (enabledList.length === 0) return 'gray';
  let bestPriority = -1;
  let bestState: AggregatedState = 'gray';
  for (const a of enabledList) {
    const p = STATE_PRIORITY[a.state];
    if (p > bestPriority) {
      bestPriority = p;
      bestState = a.state;
    }
  }
  return bestState;
}

interface StatusState {
  agents: Record<AgentName, AgentStatus>;
  mainState: AggregatedState;

  globalVolume: number;
  muted: boolean;
  followSystemDND: boolean;

  setAgentState: (name: AgentName, state: AggregatedState, message?: string) => void;
  toggleAgent: (name: AgentName, enabled: boolean) => void;

  setVolume: (volume: number) => void;
  toggleMuted: () => void;
  setFollowSystemDND: (v: boolean) => void;

  setMainStateDemo: (state: AggregatedState) => void;
  resetDemo: () => void;
}

const initialAgents = buildInitialAgents();

export const useStatusStore = create<StatusState>()((set) => ({
  agents: initialAgents,
  mainState: computeMainState(initialAgents),

  globalVolume: 0.3,
  muted: false,
  followSystemDND: true,

  setAgentState: (name, state, message) =>
    set((s) => {
      const agents = {
        ...s.agents,
        [name]: { ...s.agents[name], state, message, lastUpdate: Date.now() },
      };
      return { agents, mainState: computeMainState(agents) };
    }),

  toggleAgent: (name, enabled) =>
    set((s) => {
      const agents = {
        ...s.agents,
        [name]: { ...s.agents[name], enabled, lastUpdate: Date.now() },
      };
      return { agents, mainState: computeMainState(agents) };
    }),

  setVolume: (volume) => set({ globalVolume: Math.max(0, Math.min(1, volume)) }),
  toggleMuted: () => set((s) => ({ muted: !s.muted })),
  setFollowSystemDND: (v) => set({ followSystemDND: v }),

  setMainStateDemo: (state) => set({ mainState: state }),
  resetDemo: () => {
    const fresh = buildInitialAgents();
    set({ agents: fresh, mainState: computeMainState(fresh) });
  },
}));
