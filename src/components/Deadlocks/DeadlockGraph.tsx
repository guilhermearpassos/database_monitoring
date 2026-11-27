import React, { useMemo, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2, useTheme2 } from '@grafana/ui';
import { DeadlockItem } from './types';

const getStyles = (theme: GrafanaTheme2) => ({
  wrapper: css`
    background: ${theme.colors.background.canvas};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.5)};
    overflow-x: auto;
  `,
  hintRow: css`
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(1)};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
  svg: css`
    display: block;
    width: 100%;
    min-width: 720px;
    height: auto;
  `,
  nodeGroup: css`
    cursor: pointer;
  `,
});

interface NodeBox {
  x: number;
  y: number;
  w: number;
  h: number;
}

interface GraphEdge {
  key: string;
  kind: 'owns' | 'waits';
  fromId: string;
  toId: string;
  label: string;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
}

function truncate(s: string, max: number): string {
  const clean = (s || '').replace(/\s+/g, ' ').trim();
  if (clean.length <= max) {
    return clean;
  }
  return `${clean.slice(0, max - 1)}…`;
}

export const DeadlockGraph: React.FC<{ deadlock: DeadlockItem }> = ({ deadlock }) => {
  const styles = useStyles2(getStyles);
  const theme = useTheme2();
  const [highlightedNodeId, setHighlightedNodeId] = useState<string | null>(null);

  const layout = useMemo(() => {
    const procs = deadlock.processes ?? [];
    const resources = deadlock.resources ?? [];
    const rows = Math.max(procs.length, resources.length, 1);
    const nodeH = 112;
    const gapY = 28;
    const padY = 24;
    const svgH = Math.max(250, padY * 2 + rows * nodeH + (rows - 1) * gapY);
    const svgW = 900;

    const procW = 285;
    const resW = 265;
    const procX = 20;
    const resX = svgW - resW - 20;

    const procPos: Record<string, NodeBox> = {};
    const totalProcH = procs.length * nodeH + Math.max(0, procs.length - 1) * gapY;
    const procStartY = (svgH - totalProcH) / 2;
    procs.forEach((p, idx) => {
      procPos[p.id] = { x: procX, y: procStartY + idx * (nodeH + gapY), w: procW, h: nodeH };
    });

    const resPos: Record<string, NodeBox> = {};
    const totalResH = resources.length * nodeH + Math.max(0, resources.length - 1) * gapY;
    const resStartY = (svgH - totalResH) / 2;
    resources.forEach((r, idx) => {
      resPos[r.id] = { x: resX, y: resStartY + idx * (nodeH + gapY), w: resW, h: nodeH };
    });

    const edges: GraphEdge[] = [];
    resources.forEach((r) => {
      const rp = resPos[r.id];
      (r.owners ?? []).forEach((o, idx) => {
        const pp = procPos[o.process_id];
        if (!pp || !rp) {
          return;
        }
        edges.push({
          key: `owns-${r.id}-${o.process_id}-${idx}`,
          kind: 'owns',
          fromId: r.id,
          toId: o.process_id,
          label: `OWNS ${o.mode}`,
          x1: rp.x,
          y1: rp.y + rp.h * 0.34 + idx * 14,
          x2: pp.x + pp.w,
          y2: pp.y + pp.h * 0.34 + idx * 14,
        });
      });
      (r.waiters ?? []).forEach((w, idx) => {
        const pp = procPos[w.process_id];
        if (!pp || !rp) {
          return;
        }
        edges.push({
          key: `waits-${w.process_id}-${r.id}-${idx}`,
          kind: 'waits',
          fromId: w.process_id,
          toId: r.id,
          label: w.request_type === 'convert' ? `CONVERT -> ${w.mode}` : `WAITS ${w.mode}`,
          x1: pp.x + pp.w,
          y1: pp.y + pp.h * 0.68 + idx * 14,
          x2: rp.x,
          y2: rp.y + rp.h * 0.68 + idx * 14,
        });
      });
    });

    return { svgW, svgH, procPos, resPos, edges, procs, resources };
  }, [deadlock]);

  const ownsColor = theme.colors.success.text;
  const waitsColor = theme.colors.error.text;
  const primaryText = theme.colors.text.primary;
  const secondaryText = theme.colors.text.secondary;
  const bgSurface = theme.colors.background.primary;

  const toggleHighlight = (id: string) => {
    setHighlightedNodeId((prev) => (prev === id ? null : id));
  };

  return (
    <div className={styles.wrapper}>
      <div className={styles.hintRow}>
        <span>
          <strong>Interactive Wait-For Graph</strong> — click any Process or Resource node to isolate its lock edges
        </span>
        <span style={{ fontFamily: theme.typography.fontFamilyMonospace, fontSize: 11 }}>
          <span style={{ color: ownsColor }}>━━▶ OWNS lock</span>
          {'   '}
          <span style={{ color: waitsColor }}>╌╌▶ WAITS / CONVERTS</span>
        </span>
      </div>

      <svg className={styles.svg} viewBox={`0 0 ${layout.svgW} ${layout.svgH}`} role="img" aria-label="Deadlock wait-for graph">
        <defs>
          <marker id="dl-arrow-owns" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill={ownsColor} />
          </marker>
          <marker id="dl-arrow-waits" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill={waitsColor} />
          </marker>
        </defs>

        {layout.edges.map((e, idx) => {
          const focused = !highlightedNodeId || e.fromId === highlightedNodeId || e.toId === highlightedNodeId;
          const color = e.kind === 'owns' ? ownsColor : waitsColor;
          const curveOffset = e.kind === 'owns' ? -18 - idx * 3 : 18 + idx * 3;
          const midX = (e.x1 + e.x2) / 2;
          const midY = (e.y1 + e.y2) / 2 + curveOffset;
          const labelY = (e.y1 + e.y2) / 2 + curveOffset * 0.52;
          const pillW = Math.max(78, e.label.length * 7.2 + 14);

          return (
            <g key={e.key} opacity={focused ? 1 : 0.2}>
              <path
                d={`M ${e.x1} ${e.y1} Q ${midX} ${midY} ${e.x2} ${e.y2}`}
                fill="none"
                stroke={color}
                strokeWidth={2}
                strokeDasharray={e.kind === 'waits' ? '6,4' : undefined}
                markerEnd={e.kind === 'owns' ? 'url(#dl-arrow-owns)' : 'url(#dl-arrow-waits)'}
              />
              <rect
                x={midX - pillW / 2}
                y={labelY - 10}
                width={pillW}
                height={20}
                rx={10}
                fill={bgSurface}
                stroke={color}
                strokeWidth={1}
              />
              <text
                x={midX}
                y={labelY + 4}
                textAnchor="middle"
                fill={color}
                fontSize={10.5}
                fontFamily={theme.typography.fontFamilyMonospace}
                fontWeight={700}
              >
                {e.label}
              </text>
            </g>
          );
        })}

        {layout.procs.map((p) => {
          const pos = layout.procPos[p.id];
          if (!pos) {
            return null;
          }
          const selected = highlightedNodeId === p.id;
          const stroke = p.is_victim
            ? theme.colors.error.main
            : selected
              ? theme.colors.primary.text
              : theme.colors.primary.border;
          const fill = p.is_victim ? theme.colors.error.transparent : theme.colors.primary.transparent;

          return (
            <g
              key={p.id}
              className={styles.nodeGroup}
              data-testid={`deadlock-node-${p.id}`}
              onClick={() => toggleHighlight(p.id)}
            >
              <rect
                x={pos.x}
                y={pos.y}
                width={pos.w}
                height={pos.h}
                rx={6}
                fill={fill}
                stroke={stroke}
                strokeWidth={selected ? 2.5 : 1.5}
              />
              <text
                x={pos.x + 12}
                y={pos.y + 22}
                fill={primaryText}
                fontSize={13}
                fontWeight={700}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                SPID {p.spid}
              </text>
              {p.is_victim ? (
                <>
                  <rect x={pos.x + pos.w - 72} y={pos.y + 9} width={60} height={19} rx={3} fill={theme.colors.error.main} />
                  <text x={pos.x + pos.w - 42} y={pos.y + 22} textAnchor="middle" fill="#fff" fontSize={10} fontWeight={700}>
                    VICTIM
                  </text>
                </>
              ) : (
                <>
                  <rect x={pos.x + pos.w - 82} y={pos.y + 9} width={70} height={19} rx={3} fill={theme.colors.primary.shade} />
                  <text x={pos.x + pos.w - 47} y={pos.y + 22} textAnchor="middle" fill="#fff" fontSize={10} fontWeight={600}>
                    SURVIVOR
                  </text>
                </>
              )}
              <text x={pos.x + 12} y={pos.y + 42} fill={primaryText} fontSize={11.5}>
                {truncate(`${p.client_app} · ${p.login_name}`, 36)}
              </text>
              <text
                x={pos.x + 12}
                y={pos.y + 60}
                fill={secondaryText}
                fontSize={11}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                Wait: {p.wait_time_ms}ms · Log: {p.log_used.toLocaleString()}B
              </text>
              <text x={pos.x + 12} y={pos.y + 77} fill={secondaryText} fontSize={10.5}>
                {truncate(p.isolation_level, 36)}
              </text>
              <text
                x={pos.x + 12}
                y={pos.y + 98}
                fill={theme.colors.info.text}
                fontSize={10.5}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                {truncate(p.active_sql, 38)}
              </text>
            </g>
          );
        })}

        {layout.resources.map((r) => {
          const pos = layout.resPos[r.id];
          if (!pos) {
            return null;
          }
          const selected = highlightedNodeId === r.id;
          const typeLabel = `${r.type.replace(/lock$/i, '').toUpperCase()} LOCK`;
          const detailLine = r.index_name
            ? `Index: ${r.index_name}`
            : r.page_id
              ? `File ${r.file_id} : Page ${r.page_id}`
              : `HoBT: ${r.hobt_id || '-'}`;

          return (
            <g
              key={r.id}
              className={styles.nodeGroup}
              data-testid={`deadlock-node-${r.id}`}
              onClick={() => toggleHighlight(r.id)}
            >
              <rect
                x={pos.x}
                y={pos.y}
                width={pos.w}
                height={pos.h}
                rx={6}
                fill={theme.colors.warning.transparent}
                stroke={selected ? theme.colors.warning.text : theme.colors.warning.border}
                strokeWidth={selected ? 2.5 : 1.5}
              />
              <rect
                x={pos.x + 12}
                y={pos.y + 10}
                width={Math.max(76, typeLabel.length * 7 + 12)}
                height={18}
                rx={3}
                fill={theme.colors.warning.transparent}
                stroke={theme.colors.warning.border}
              />
              <text
                x={pos.x + 18}
                y={pos.y + 23}
                fill={theme.colors.warning.text}
                fontSize={10}
                fontWeight={700}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                {typeLabel}
              </text>
              <text
                x={pos.x + 12}
                y={pos.y + 48}
                fill={primaryText}
                fontSize={12.5}
                fontWeight={700}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                {truncate(r.short_object || r.object_name, 30)}
              </text>
              <text
                x={pos.x + 12}
                y={pos.y + 68}
                fill={primaryText}
                fontSize={11}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                {truncate(detailLine, 34)}
              </text>
              <text
                x={pos.x + 12}
                y={pos.y + 88}
                fill={secondaryText}
                fontSize={10.5}
                fontFamily={theme.typography.fontFamilyMonospace}
              >
                DB: {r.database || r.dbid || '-'} · Mode: {r.mode || '-'}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
};
