import { dateMath, dateTime, TimeRange } from '@grafana/data';

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

export function parseCSVParam(params: URLSearchParams, key: string): string[] {
  const raw = params.get(key);
  if (!raw) {
    return [];
  }
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
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
