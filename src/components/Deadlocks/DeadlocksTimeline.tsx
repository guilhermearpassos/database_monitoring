import React, { useMemo } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { Card, Tooltip, useStyles2, useTheme2 } from '@grafana/ui';
import { DeadlockItem } from './types';

const getStyles = (theme: GrafanaTheme2) => ({
  legend: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(2)};
    flex-wrap: wrap;
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
  legendDot: css`
    display: inline-block;
    width: 9px;
    height: 9px;
    border-radius: 2px;
    margin-right: 5px;
    vertical-align: middle;
  `,
  barsGrid: css`
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 6px;
    align-items: end;
    height: 76px;
    border-bottom: 1px solid ${theme.colors.border.weak};
    padding-bottom: 4px;
  `,
  barCol: css`
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    height: 100%;
  `,
  barStack: css`
    width: 100%;
    max-width: 34px;
    display: flex;
    flex-direction: column-reverse;
    gap: 1px;
  `,
  barSeg: css`
    border-radius: 2px;
    min-height: 4px;
  `,
  labelsGrid: css`
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 6px;
    margin-top: 6px;
    text-align: center;
    font-size: 10px;
    color: ${theme.colors.text.secondary};
    font-family: ${theme.typography.fontFamilyMonospace};
  `,
});

interface Bucket {
  label: string;
  reverse_order: number;
  bookmark_lookup: number;
  lock_upgrade: number;
  other: number;
}

export const DeadlocksTimeline: React.FC<{
  deadlocks: DeadlockItem[];
  timeRange: TimeRange;
}> = ({ deadlocks, timeRange }) => {
  const styles = useStyles2(getStyles);
  const theme = useTheme2();

  const { buckets, maxTotal } = useMemo(() => {
    const count = 12;
    const fromMs = timeRange.from.valueOf();
    const toMs = timeRange.to.valueOf();
    const spanMs = Math.max(60_000, toMs - fromMs);
    const stepMs = spanMs / count;

    const multiDay = spanMs > 86_400_000;
    const list: Bucket[] = [];
    for (let i = 0; i < count; i++) {
      const d = new Date(fromMs + i * stepMs);
      const hhmm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
      const label = multiDay
        ? `${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')} ${hhmm}`
        : hhmm;
      list.push({ label, reverse_order: 0, bookmark_lookup: 0, lock_upgrade: 0, other: 0 });
    }

    for (const dl of deadlocks) {
      const ts = new Date(dl.timestamp).getTime();
      if (!Number.isFinite(ts) || ts < fromMs || ts > toMs) {
        continue;
      }
      const idx = Math.min(count - 1, Math.floor((ts - fromMs) / stepMs));
      const p = dl.diagnosis?.pattern || '';
      if (p === 'reverse_order') {
        list[idx].reverse_order++;
      } else if (p === 'bookmark_lookup') {
        list[idx].bookmark_lookup++;
      } else if (p === 'lock_upgrade') {
        list[idx].lock_upgrade++;
      } else {
        list[idx].other++;
      }
    }

    const peak = Math.max(
      1,
      ...list.map((b) => b.reverse_order + b.bookmark_lookup + b.lock_upgrade + b.other)
    );
    return { buckets: list, maxTotal: peak };
  }, [deadlocks, timeRange]);

  const colors = {
    reverse_order: theme.colors.error.main,
    bookmark_lookup: theme.colors.warning.main,
    lock_upgrade: '#a352cc',
    other: theme.colors.primary.main,
  };
  const scale = 60 / maxTotal;

  return (
    <Card>
      <Card.Heading>Deadlocks Over Time</Card.Heading>
      <Card.Actions>
        <div className={styles.legend}>
          <span>
            <span className={styles.legendDot} style={{ background: colors.reverse_order }} />
            Reverse-Order Cycle
          </span>
          <span>
            <span className={styles.legendDot} style={{ background: colors.bookmark_lookup }} />
            Bookmark Lookup
          </span>
          <span>
            <span className={styles.legendDot} style={{ background: colors.lock_upgrade }} />
            Lock Upgrade (S→X)
          </span>
          <span>
            <span className={styles.legendDot} style={{ background: colors.other }} />
            Parallelism / Other
          </span>
        </div>
      </Card.Actions>
      <Card.Description>
        <div className={styles.barsGrid}>
          {buckets.map((b, idx) => {
            const total = b.reverse_order + b.bookmark_lookup + b.lock_upgrade + b.other;
            return (
              <div key={idx} className={styles.barCol}>
                {total > 0 && (
                  <Tooltip
                    content={`${b.label}: ${total} deadlock(s) (Reverse: ${b.reverse_order}, Lookup: ${b.bookmark_lookup}, Upgrade: ${b.lock_upgrade}, Other: ${b.other})`}
                  >
                    <div className={styles.barStack}>
                      {b.reverse_order > 0 && (
                        <div
                          className={styles.barSeg}
                          style={{ height: b.reverse_order * scale, background: colors.reverse_order }}
                        />
                      )}
                      {b.bookmark_lookup > 0 && (
                        <div
                          className={styles.barSeg}
                          style={{ height: b.bookmark_lookup * scale, background: colors.bookmark_lookup }}
                        />
                      )}
                      {b.lock_upgrade > 0 && (
                        <div
                          className={styles.barSeg}
                          style={{ height: b.lock_upgrade * scale, background: colors.lock_upgrade }}
                        />
                      )}
                      {b.other > 0 && (
                        <div className={styles.barSeg} style={{ height: b.other * scale, background: colors.other }} />
                      )}
                    </div>
                  </Tooltip>
                )}
              </div>
            );
          })}
        </div>
        <div className={styles.labelsGrid}>
          {buckets.map((b, idx) => (
            <div key={idx}>{b.label}</div>
          ))}
        </div>
      </Card.Description>
    </Card>
  );
};
