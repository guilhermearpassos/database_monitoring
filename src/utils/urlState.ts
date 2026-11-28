import { dateMath, dateTime, TimeRange } from '@grafana/data';
import { isFetchError } from '@grafana/runtime';

export function resolveTimeRange(range: TimeRange): TimeRange {
  const from =
    typeof range.raw.from === 'string'
      ? (dateMath.parse(range.raw.from, false) ?? range.from)
      : range.from;
  const to =
    typeof range.raw.to === 'string'
      ? (dateMath.parse(range.raw.to, true) ?? range.to)
      : range.to;
  return { from, to, raw: range.raw };
}

export function shiftTimeRange(range: TimeRange, direction: -1 | 1): TimeRange {
  const resolved = resolveTimeRange(range);
  const fromMs = resolved.from.valueOf();
  const toMs = resolved.to.valueOf();
  const halfSpan = Math.max(60_000, Math.round((toMs - fromMs) / 2));
  const nextFrom = dateTime(fromMs + direction * halfSpan);
  const nextTo = dateTime(toMs + direction * halfSpan);
  return {
    from: nextFrom,
    to: nextTo,
    raw: {
      from: nextFrom.toISOString(),
      to: nextTo.toISOString(),
    },
  };
}

export function zoomOutTimeRange(range: TimeRange): TimeRange {
  const resolved = resolveTimeRange(range);
  const fromMs = resolved.from.valueOf();
  const toMs = resolved.to.valueOf();
  const span = Math.max(60_000, toMs - fromMs);
  const center = (fromMs + toMs) / 2;
  const nextFrom = dateTime(Math.round(center - span));
  const nextTo = dateTime(Math.round(center + span));
  return {
    from: nextFrom,
    to: nextTo,
    raw: {
      from: nextFrom.toISOString(),
      to: nextTo.toISOString(),
    },
  };
}

export function extractErrorMessage(err: unknown, fallback: string): string {
  if (isFetchError(err)) {
    const data = err.data;
    if (typeof data === 'string' && data.trim()) {
      return data.trim();
    }
    if (data && typeof data === 'object') {
      const msg = (data as any).message || (data as any).error?.message || (data as any).error;
      if (typeof msg === 'string' && msg.trim()) {
        return msg.trim();
      }
    }
    if (err.statusText) {
      return `${err.status ? `${err.status} ` : ''}${err.statusText}`;
    }
  }
  if (err instanceof Error && err.message) {
    return err.message;
  }
  if (typeof err === 'string' && err.trim()) {
    return err.trim();
  }
  return fallback;
}

export function parseTimeRangeFromSearchParams(params: URLSearchParams): TimeRange {
  const rawFrom = params.get('from') || 'now-1h';
  const rawTo = params.get('to') || 'now';

  const parsedFrom = dateMath.parse(rawFrom, false) ?? dateTime().subtract(1, 'hour');
  const parsedTo = dateMath.parse(rawTo, true) ?? dateTime();

  return {
    from: parsedFrom,
    to: parsedTo,
    raw: {
      from: rawFrom,
      to: rawTo,
    },
  };
}

export function serializeTimeRangePoint(raw: string | ReturnType<typeof dateTime>, fallback: ReturnType<typeof dateTime>): string {
  if (typeof raw === 'string' && raw.length > 0) {
    return raw;
  }
  return fallback.toISOString();
}

export function applyTimeRangeToSearchParams(params: URLSearchParams, range: TimeRange): void {
  params.set('from', serializeTimeRangePoint(range.raw.from, range.from));
  params.set('to', serializeTimeRangePoint(range.raw.to, range.to));
}

export function parseCSVString(raw: string | null | undefined): string[] {
  if (!raw) {
    return [];
  }
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
}

export function parseCSVParam(params: URLSearchParams, key: string): string[] {
  return parseCSVString(params.get(key));
}

export function setCSVParam(params: URLSearchParams, key: string, values: string[]): void {
  if (values.length === 0) {
    params.delete(key);
  } else {
    params.set(key, values.join(','));
  }
}

export function setOptionalParam(params: URLSearchParams, key: string, value: string | null | undefined): void {
  if (!value) {
    params.delete(key);
  } else {
    params.set(key, value);
  }
}

export function buildPreservedNavSearch(params: URLSearchParams, overrides?: Record<string, string | null | undefined>): string {
  const next = new URLSearchParams();
  for (const key of ['server', 'from', 'to']) {
    const val = params.get(key);
    if (val) {
      next.set(key, val);
    }
  }
  if (overrides) {
    for (const [k, v] of Object.entries(overrides)) {
      if (v) {
        next.set(k, v);
      } else {
        next.delete(k);
      }
    }
  }
  const s = next.toString();
  return s ? `?${s}` : '';
}
