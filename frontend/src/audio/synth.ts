// Web Audio API 实时合成器, 迁移自 ui/audio-preview.html.
// 零文件, 跨平台一致, 完全代码化.

export type Wave = 'square' | 'sine' | 'triangle' | 'sawtooth';

export type Envelope = 'exp' | 'linear';

let audioCtx: AudioContext | null = null;
let globalVolume = 0.3;
let userMuted = false;

function getCtx(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  if (!audioCtx) {
    const Ctor =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!Ctor) return null;
    audioCtx = new Ctor();
  }
  return audioCtx;
}

export function setGlobalVolume(volume: number): void {
  globalVolume = Math.max(0, Math.min(1, volume));
}

export function getGlobalVolume(): number {
  return globalVolume;
}

export function setMuted(muted: boolean): void {
  userMuted = muted;
}

export function isMuted(): boolean {
  return userMuted;
}

// 浏览器/WebView2 通常要求用户首次交互后才能解锁 AudioContext.
// 在 App 启动后第一次点击/键盘事件里调用一次即可.
export function unlockAudio(): void {
  const ctx = getCtx();
  if (ctx && ctx.state === 'suspended') {
    ctx.resume().catch(() => undefined);
  }
}

export interface SequenceOptions {
  wave?: Wave;
  freqs: number[];
  duration: number; // 单声时长 ms
  gap?: number; // 间隔 ms
  times: number; // 重复次数
  volume?: number; // 0..1, 与全局相乘
  envelope?: Envelope;
}

export function playSequence(opts: SequenceOptions): void {
  if (userMuted) return;
  const ctx = getCtx();
  if (!ctx) return;

  const {
    wave = 'square',
    freqs,
    duration,
    gap = 0,
    times,
    volume = 0.5,
    envelope = 'exp',
  } = opts;

  if (freqs.length === 0 || times <= 0 || duration <= 0) return;

  const finalVol = globalVolume * volume;
  let t = ctx.currentTime;

  for (let i = 0; i < times; i++) {
    const freq = freqs[i % freqs.length];
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = wave;
    osc.frequency.value = freq;

    gain.gain.setValueAtTime(0, t);
    gain.gain.linearRampToValueAtTime(finalVol, t + 0.005);

    if (envelope === 'exp') {
      gain.gain.exponentialRampToValueAtTime(0.001, t + duration / 1000);
    } else {
      gain.gain.setValueAtTime(finalVol, t + duration / 1000 - 0.01);
      gain.gain.linearRampToValueAtTime(0, t + duration / 1000);
    }

    osc.connect(gain).connect(ctx.destination);
    osc.start(t);
    osc.stop(t + duration / 1000);

    t += (duration + gap) / 1000;
  }
}

export interface GlideOptions {
  wave?: Wave;
  freqStart: number;
  freqEnd: number;
  duration: number;
  volume?: number;
}

export function playGlide(opts: GlideOptions): void {
  if (userMuted) return;
  const ctx = getCtx();
  if (!ctx) return;

  const { wave = 'sine', freqStart, freqEnd, duration, volume = 0.5 } = opts;

  const now = ctx.currentTime;
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = wave;
  osc.frequency.setValueAtTime(freqStart, now);
  osc.frequency.linearRampToValueAtTime(freqEnd, now + duration / 1000);

  const finalVol = globalVolume * volume;
  gain.gain.setValueAtTime(0, now);
  gain.gain.linearRampToValueAtTime(finalVol, now + 0.01);
  gain.gain.exponentialRampToValueAtTime(0.001, now + duration / 1000);

  osc.connect(gain).connect(ctx.destination);
  osc.start(now);
  osc.stop(now + duration / 1000);
}
