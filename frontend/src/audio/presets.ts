import { playGlide, playSequence } from './synth';

export type SoundCategory = 'waiting' | 'error' | 'done' | 'real';

export interface SoundPreset {
  id: string;
  category: SoundCategory;
  recommended: boolean;
  title: string;
  desc: string;
  paramSummary: string;
  play: () => void;
}

// 4 类共 20 个 preset, 全部参数对齐 ui/audio-preview.html.
// 推荐项: W1 (waiting), E1 (error), D1 (done), R1 (real).
export const SOUND_PRESETS: readonly SoundPreset[] = [
  // ---------- 1. 等待用户授权 (waiting) - 6 个 ----------
  {
    id: 'W1',
    category: 'waiting',
    recommended: true,
    title: '三声 800Hz 短促 (标准电子嘀)',
    desc: '最经典的电子嘀, 清晰但不刺耳. 中文红绿灯快频参考.',
    paramSummary: 'square · 800Hz · 100ms × 3 · 80ms 间隔',
    play: () =>
      playSequence({ wave: 'square', freqs: [800], duration: 100, gap: 80, times: 3, volume: 0.6 }),
  },
  {
    id: 'W2',
    category: 'waiting',
    recommended: false,
    title: '三声 1000Hz 高频 (更清脆)',
    desc: '比 800 高一档, 嘈杂环境更突出, 但夜里可能太刺耳.',
    paramSummary: 'square · 1000Hz · 100ms × 3 · 80ms',
    play: () =>
      playSequence({ wave: 'square', freqs: [1000], duration: 100, gap: 80, times: 3, volume: 0.6 }),
  },
  {
    id: 'W3',
    category: 'waiting',
    recommended: false,
    title: '五声 600Hz 中频 (柔和)',
    desc: '更低更柔, 办公室友好. 注意重复多, 可能被嫌烦.',
    paramSummary: 'square · 600Hz · 80ms × 5 · 60ms',
    play: () =>
      playSequence({ wave: 'square', freqs: [600], duration: 80, gap: 60, times: 5, volume: 0.5 }),
  },
  {
    id: 'W4',
    category: 'waiting',
    recommended: false,
    title: '"嘟嘀嘀" 摩斯码风',
    desc: '1 长 + 2 短, 有节奏感, 程序员可能爱.',
    paramSummary: 'square · 800Hz · 250ms + 100ms + 100ms',
    play: () => {
      playSequence({ wave: 'square', freqs: [800], duration: 250, gap: 120, times: 1, volume: 0.5 });
      window.setTimeout(
        () =>
          playSequence({ wave: 'square', freqs: [800], duration: 100, gap: 80, times: 2, volume: 0.5 }),
        400
      );
    },
  },
  {
    id: 'W5',
    category: 'waiting',
    recommended: false,
    title: '上升 Skype 来电式',
    desc: '三声向上 800→1000→1200, 温和但醒目, 像 Skype 来电.',
    paramSummary: 'sine · 800/1000/1200 · 120ms × 3',
    play: () =>
      playSequence({
        wave: 'sine',
        freqs: [800, 1000, 1200],
        duration: 120,
        gap: 60,
        times: 3,
        volume: 0.6,
      }),
  },
  {
    id: 'W6',
    category: 'waiting',
    recommended: false,
    title: '双声 1200Hz 圆润',
    desc: 'sine 波双声, 圆润干净, 适合现代极简风.',
    paramSummary: 'sine · 1200Hz · 150ms × 2 · 100ms',
    play: () =>
      playSequence({ wave: 'sine', freqs: [1200], duration: 150, gap: 100, times: 2, volume: 0.6 }),
  },

  // ---------- 2. 错误警报 (error) - 5 个 ----------
  {
    id: 'E1',
    category: 'error',
    recommended: true,
    title: '一声 400Hz 长嘟 (标准低频)',
    desc: '红灯禁行参考音, 沉重克制. 最经典的"出错了"提示.',
    paramSummary: 'square · 400Hz · 350ms × 1',
    play: () =>
      playSequence({ wave: 'square', freqs: [400], duration: 350, gap: 0, times: 1, volume: 0.6 }),
  },
  {
    id: 'E2',
    category: 'error',
    recommended: false,
    title: '双声 350Hz 警报',
    desc: '双声警报, 像车祸提示. 略重, 适合严重 error.',
    paramSummary: 'square · 350Hz · 200ms × 2 · 100ms',
    play: () =>
      playSequence({ wave: 'square', freqs: [350], duration: 200, gap: 100, times: 2, volume: 0.7 }),
  },
  {
    id: 'E3',
    category: 'error',
    recommended: false,
    title: '下降 600→200 滑音 (沮丧感)',
    desc: '频率从 600 滑到 200, 像"哎"的下降语调.',
    paramSummary: 'sine · 600 → 200Hz · 500ms 滑音',
    play: () => playGlide({ wave: 'sine', freqStart: 600, freqEnd: 200, duration: 500, volume: 0.6 }),
  },
  {
    id: 'E4',
    category: 'error',
    recommended: false,
    title: '三声急促 500Hz (红灯禁行)',
    desc: '中频快速三声, 有紧迫感, 适合 API 限流类.',
    paramSummary: 'square · 500Hz · 80ms × 3 · 50ms',
    play: () =>
      playSequence({ wave: 'square', freqs: [500], duration: 80, gap: 50, times: 3, volume: 0.7 }),
  },
  {
    id: 'E5',
    category: 'error',
    recommended: false,
    title: '低频锯齿声 (机械故障)',
    desc: 'sawtooth 波, 粗糙感, 像老式电脑报警.',
    paramSummary: 'sawtooth · 300Hz · 400ms × 1',
    play: () =>
      playSequence({ wave: 'sawtooth', freqs: [300], duration: 400, gap: 0, times: 1, volume: 0.5 }),
  },

  // ---------- 3. 任务完成 (done) - 4 个, 默认全部关 ----------
  {
    id: 'D1',
    category: 'done',
    recommended: true,
    title: '"叮咚" 上行 C-E-G',
    desc: '像 iMessage 收消息, 积极轻快. 默认关, 频繁会烦.',
    paramSummary: 'sine · 523/659/784Hz · 80ms × 3 · 30ms',
    play: () =>
      playSequence({
        wave: 'sine',
        freqs: [523.25, 659.25, 783.99],
        duration: 80,
        gap: 30,
        times: 3,
        volume: 0.5,
      }),
  },
  {
    id: 'D2',
    category: 'done',
    recommended: false,
    title: '一声 1500Hz 清脆铃',
    desc: '最简单的"叮", 像 macOS Glass 提示.',
    paramSummary: 'sine · 1500Hz · 200ms × 1',
    play: () =>
      playSequence({ wave: 'sine', freqs: [1500], duration: 200, gap: 0, times: 1, volume: 0.5 }),
  },
  {
    id: 'D3',
    category: 'done',
    recommended: false,
    title: '木琴风短调',
    desc: '三角波模拟木琴, 像 Material Design 的成功音.',
    paramSummary: 'triangle · 700/900/1100Hz · 60ms × 3',
    play: () =>
      playSequence({
        wave: 'triangle',
        freqs: [700, 900, 1100],
        duration: 60,
        gap: 30,
        times: 3,
        volume: 0.5,
      }),
  },
  {
    id: 'D4',
    category: 'done',
    recommended: false,
    title: '上升 4 音阶 (欢快)',
    desc: 'C-E-G-C 完整三和弦上行, 庆祝感强. 可能过于喜庆.',
    paramSummary: 'sine · 523/659/784/1047Hz',
    play: () =>
      playSequence({
        wave: 'sine',
        freqs: [523.25, 659.25, 783.99, 1046.5],
        duration: 80,
        gap: 25,
        times: 4,
        volume: 0.5,
      }),
  },

  // ---------- 4. 真实红绿灯模拟 (real) - 5 个 ----------
  {
    id: 'R1',
    category: 'real',
    recommended: true,
    title: '中国式过街绿灯 (视障引导)',
    desc: '快频高音, 模拟"嘀嘀嘀嘀嘀嘀嘀"通行音. 推荐取 3 声作 waiting.',
    paramSummary: 'square · 1500Hz · 50ms × 7 · 80ms',
    play: () =>
      playSequence({ wave: 'square', freqs: [1500], duration: 50, gap: 80, times: 7, volume: 0.5 }),
  },
  {
    id: 'R2',
    category: 'real',
    recommended: false,
    title: '中国式过街红灯',
    desc: '低频慢节奏"嘟……嘟……嘟……", 经典禁行提示.',
    paramSummary: 'sine · 300Hz · 300ms × 3 · 400ms',
    play: () =>
      playSequence({ wave: 'sine', freqs: [300], duration: 300, gap: 400, times: 3, volume: 0.5 }),
  },
  {
    id: 'R3',
    category: 'real',
    recommended: false,
    title: '日本式 Cuckoo 杜鹃叫',
    desc: '两声 down-up, 日本部分路口用, 温柔可爱.',
    paramSummary: 'sine · 700/600Hz 滑动 × 2',
    play: () => {
      playGlide({ wave: 'sine', freqStart: 700, freqEnd: 600, duration: 300, volume: 0.5 });
      window.setTimeout(
        () => playGlide({ wave: 'sine', freqStart: 700, freqEnd: 600, duration: 300, volume: 0.5 }),
        500
      );
    },
  },
  {
    id: 'R4',
    category: 'real',
    recommended: false,
    title: '美国式 Chirp 鸟叫',
    desc: '快速上升再下降, 模拟鸟叫, 部分美国路口用.',
    paramSummary: 'sine · 1800 → 2400 → 1800Hz',
    play: () => {
      playGlide({ wave: 'sine', freqStart: 1800, freqEnd: 2400, duration: 80, volume: 0.5 });
      window.setTimeout(
        () => playGlide({ wave: 'sine', freqStart: 2400, freqEnd: 1800, duration: 80, volume: 0.5 }),
        100
      );
    },
  },
  {
    id: 'R5',
    category: 'real',
    recommended: false,
    title: '欧式 Beep-Beep 中频',
    desc: '欧洲常见, 中频两声, 稳重.',
    paramSummary: 'square · 900Hz · 200ms × 2 · 200ms',
    play: () =>
      playSequence({ wave: 'square', freqs: [900], duration: 200, gap: 200, times: 2, volume: 0.6 }),
  },
] as const;

export function findPreset(id: string): SoundPreset | undefined {
  return SOUND_PRESETS.find((p) => p.id === id);
}

export function presetsByCategory(category: SoundCategory): readonly SoundPreset[] {
  return SOUND_PRESETS.filter((p) => p.category === category);
}

// 默认绑定: waiting → W1, error → E1
export const DEFAULT_PRESET_BINDINGS: Record<'waiting' | 'error' | 'done', string | null> = {
  waiting: 'W1',
  error: 'E1',
  done: null,
};
