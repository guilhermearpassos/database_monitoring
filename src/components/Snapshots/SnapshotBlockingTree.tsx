import React, { useMemo } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { useStyles2 } from '@grafana/ui';
import { Clock, Lock } from 'lucide-react';
import { SnapshotSampleItem } from './types';
import { formatDurationMs } from '../../utils/sqlHelpers';
import { TempoTraceLink } from '../TempoTraceLink';

interface TreeNode {
  sample: SnapshotSampleItem;
  children: TreeNode[];
  level: number;
}

const getStyles = (theme: GrafanaTheme2) => ({
  wrapper: css`
    display: flex;
    flex-direction: column;
    gap: ${theme.spacing(1.5)};
    padding: ${theme.spacing(1.5)};
  `,
  chainCard: css`
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.primary};
    padding: ${theme.spacing(1.25)};
  `,
  nodeRow: css`
    border-left: 2px solid ${theme.colors.border.medium};
    padding-left: ${theme.spacing(1.5)};
    margin-top: ${theme.spacing(0.75)};
  `,
  nodeBox: css`
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: ${theme.spacing(1, 1.25)};
    border-radius: ${theme.shape.radius.default};
    border: 1px solid ${theme.colors.border.weak};
    background: ${theme.colors.background.secondary};
    cursor: pointer;

    &:hover {
      border-color: ${theme.colors.primary.border};
      background: ${theme.colors.action.hover};
    }
  `,
  headBlockerBox: css`
    border-left: 3px solid ${theme.colors.error.main};
  `,
  nodeHeader: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
  leftMeta: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  blockerTag: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 1px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border: 1px solid ${theme.colors.error.border};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  waiterTag: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 1px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border: 1px solid ${theme.colors.warning.border};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  sqlLine: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 12px;
    color: ${theme.colors.text.secondary};
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  `,
  empty: css`
    padding: ${theme.spacing(3)};
    text-align: center;
    color: ${theme.colors.text.secondary};
  `,
});

export function buildBlockingTrees(samples: SnapshotSampleItem[]): TreeNode[] {
  const bySession = new Map<string, SnapshotSampleItem>();
  for (const s of samples) {
    if (!bySession.has(s.session_id)) {
      bySession.set(s.session_id, s);
    }
  }

  // Find root blockers: sessions that block others and are not themselves blocked (or whose blocker is missing)
  const rootSamples = samples.filter(
    (s) => (s.is_blocker && (!s.is_blocked || !bySession.has(s.blocked_by))) || (s.is_blocked && !bySession.has(s.blocked_by))
  );

  const buildNode = (sample: SnapshotSampleItem, level: number, visited: Set<string>): TreeNode => {
    visited.add(sample.session_id);
    const children: TreeNode[] = [];
    for (const childSid of sample.blocked_sessions ?? []) {
      if (visited.has(childSid)) {
        continue;
      }
      const childSample = bySession.get(childSid);
      if (childSample) {
        children.push(buildNode(childSample, level + 1, visited));
      }
    }
    return { sample, children, level };
  };

  const visited = new Set<string>();
  const roots: TreeNode[] = [];
  for (const root of rootSamples) {
    if (visited.has(root.session_id)) {
      continue;
    }
    roots.push(buildNode(root, 0, visited));
  }
  // Include any remaining unvisited blocking/blocked sessions (e.g., mutual deadlock cycles A <-> B).
  for (const s of samples) {
    if ((s.is_blocker || s.is_blocked) && !visited.has(s.session_id)) {
      roots.push(buildNode(s, 0, visited));
    }
  }
  return roots;
}

const TreeNodeView: React.FC<{
  node: TreeNode;
  timeRange: TimeRange;
  onSelectSample: (snapId: string, sampleId: string) => void;
}> = ({ node, timeRange, onSelectSample }) => {
  const styles = useStyles2(getStyles);
  const { sample } = node;

  return (
    <div className={node.level > 0 ? styles.nodeRow : undefined} style={{ marginLeft: node.level > 0 ? 12 : 0 }}>
      <div
        className={`${styles.nodeBox} ${node.level === 0 && sample.is_blocker ? styles.headBlockerBox : ''}`}
        onClick={() => onSelectSample(sample.snap_id, sample.sample_id)}
      >
        <div className={styles.nodeHeader}>
          <div className={styles.leftMeta}>
            <strong>SID {sample.session_id}</strong>
            {sample.is_blocker && (
              <span className={styles.blockerTag}>
                <Lock size={11} />
                {node.level === 0 ? 'Head Blocker' : 'Blocker'} ({sample.blocked_sessions.length} waiting)
              </span>
            )}
            {sample.is_blocked && (
              <span className={styles.waiterTag}>
                <Clock size={11} />
                Blocked by {sample.blocked_by} ({formatDurationMs(sample.wait_time_ms)})
              </span>
            )}
            <span>{sample.database || '-'}</span>
            {sample.wait_event && <span style={{ opacity: 0.8 }}>{sample.wait_event}</span>}
            {sample.trace_id && <TempoTraceLink traceId={sample.trace_id} timeRange={timeRange} compact />}
          </div>
          <div className={styles.leftMeta}>
            <span>Elapsed: {formatDurationMs(sample.elapsed_ms)}</span>
            <span>User: {sample.user || '-'}</span>
            <span>Status: {sample.status}</span>
          </div>
        </div>
        <div className={styles.sqlLine} title={sample.text}>
          {sample.text}
        </div>
      </div>

      {node.children.map((child) => (
        <TreeNodeView
          key={`${child.sample.session_id}-${child.sample.sample_id}`}
          node={child}
          timeRange={timeRange}
          onSelectSample={onSelectSample}
        />
      ))}
    </div>
  );
};

export const SnapshotBlockingTree: React.FC<{
  samples: SnapshotSampleItem[];
  timeRange: TimeRange;
  onSelectSample: (snapId: string, sampleId: string) => void;
}> = ({ samples, timeRange, onSelectSample }) => {
  const styles = useStyles2(getStyles);
  const roots = useMemo(() => buildBlockingTrees(samples), [samples]);

  if (roots.length === 0) {
    return <div className={styles.empty}>No blocking chains detected in this snapshot.</div>;
  }

  return (
    <div className={styles.wrapper}>
      {roots.map((root) => (
        <div key={root.sample.session_id} className={styles.chainCard}>
          <TreeNodeView node={root} timeRange={timeRange} onSelectSample={onSelectSample} />
        </div>
      ))}
    </div>
  );
};
