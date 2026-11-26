import React, { useState } from 'react';
import { ChevronDown, ChevronRight, AlertCircle } from 'lucide-react';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { useStyles2 } from '@grafana/ui';

// Matches pkg/plugin/types.go JSON serialization (snake_case)
export interface PlanNodeHeader {
  physical_op?: string;
  logical_op?: string;
  estimate_cpu?: number;
  estimate_io?: number;
  estimate_rows?: number;
  estimated_cost?: number;
  parallel?: string;
}

export interface PlanNode {
  name: string;
  estimated_rows: number;
  subtree_cost: number;
  node_cost: number;
  header?: PlanNodeHeader;
  nodes: PlanNode[];
}

export interface ExpandSignal {
  expanded: boolean;
  version: number;
}

const RISKY_OPS = new Set([
  'Table Scan',
  'Clustered Index Scan',
  'Key Lookup',
  'RID Lookup',
  'Sort',
  'Table Spool',
  'Index Spool',
]);

export const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    background: ${theme.colors.background.primary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(2)};
    max-width: 100%;
    overflow: auto;
  `,
  header: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(2)};
  `,
  headerLeft: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
  `,
  toolbar: css`
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
  `,
  title: css`
    font-size: ${theme.typography.h4.fontSize};
    font-weight: ${theme.typography.h4.fontWeight};
    color: ${theme.colors.text.primary};
    margin: 0;
  `,
  section: css`
    margin-bottom: ${theme.spacing(2.5)};
  `,
  sectionTitle: css`
    font-size: ${theme.typography.h5.fontSize};
    font-weight: ${theme.typography.h5.fontWeight};
    color: ${theme.colors.text.primary};
    margin-bottom: ${theme.spacing(1.5)};
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
  `,
  xmlPre: css`
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(2)};
    font-size: ${theme.typography.bodySmall.fontSize};
    font-family: ${theme.typography.fontFamilyMonospace};
    overflow: auto;
    max-height: 360px;
    margin-top: ${theme.spacing(1.5)};
    white-space: pre-wrap;
    word-break: break-all;
  `,
  planNode: css`
    border-left: 2px solid ${theme.colors.border.medium};
    padding-left: ${theme.spacing(1.5)};
    padding-top: ${theme.spacing(0.75)};
    padding-bottom: ${theme.spacing(0.75)};
  `,
  planNodeCard: css`
    background: ${theme.colors.background.primary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.25, 1.5)};
    transition: border-color 0.15s ease;

    &:hover {
      border-color: ${theme.colors.border.strong};
    }
  `,
  planNodeCardHot: css`
    border-left: 3px solid ${theme.colors.error.main};
  `,
  planNodeCardWarn: css`
    border-left: 3px solid ${theme.colors.warning.main};
  `,
  planNodeHeader: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  planNodeName: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(0.75)};
    font-weight: ${theme.typography.fontWeightMedium};
    color: ${theme.colors.primary.text};
  `,
  riskyBadge: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    padding: 1px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border: 1px solid ${theme.colors.warning.border};
  `,
  parallelBadge: css`
    font-size: 11px;
    padding: 1px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.info.transparent};
    color: ${theme.colors.info.text};
    border: 1px solid ${theme.colors.info.border};
  `,
  costArea: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
  costBarTrack: css`
    width: 72px;
    height: 6px;
    border-radius: 3px;
    background: ${theme.colors.background.secondary};
    overflow: hidden;
  `,
  planNodeGrid: css`
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: ${theme.spacing(0.75, 1.5)};
    margin-top: ${theme.spacing(1)};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
  gridLabel: css`
    color: ${theme.colors.text.secondary};
  `,
  gridValue: css`
    color: ${theme.colors.text.primary};
    margin-left: ${theme.spacing(0.5)};
    font-family: ${theme.typography.fontFamilyMonospace};
  `,
  expandButton: css`
    background: none;
    border: none;
    padding: 2px;
    cursor: pointer;
    display: flex;
    align-items: center;
    color: ${theme.colors.text.secondary};
    border-radius: ${theme.shape.radius.default};

    &:hover {
      background: ${theme.colors.action.hover};
      color: ${theme.colors.text.primary};
    }
  `,
  statsTable: css`
    width: 100%;
    border-collapse: collapse;
    font-size: ${theme.typography.bodySmall.fontSize};

    th,
    td {
      text-align: left;
      padding: ${theme.spacing(0.75, 1)};
      border-bottom: 1px solid ${theme.colors.border.weak};
    }

    th {
      color: ${theme.colors.text.secondary};
      font-weight: ${theme.typography.fontWeightMedium};
      background: ${theme.colors.background.secondary};
    }
  `,
  warningCard: css`
    background: ${theme.colors.warning.transparent};
    border: 1px solid ${theme.colors.warning.border};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.5)};
    display: flex;
    gap: ${theme.spacing(1.25)};
  `,
  warningCardError: css`
    background: ${theme.colors.error.transparent};
    border: 1px solid ${theme.colors.error.border};
  `,
  warningIcon: css`
    color: ${theme.colors.warning.text};
    flex-shrink: 0;
    margin-top: 2px;
  `,
  warningIconError: css`
    color: ${theme.colors.error.text};
  `,
  warningContent: css`
    flex: 1;
    min-width: 0;
  `,
  warningHeaderRow: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  warningTitle: css`
    color: ${theme.colors.text.primary};
    font-weight: ${theme.typography.fontWeightMedium};
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
  `,
  severityPill: css`
    font-size: 11px;
    padding: 1px 6px;
    border-radius: ${theme.shape.radius.default};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  severityHigh: css`
    background: ${theme.colors.error.main};
    color: ${theme.colors.error.contrastText};
  `,
  severityMedium: css`
    background: ${theme.colors.warning.main};
    color: ${theme.colors.warning.contrastText};
  `,
  warningCode: css`
    margin-top: ${theme.spacing(1)};
    padding: ${theme.spacing(1)};
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.canvas};
    border: 1px solid ${theme.colors.border.weak};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    white-space: pre-wrap;
    word-break: break-word;
    color: ${theme.colors.text.primary};
  `,
  planTree: css`
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.5)};
    overflow-x: auto;
  `,
  childNodes: css`
    margin-top: ${theme.spacing(0.75)};
  `,
  warningsList: css`
    display: flex;
    flex-direction: column;
    gap: ${theme.spacing(1)};
  `,
});

function formatNumber(n?: number): string {
  if (n === undefined || n === null || !Number.isFinite(n)) {
    return '-';
  }
  if (Math.abs(n) >= 1000) {
    return n.toLocaleString(undefined, { maximumFractionDigits: 0 });
  }
  return n < 0.01 && n > 0 ? n.toExponential(2) : n.toFixed(2);
}

export const PlanNodeComponent: React.FC<{
  node: PlanNode;
  level: number;
  totalPlanCost: number;
  expandSignal?: ExpandSignal;
}> = ({ node, level, totalPlanCost, expandSignal }) => {
  const signalVersion = expandSignal?.version ?? 0;
  const [localToggle, setLocalToggle] = useState<{ version: number; expanded: boolean } | null>(null);
  const styles = useStyles2(getStyles);

  const isExpanded =
    localToggle && localToggle.version === signalVersion ? localToggle.expanded : (expandSignal?.expanded ?? true);

  const toggleExpanded = () => {
    setLocalToggle({ version: signalVersion, expanded: !isExpanded });
  };

  const hasChildren = Boolean(node.nodes && node.nodes.length > 0);
  const childrenSubtreeCost = (node.nodes ?? []).reduce((sum, c) => sum + (c.subtree_cost || 0), 0);
  const selfCost = Math.max(0, (node.subtree_cost || 0) - childrenSubtreeCost);
  const costPct = totalPlanCost > 0 ? Math.min(100, (selfCost / totalPlanCost) * 100) : 0;

  const physicalOp = node.header?.physical_op || node.name;
  const logicalOp = node.header?.logical_op || '';
  const isParallel = node.header?.parallel === '1' || node.header?.parallel === 'true';
  const isRisky = RISKY_OPS.has(physicalOp);
  const isHot = costPct >= 25;
  const isWarm = !isHot && (costPct >= 10 || isRisky);

  const barColor = isHot ? '#e02f44' : isWarm ? '#ff9830' : '#5794f2';

  return (
    <div style={{ marginLeft: level > 1 ? '14px' : '0px' }} className={styles.planNode}>
      <div
        className={`${styles.planNodeCard} ${isHot ? styles.planNodeCardHot : isWarm ? styles.planNodeCardWarn : ''}`}
      >
        <div className={styles.planNodeHeader}>
          <div className={styles.planNodeName}>
            {hasChildren && (
              <button
                type="button"
                onClick={toggleExpanded}
                className={styles.expandButton}
                aria-label={isExpanded ? 'Collapse operator' : 'Expand operator'}
              >
                {isExpanded ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
              </button>
            )}
            <span>{node.name || physicalOp}</span>
            {isRisky && (
              <span className={styles.riskyBadge}>
                <AlertCircle size={11} />
                {physicalOp}
              </span>
            )}
            {isParallel && <span className={styles.parallelBadge}>Parallel</span>}
          </div>

          <div className={styles.costArea}>
            <span>Self: {costPct.toFixed(1)}%</span>
            <div className={styles.costBarTrack} title={`Operator self cost: ${selfCost.toFixed(4)} (${costPct.toFixed(1)}%)`}>
              <div style={{ width: `${Math.max(costPct, 2)}%`, height: '100%', background: barColor }} />
            </div>
            <span>Subtree: {formatNumber(node.subtree_cost)}</span>
          </div>
        </div>

        <div className={styles.planNodeGrid}>
          {physicalOp && (
            <div>
              <span className={styles.gridLabel}>Physical Op:</span>
              <span className={styles.gridValue}>{physicalOp}</span>
            </div>
          )}
          {logicalOp && (
            <div>
              <span className={styles.gridLabel}>Logical Op:</span>
              <span className={styles.gridValue}>{logicalOp}</span>
            </div>
          )}
          <div>
            <span className={styles.gridLabel}>Est. Rows:</span>
            <span className={styles.gridValue}>{formatNumber(node.estimated_rows)}</span>
          </div>
          {node.header?.estimate_cpu !== undefined && node.header.estimate_cpu > 0 && (
            <div>
              <span className={styles.gridLabel}>Est. CPU:</span>
              <span className={styles.gridValue}>{formatNumber(node.header.estimate_cpu)}</span>
            </div>
          )}
          {node.header?.estimate_io !== undefined && node.header.estimate_io > 0 && (
            <div>
              <span className={styles.gridLabel}>Est. I/O:</span>
              <span className={styles.gridValue}>{formatNumber(node.header.estimate_io)}</span>
            </div>
          )}
        </div>
      </div>

      {hasChildren && isExpanded && (
        <div className={styles.childNodes}>
          {node.nodes.map((childNode, idx) => (
            <PlanNodeComponent
              key={idx}
              node={childNode}
              level={level + 1}
              totalPlanCost={totalPlanCost}
              expandSignal={expandSignal}
            />
          ))}
        </div>
      )}
    </div>
  );
};
