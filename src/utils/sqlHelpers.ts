import { format } from 'sql-formatter';
import { MissingIndexWarning } from '../components/ExecutionPlanTree/plan';

export interface IssueTagMetadata {
  label: string;
  shortLabel: string;
  severity: 'error' | 'warning' | 'info';
  description: string;
}

export const ISSUE_TAG_META: Record<string, IssueTagMetadata> = {
  seek_plan_convert: {
    label: 'Implicit Conversion (Seek Plan)',
    shortLabel: 'Seek Plan Convert',
    severity: 'error',
    description: 'Type mismatch prevents SQL Server from performing an index seek, forcing a scan.',
  },
  missing_index: {
    label: 'Missing Index',
    shortLabel: 'Missing Index',
    severity: 'warning',
    description: 'Query optimizer recommends creating an index to reduce query cost.',
  },
  cardinality_convert: {
    label: 'Implicit Conversion (Cardinality)',
    shortLabel: 'Cardinality Convert',
    severity: 'warning',
    description: 'Implicit conversion affects optimizer cardinality estimates.',
  },
  implicit_conversion: {
    label: 'Implicit Conversion',
    shortLabel: 'Implicit Convert',
    severity: 'warning',
    description: 'Plan-affecting implicit type conversion detected.',
  },
  large_table_scan: {
    label: 'Large Table Scan',
    shortLabel: 'Table Scan',
    severity: 'error',
    description: 'Full table scan over a large number of rows.',
  },
};

export function getIssueTagMeta(tag: string): IssueTagMetadata {
  return (
    ISSUE_TAG_META[tag] ?? {
      label: tag,
      shortLabel: tag,
      severity: 'info',
      description: tag,
    }
  );
}

export function formatSQLPlain(sql: string): string {
  if (!sql) {
    return '';
  }
  const trimmed = sql.trim();
  try {
    return format(trimmed, {
      language: 'tsql',
      tabWidth: 2,
      keywordCase: 'upper',
      linesBetweenQueries: 2,
      expressionWidth: 100,
    });
  } catch {
    return sql;
  }
}

export function formatDurationMs(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) {
    return '0 ms';
  }
  if (ms < 1000) {
    return `${Math.round(ms)} ms`;
  }
  const seconds = ms / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(seconds < 10 ? 2 : 1)} s`;
  }
  const totalSec = Math.round(seconds);
  const mins = Math.floor(totalSec / 60);
  const remSec = totalSec % 60;
  return `${mins}m ${remSec}s`;
}

export function buildMissingIndexDDL(mi: MissingIndexWarning): string {
  const keyCols = [...(mi.equality_columns ?? []), ...(mi.inequality_columns ?? [])];
  const includeCols = mi.include_columns ?? [];
  const cleanTable = mi.table.replace(/[\[\]]/g, '');
  const idxName = `IX_${cleanTable}_${keyCols
    .slice(0, 3)
    .map((c) => c.replace(/[\[\]]/g, ''))
    .join('_')}`;
  const target = [mi.database, mi.schema, mi.table].filter(Boolean).join('.');
  const base = `CREATE NONCLUSTERED INDEX [${idxName}] ON ${target} (${keyCols.join(', ')})`;
  if (includeCols.length > 0) {
    return `${base} INCLUDE (${includeCols.join(', ')});`;
  }
  return `${base};`;
}

export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // fall through
  }
  try {
    const el = document.createElement('textarea');
    el.value = text;
    el.style.position = 'fixed';
    el.style.opacity = '0';
    document.body.appendChild(el);
    el.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(el);
    return ok;
  } catch {
    return false;
  }
}

export function downloadTextFile(filename: string, content: string, mimeType = 'text/plain;charset=utf-8'): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}
