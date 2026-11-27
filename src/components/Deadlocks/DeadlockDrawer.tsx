import React, { useMemo, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { Alert, Button, Card, Tab, TabsBar, useStyles2 } from '@grafana/ui';
import { Activity, BarChart3, Check, Copy, Download, FileSearch } from 'lucide-react';
import { DeadlockItem, DeadlockProcess, LOCK_MODE_DESCRIPTIONS } from './types';
import { DeadlockGraph } from './DeadlockGraph';
import { SqlQueryCard } from '../SqlQueryCard';
import { copyToClipboard, downloadTextFile } from '../../utils/sqlHelpers';

type DrawerTabId = 'graph' | 'processes' | 'resources' | 'xml';

const getStyles = (theme: GrafanaTheme2) => ({
  headerActions: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(2)};
  `,
  section: css`
    margin-top: ${theme.spacing(2)};
  `,
  diagnosisBox: css`
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.medium};
    border-left: 4px solid ${theme.colors.error.main};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.75, 2)};
    margin-bottom: ${theme.spacing(2)};
  `,
  diagnosisTitle: css`
    font-size: ${theme.typography.h5.fontSize};
    font-weight: ${theme.typography.fontWeightBold};
    color: ${theme.colors.text.primary};
    margin-bottom: ${theme.spacing(0.75)};
  `,
  diagnosisSummary: css`
    margin: 0 0 ${theme.spacing(1)} 0;
    color: ${theme.colors.text.primary};
    font-size: ${theme.typography.body.fontSize};
  `,
  recList: css`
    margin: ${theme.spacing(0.5, 0, 0, 0)};
    padding-left: ${theme.spacing(2.5)};
    font-size: ${theme.typography.bodySmall.fontSize};
    li {
      margin-bottom: 4px;
    }
  `,
  stepsList: css`
    display: flex;
    flex-direction: column;
    gap: ${theme.spacing(1)};
    margin-top: ${theme.spacing(1.5)};
  `,
  stepRow: css`
    display: flex;
    align-items: flex-start;
    gap: ${theme.spacing(1.25)};
    padding: ${theme.spacing(1, 1.5)};
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
  stepNum: css`
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: ${theme.colors.primary.transparent};
    color: ${theme.colors.primary.text};
    border: 1px solid ${theme.colors.primary.border};
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightBold};
    flex-shrink: 0;
  `,
  processGrid: css`
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
    gap: ${theme.spacing(2)};
  `,
  processCard: css`
    background: ${theme.colors.background.primary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    overflow: hidden;
  `,
  processCardVictim: css`
    border-color: ${theme.colors.error.border};
  `,
  processHeader: css`
    padding: ${theme.spacing(1.25, 1.5)};
    background: ${theme.colors.background.secondary};
    border-bottom: 1px solid ${theme.colors.border.weak};
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  processHeaderVictim: css`
    background: ${theme.colors.error.transparent};
  `,
  processBody: css`
    padding: ${theme.spacing(1.5)};
  `,
  metaGrid: css`
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: ${theme.spacing(1, 1.5)};
    margin-bottom: ${theme.spacing(1.5)};
  `,
  metaLabel: css`
    font-size: 11px;
    color: ${theme.colors.text.secondary};
  `,
  metaValue: css`
    font-size: ${theme.typography.bodySmall.fontSize};
    font-family: ${theme.typography.fontFamilyMonospace};
    color: ${theme.colors.text.primary};
    word-break: break-word;
  `,
  stackFrame: css`
    padding: ${theme.spacing(0.75, 1)};
    border-left: 2px solid ${theme.colors.primary.main};
    background: ${theme.colors.background.secondary};
    margin-bottom: ${theme.spacing(0.75)};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 11px;
  `,
  preBlock: css`
    margin: ${theme.spacing(0.75, 0)};
    padding: ${theme.spacing(1.25)};
    background: ${theme.colors.background.canvas};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 11.5px;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 60vh;
    overflow: auto;
  `,
  badge: css`
    display: inline-flex;
    align-items: center;
    padding: 2px 7px;
    border-radius: ${theme.shape.radius.default};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightMedium};
    white-space: nowrap;
  `,
  badgeError: css`
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border: 1px solid ${theme.colors.error.border};
  `,
  badgeSuccess: css`
    background: ${theme.colors.success.transparent};
    color: ${theme.colors.success.text};
    border: 1px solid ${theme.colors.success.border};
  `,
  badgeWarning: css`
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border: 1px solid ${theme.colors.warning.border};
  `,
  badgeInfo: css`
    background: ${theme.colors.info.transparent};
    color: ${theme.colors.info.text};
    border: 1px solid ${theme.colors.info.border};
  `,
  resTable: css`
    width: 100%;
    border-collapse: collapse;
    font-size: ${theme.typography.bodySmall.fontSize};
    th {
      text-align: left;
      padding: ${theme.spacing(1)};
      background: ${theme.colors.background.secondary};
      color: ${theme.colors.text.secondary};
      border-bottom: 1px solid ${theme.colors.border.medium};
    }
    td {
      padding: ${theme.spacing(1)};
      border-bottom: 1px solid ${theme.colors.border.weak};
      vertical-align: top;
    }
  `,
  mono: css`
    font-family: ${theme.typography.fontFamilyMonospace};
  `,
});

export const DeadlockDrawer: React.FC<{
  deadlock: DeadlockItem;
  onOpenSnapshotsAround: (timestampIso: string) => void;
  onOpenPlanAnalysis: (database?: string) => void;
  onOpenQueryMetrics: (database?: string) => void;
}> = ({ deadlock, onOpenSnapshotsAround, onOpenPlanAnalysis, onOpenQueryMetrics }) => {
  const styles = useStyles2(getStyles);
  const [activeTab, setActiveTab] = useState<DrawerTabId>('graph');
  const [copiedXml, setCopiedXml] = useState(false);

  const procById = useMemo(() => {
    const map: Record<string, DeadlockProcess> = {};
    for (const p of deadlock.processes ?? []) {
      map[p.id] = p;
    }
    return map;
  }, [deadlock.processes]);

  const handleDownloadXdl = () => {
    downloadTextFile(
      `deadlock-${deadlock.server || 'sql'}-${deadlock.id.slice(0, 8)}.xdl`,
      deadlock.xml_report,
      'application/xml;charset=utf-8'
    );
  };

  const handleCopyXml = async () => {
    if (await copyToClipboard(deadlock.xml_report)) {
      setCopiedXml(true);
      setTimeout(() => setCopiedXml(false), 1800);
    }
  };

  const primaryDb = deadlock.databases?.[0];

  return (
    <div>
      <div className={styles.headerActions}>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
          <span className={`${styles.badge} ${styles.badgeError}`}>
            {deadlock.diagnosis?.pattern_label || 'Deadlock'}
          </span>
          <span className={styles.mono} style={{ fontSize: 12, opacity: 0.8 }}>
            {new Date(deadlock.timestamp).toLocaleString()}
          </span>
          {deadlock.pattern_occurrences > 1 && (
            <span className={`${styles.badge} ${styles.badgeWarning}`}>
              {deadlock.pattern_occurrences}× in selected window
            </span>
          )}
        </div>

        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenSnapshotsAround(deadlock.timestamp)}>
            <Activity size={13} />
            <span style={{ marginLeft: 5 }}>Snapshots (±2m)</span>
          </Button>
          <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenPlanAnalysis(primaryDb)}>
            <FileSearch size={13} />
            <span style={{ marginLeft: 5 }}>Plan Analysis</span>
          </Button>
          <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenQueryMetrics(primaryDb)}>
            <BarChart3 size={13} />
            <span style={{ marginLeft: 5 }}>Query Metrics</span>
          </Button>
          <Button size="sm" variant="secondary" fill="outline" onClick={handleDownloadXdl}>
            <Download size={13} />
            <span style={{ marginLeft: 5 }}>Download .xdl</span>
          </Button>
        </div>
      </div>

      <TabsBar>
        <Tab label="Graph & Diagnosis" active={activeTab === 'graph'} onChangeTab={() => setActiveTab('graph')} />
        <Tab
          label="Processes & SQL"
          counter={(deadlock.processes ?? []).length}
          active={activeTab === 'processes'}
          onChangeTab={() => setActiveTab('processes')}
        />
        <Tab
          label="Contested Resources"
          counter={(deadlock.resources ?? []).length}
          active={activeTab === 'resources'}
          onChangeTab={() => setActiveTab('resources')}
        />
        <Tab label="Raw XML (.xdl)" active={activeTab === 'xml'} onChangeTab={() => setActiveTab('xml')} />
      </TabsBar>

      <div className={styles.section}>
        {activeTab === 'graph' && (
          <div>
            <div className={styles.diagnosisBox}>
              <div className={styles.diagnosisTitle}>
                Automated Diagnosis: {deadlock.diagnosis?.pattern_label || 'Deadlock'}
              </div>
              <p className={styles.diagnosisSummary}>{deadlock.diagnosis?.summary}</p>
              {(deadlock.diagnosis?.recommendations ?? []).length > 0 && (
                <>
                  <strong style={{ fontSize: 12 }}>Recommended Fixes:</strong>
                  <ul className={styles.recList}>
                    {deadlock.diagnosis.recommendations.map((rec, i) => (
                      <li key={i}>{rec}</li>
                    ))}
                  </ul>
                </>
              )}
            </div>

            <DeadlockGraph deadlock={deadlock} />

            {(deadlock.diagnosis?.steps ?? []).length > 0 && (
              <div style={{ marginTop: 18 }}>
                <h4 style={{ margin: '0 0 8px 0' }}>Circular Wait Breakdown</h4>
                <div className={styles.stepsList}>
                  {deadlock.diagnosis.steps.map((step, i) => (
                    <div key={i} className={styles.stepRow}>
                      <span className={styles.stepNum}>{i + 1}</span>
                      <span>{step.text}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'processes' && (
          <div className={styles.processGrid}>
            {(deadlock.processes ?? []).map((p) => (
              <div key={p.id} className={`${styles.processCard} ${p.is_victim ? styles.processCardVictim : ''}`}>
                <div className={`${styles.processHeader} ${p.is_victim ? styles.processHeaderVictim : ''}`}>
                  <div>
                    <strong className={styles.mono} style={{ fontSize: 14 }}>
                      SPID {p.spid}
                    </strong>
                    <span style={{ marginLeft: 8, opacity: 0.75 }}>{p.client_app}</span>
                  </div>
                  <span className={`${styles.badge} ${p.is_victim ? styles.badgeError : styles.badgeInfo}`}>
                    {p.is_victim ? 'DEADLOCK VICTIM (1205)' : 'SURVIVOR'}
                  </span>
                </div>

                <div className={styles.processBody}>
                  <div className={styles.metaGrid}>
                    <div>
                      <div className={styles.metaLabel}>Database</div>
                      <div className={styles.metaValue}>{p.current_db_name || p.current_db || '-'}</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Wait Time</div>
                      <div className={styles.metaValue}>{p.wait_time_ms} ms</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Log Used</div>
                      <div className={styles.metaValue}>{p.log_used.toLocaleString()} B</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Isolation Level</div>
                      <div className={styles.metaValue}>{p.isolation_level}</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Requested Mode</div>
                      <div className={styles.metaValue}>{p.lock_mode || '-'}</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Deadlock Priority</div>
                      <div className={styles.metaValue}>{p.priority}</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Login</div>
                      <div className={styles.metaValue}>{p.login_name}</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Host</div>
                      <div className={styles.metaValue}>{p.host_name}</div>
                    </div>
                    <div>
                      <div className={styles.metaLabel}>Open Tran</div>
                      <div className={styles.metaValue}>
                        {p.tran_count} ({p.transaction_name || '-'})
                      </div>
                    </div>
                  </div>

                  <SqlQueryCard sql={p.active_sql || p.input_buf || '(no SQL text)'} heading="Active Statement" />

                  {(p.frames ?? []).length > 0 && (
                    <div style={{ marginTop: 12 }}>
                      <div className={styles.metaLabel} style={{ marginBottom: 4 }}>
                        Call Stack ({p.frames.length})
                      </div>
                      {p.frames.map((f, idx) => (
                        <div key={idx} className={styles.stackFrame}>
                          <strong>{f.proc_name}</strong>
                          {f.line ? ` : line ${f.line}` : ''}
                          {f.sql_handle ? <span style={{ opacity: 0.7 }}> ({f.sql_handle})</span> : null}
                        </div>
                      ))}
                    </div>
                  )}

                  {p.input_buf && (
                    <details style={{ marginTop: 10 }}>
                      <summary style={{ cursor: 'pointer', fontSize: 12, opacity: 0.8 }}>
                        Show full &lt;inputbuf&gt; batch
                      </summary>
                      <pre className={styles.preBlock}>{p.input_buf}</pre>
                    </details>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {activeTab === 'resources' && (
          <div>
            <div style={{ overflowX: 'auto' }}>
              <table className={styles.resTable}>
                <thead>
                  <tr>
                    <th>Lock Type</th>
                    <th>Object &amp; Index</th>
                    <th>Physical Target</th>
                    <th>Owners (Holding)</th>
                    <th>Waiters (Blocked)</th>
                  </tr>
                </thead>
                <tbody>
                  {(deadlock.resources ?? []).map((r) => {
                    const phys = r.page_id
                      ? `File ${r.file_id} : Page ${r.page_id}${r.slot ? ` : Slot ${r.slot}` : ''}`
                      : `HoBT ${r.hobt_id || '-'}`;
                    return (
                      <tr key={r.id}>
                        <td>
                          <span className={`${styles.badge} ${styles.badgeWarning}`}>{r.type}</span>
                        </td>
                        <td>
                          <div className={styles.mono}>{r.object_name}</div>
                          {r.index_name && (
                            <div className={styles.mono} style={{ fontSize: 11, opacity: 0.75, marginTop: 2 }}>
                              Index: {r.index_name}
                            </div>
                          )}
                        </td>
                        <td className={styles.mono}>{phys}</td>
                        <td>
                          {(r.owners ?? []).length === 0
                            ? '-'
                            : r.owners.map((o, i) => {
                                const proc = procById[o.process_id];
                                return (
                                  <div key={i} style={{ marginBottom: 4 }}>
                                    <span className={`${styles.badge} ${styles.badgeSuccess}`}>{o.mode}</span>{' '}
                                    <span className={styles.mono}>SPID {proc?.spid || o.process_id}</span>{' '}
                                    <span style={{ fontSize: 11, opacity: 0.75 }}>
                                      {LOCK_MODE_DESCRIPTIONS[o.mode] || ''}
                                    </span>
                                  </div>
                                );
                              })}
                        </td>
                        <td>
                          {(r.waiters ?? []).length === 0
                            ? '-'
                            : r.waiters.map((w, i) => {
                                const proc = procById[w.process_id];
                                return (
                                  <div key={i} style={{ marginBottom: 4 }}>
                                    <span className={`${styles.badge} ${styles.badgeError}`}>
                                      {w.mode}
                                      {w.request_type === 'convert' ? ' (convert)' : ''}
                                    </span>{' '}
                                    <span className={styles.mono}>SPID {proc?.spid || w.process_id}</span>{' '}
                                    <span style={{ fontSize: 11, opacity: 0.75 }}>
                                      {LOCK_MODE_DESCRIPTIONS[w.mode] || ''}
                                    </span>
                                  </div>
                                );
                              })}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            <div style={{ marginTop: 14 }}>
              <Alert title="Lock Compatibility" severity="info">
                Exclusive (X) locks conflict with every other data lock mode (S, U, X, IS, IX). Two Shared (S) locks are
                compatible with each other, which is why two sessions can hold S locks simultaneously and then deadlock
                when both attempt a convert to X.
              </Alert>
            </div>
          </div>
        )}

        {activeTab === 'xml' && (
          <Card>
            <Card.Heading>Original xml_deadlock_report (.xdl)</Card.Heading>
            <Card.Actions>
              <Button size="sm" variant="secondary" fill="outline" onClick={handleCopyXml}>
                {copiedXml ? <Check size={13} /> : <Copy size={13} />}
                <span style={{ marginLeft: 5 }}>{copiedXml ? 'Copied' : 'Copy XML'}</span>
              </Button>
              <Button size="sm" variant="secondary" fill="outline" onClick={handleDownloadXdl}>
                <Download size={13} />
                <span style={{ marginLeft: 5 }}>Download .xdl</span>
              </Button>
            </Card.Actions>
            <Card.Description>
              <pre className={styles.preBlock}>{deadlock.xml_report}</pre>
            </Card.Description>
          </Card>
        )}
      </div>
    </div>
  );
};
