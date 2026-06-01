/** Web Audio API 实时合成: 零文件, 跨平台一致 */

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

/** 设置全局音量 (0..1), 与各 preset 自带音量相乘 */
export function setGlobalVolume(volume: number): void {
  globalVolume = Math.max(0, Math.min(1, volume));
}

export function getGlobalVolume(): number {
  return globalVolume;
}

/** 全局静音开关; 静音时 playSequence/playGlide 无副作用 */
export function setMuted(muted: boolean): void {
  userMuted = muted;
}

export function isMuted(): boolean {
  return userMuted;
}

/** 浏览器与 WebView2 要求用户手势后才能播放音频, 启动时调一次即可 */
export function unlockAudio(): void {
  const ctx = getCtx();
  if (ctx && ctx.state === 'suspended') {
    ctx.resume().catch(() => undefined);
  }
}

export interface SequenceOptions {
  wave?: Wave;
  freqs: number[];
  /** 单声时长, 毫秒 */
  duration: number;
  /** 两声之间的静音, 毫秒 */
  gap?: number;
  times: number;
  /** 0..1, 与全局音量相乘 */
  volume?: number;
  envelope?: Envelope;
}

/** 顺序播放一组等参数音; freqs 长度小于 times 时循环取用 */
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

/** 单声滑音, 频率从 freqStart 线性滑到 freqEnd */
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
