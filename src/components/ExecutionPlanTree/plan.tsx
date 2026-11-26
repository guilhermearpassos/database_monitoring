import React, { useMemo, useState } from 'react';
import { AlertTriangle, Check, ChevronDown, ChevronRight, Copy, Download, Info } from 'lucide-react';
import { Button, useStyles2, useTheme2 } from '@grafana/ui';
import { ExpandSignal, getStyles, PlanNode, PlanNodeComponent } from './planNode';
import { buildMissingIndexDDL, copyToClipboard, downloadTextFile } from '../../utils/sqlHelpers';

export interface ParsedExecutionPlan {
  plan: ExecutionPlan;
  stats_usage: StatisticsInfo[];
  warnings: PlanWarning[];
  nodes: PlanNode[];
}

export interface PlanAffectingConvert {
  convert_issue: string;
  expression: string;
}

export interface PlanWarning {
  convert?: PlanAffectingConvert;
  missing_index?: MissingIndexWarning;
}

export interface MissingIndexWarning {
  database: string;
  schema: string;
  table: string;
  impact: number;
  equality_columns?: string[];
  inequality_columns?: string[];
  include_columns?: string[];
}

export interface StatisticsInfo {
  last_update: string;
  modification_count: number;
  sampling_percent: number;
  statistics: string;
  table: string;
}

export interface ExecutionPlan {
  plan_handle: string;
  xml_plan: string;
}

const WarningComponent: React.FC<{ warning: PlanWarning }> = ({ warning }) => {
  const styles = useStyles2(getStyles);
  const [copied, setCopied] = useState(false);

  if (!warning.convert && !warning.missing_index) {
    return null;
  }

  const isSeekPlanConvert = Boolean(
    warning.convert?.convert_issue && warning.convert.convert_issue.toLowerCase().includes('seek')
  );
  const ddl = warning.missing_index ? buildMissingIndexDDL(warning.missing_index) : '';
  const copyText = warning.missing_index ? ddl : warning.convert?.expression ?? '';

  const handleCopy = async () => {
    if (await copyToClipboard(copyText)) {
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    }
  };

  return (
    <div className={`${styles.warningCard} ${isSeekPlanConvert ? styles.warningCardError : ''}`}>
      <div className={`${styles.warningIcon} ${isSeekPlanConvert ? styles.warningIconError : ''}`}>
        <AlertTriangle size={18} />
      </div>

      {warning.convert && (
        <div className={styles.warningContent}>
          <div className={styles.warningHeaderRow}>
            <div className={styles.warningTitle}>
              <span className={`${styles.severityPill} ${isSeekPlanConvert ? styles.severityHigh : styles.severityMedium}`}>
                {warning.convert.convert_issue || 'Implicit Conversion'}
              </span>
              <span>Plan-Affecting Implicit Conversion</span>
            </div>
            <Button size="sm" variant="secondary" fill="outline" onClick={handleCopy}>
              {copied ? <Check size={13} /> : <Copy size={13} />}
              <span style={{ marginLeft: 4 }}>{copied ? 'Copied' : 'Copy Expression'}</span>
            </Button>
          </div>
          <div className={styles.warningCode}>{warning.convert.expression}</div>
        </div>
      )}

      {warning.missing_index && (
        <div className={styles.warningContent}>
          <div className={styles.warningHeaderRow}>
            <div className={styles.warningTitle}>
              <span
                className={`${styles.severityPill} ${
                  warning.missing_index.impact >= 50 ? styles.severityHigh : styles.severityMedium
                }`}
              >
                Impact: {warning.missing_index.impact.toFixed(1)}%
              </span>
              <span>
                Missing Index on {warning.missing_index.schema}.{warning.missing_index.table}
              </span>
            </div>
            <Button size="sm" variant="secondary" fill="outline" onClick={handleCopy}>
              {copied ? <Check size={13} /> : <Copy size={13} />}
              <span style={{ marginLeft: 4 }}>{copied ? 'Copied DDL' : 'Copy CREATE INDEX'}</span>
            </Button>
          </div>
          <div className={styles.warningCode}>{ddl}</div>
        </div>
      )}
    </div>
  );
};

export const ExecutionPlanViewer: React.FC<{ executionPlan: ParsedExecutionPlan }> = ({ executionPlan }) => {
  const [showRawXml, setShowRawXml] = useState(false);
  const [copiedXml, setCopiedXml] = useState(false);
  const [expandSignal, setExpandSignal] = useState<ExpandSignal | undefined>(undefined);
  const styles = useStyles2(getStyles);
  const theme = useTheme2();

  const totalPlanCost = useMemo(() => {
    return (executionPlan.nodes ?? []).reduce((sum, n) => sum + (n.subtree_cost || 0), 0);
  }, [executionPlan.nodes]);

  const hasPlanXml = Boolean(executionPlan.plan?.xml_plan);

  const handleCopyXml = async () => {
    if (!hasPlanXml) {
      return;
    }
    if (await copyToClipboard(executionPlan.plan.xml_plan)) {
      setCopiedXml(true);
      setTimeout(() => setCopiedXml(false), 1800);
    }
  };

  const handleDownloadSqlPlan = () => {
    if (!hasPlanXml) {
      return;
    }
    const safeHandle = (executionPlan.plan.plan_handle || 'plan').replace(/[^a-zA-Z0-9_-]/g, '').slice(0, 16);
    downloadTextFile(`sqlsights-${safeHandle}.sqlplan`, executionPlan.plan.xml_plan, 'application/xml;charset=utf-8');
  };

  if (!hasPlanXml && (!executionPlan.nodes || executionPlan.nodes.length === 0)) {
    return (
      <div className={styles.container}>
        <p style={{ color: theme.colors.text.secondary, margin: 0 }}>
          No execution plan has been captured for this sample yet.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <Info size={20} color={theme.colors.primary.text} />
          <h3 className={styles.title}>Execution Plan</h3>
        </div>
        <div className={styles.toolbar}>
          <Button
            size="sm"
            variant="secondary"
            fill="outline"
            onClick={() => setExpandSignal((prev) => ({ expanded: true, version: (prev?.version ?? 0) + 1 }))}
          >
            Expand All
          </Button>
          <Button
            size="sm"
            variant="secondary"
            fill="outline"
            onClick={() => setExpandSignal((prev) => ({ expanded: false, version: (prev?.version ?? 0) + 1 }))}
          >
            Collapse All
          </Button>
          {hasPlanXml && (
            <>
              <Button size="sm" variant="secondary" fill="outline" onClick={handleCopyXml}>
                {copiedXml ? <Check size={13} /> : <Copy size={13} />}
                <span style={{ marginLeft: 4 }}>{copiedXml ? 'Copied XML' : 'Copy XML'}</span>
              </Button>
              <Button size="sm" variant="secondary" fill="outline" onClick={handleDownloadSqlPlan}>
                <Download size={13} />
                <span style={{ marginLeft: 4 }}>Download .sqlplan</span>
              </Button>
              <Button size="sm" variant="secondary" fill="outline" onClick={() => setShowRawXml((v) => !v)}>
                {showRawXml ? <ChevronDown size={13} /> : <ChevronRight size={13} />}
                <span style={{ marginLeft: 4 }}>{showRawXml ? 'Hide XML' : 'View XML'}</span>
              </Button>
            </>
          )}
        </div>
      </div>

      {showRawXml && hasPlanXml && (
        <div className={styles.section}>
          <pre className={styles.xmlPre}>{executionPlan.plan.xml_plan}</pre>
        </div>
      )}

      {executionPlan.warnings && executionPlan.warnings.length > 0 && (
        <div className={styles.section}>
          <h4 className={styles.sectionTitle}>
            <AlertTriangle size={18} color={theme.colors.warning.text} />
            Warnings ({executionPlan.warnings.length})
          </h4>
          <div className={styles.warningsList}>
            {executionPlan.warnings.map((warning, idx) => (
              <WarningComponent key={idx} warning={warning} />
            ))}
          </div>
        </div>
      )}

      <div className={styles.section}>
        <h4 className={styles.sectionTitle}>Operator Tree</h4>
        <div className={styles.planTree}>
          {executionPlan.nodes &&
            executionPlan.nodes.map((node, idx) => (
              <PlanNodeComponent
                key={idx}
                node={node}
                level={1}
                totalPlanCost={totalPlanCost}
                expandSignal={expandSignal}
              />
            ))}
        </div>
      </div>

      {executionPlan.stats_usage && executionPlan.stats_usage.length > 0 && (
        <div className={styles.section}>
          <h4 className={styles.sectionTitle}>Optimizer Statistics Used ({executionPlan.stats_usage.length})</h4>
          <table className={styles.statsTable}>
            <thead>
              <tr>
                <th>Table</th>
                <th>Statistics</th>
                <th>Sampling</th>
                <th>Modifications</th>
                <th>Last Updated</th>
              </tr>
            </thead>
            <tbody>
              {executionPlan.stats_usage.map((stats, idx) => (
                <tr key={idx}>
                  <td>{stats.table}</td>
                  <td>{stats.statistics}</td>
                  <td>{stats.sampling_percent.toFixed(1)}%</td>
                  <td>{stats.modification_count.toLocaleString()}</td>
                  <td>{stats.last_update}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
