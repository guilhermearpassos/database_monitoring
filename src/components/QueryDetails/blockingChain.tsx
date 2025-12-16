import React from 'react';
import { css } from '@emotion/css';
import { useTheme2, Tooltip } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';
import { format } from 'sql-formatter';

// Format SQL with sql-formatter library
const formatSQL = (sql: string): string => {
  if (!sql) {
      return '';
  }

  try {
    return format(sql, {
      language: 'tsql', // Use 'tsql' for SQL Server, 'sql' for generic, 'postgresql', 'mysql', etc.
      tabWidth: 2,
      keywordCase: 'upper',
      linesBetweenQueries: 2,
    });
  } catch (error) {
    // If formatting fails, return original
    console.error('SQL formatting error:', error);
    return sql;
  }
};

interface QuerySample {
    sid: string;
    query: string;
    status: string;
    execution_time: string;
    is_blocker: boolean;
    sample_id?: string;
}

interface BlockingNode {
    level: number;
    query_sample: QuerySample;
    child_nodes?: BlockingNode[];
}

interface BlockingChainProps {
    chain: BlockingChain;
    currentSampleId?: string;
}
export interface BlockingChain {
    roots: BlockingNode[];
}

const getStyles = (theme: GrafanaTheme2) => ({
    container: css`
    padding: ${theme.spacing(2)};
    background: ${theme.colors.background.primary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    max-height: 30vh;
    overflow-y: auto;
  `,
    title: css`
    font-weight: ${theme.typography.fontWeightMedium};
    font-size: ${theme.typography.h3.fontSize};
    margin-bottom: ${theme.spacing(2)};
    color: ${theme.colors.text.primary};
  `,
    nodeWrapper: css`
    position: relative;
    margin-left: var(--level-indent);
    padding-left: ${theme.spacing(2)};
    padding-top: ${theme.spacing(1)};
    padding-bottom: ${theme.spacing(1)};
    border-left: 2px solid ${theme.colors.border.medium};
  `,
    nodeContent: css`
    display: flex;
    align-items: flex-start;
    gap: ${theme.spacing(1)};
    padding: ${theme.spacing(1.5)};
    background: ${theme.colors.background.secondary};
    border-radius: ${theme.shape.radius.default};
    border: 1px solid ${theme.colors.border.weak};
    transition: all 0.2s ease;

    &:hover {
      background: ${theme.colors.emphasize(theme.colors.background.secondary, 0.03)};
      border-color: ${theme.colors.border.medium};
    }
  `,
    currentNode: css`
    background: ${theme.colors.emphasize(theme.colors.secondary.main, 0.1)};
    border: 3px solid ${theme.colors.primary.border};
    
    &:hover {
      background: ${theme.colors.emphasize(theme.colors.primary.main, 0.15)};
    }
  `,
    indicator: css`
    width: 12px;
    height: 12px;
    border-radius: 50%;
    flex-shrink: 0;
    margin-top: 4px;
  `,
    blockerIndicator: css`
    background: ${theme.colors.error.main};
    box-shadow: 0 0 0 2px ${theme.colors.error.transparent};
  `,
    blockedIndicator: css`
    background: ${theme.colors.warning.main};
    box-shadow: 0 0 0 2px ${theme.colors.warning.transparent};
  `,
    nodeInfo: css`
    flex: 1;
    min-width: 0;
  `,
    sessionId: css`
    font-weight: ${theme.typography.fontWeightMedium};
    color: ${theme.colors.text.primary};
    margin-bottom: ${theme.spacing(0.5)};
  `,
    queryText: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 500px;
    margin-bottom: ${theme.spacing(0.5)};
    cursor: help;
    
    &:hover {
      color: ${theme.colors.text.primary};
    }
  `,
  tooltipContent: css`
    max-width: 600px;
    max-height: 400px;
    overflow: auto;
  `,
  formattedSQL: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    white-space: pre-wrap;
    word-break: break-word;
    background: ${theme.colors.background.canvas};
    padding: ${theme.spacing(1.5)};
    border-radius: ${theme.shape.radius.default};
    border: 1px solid ${theme.colors.border.weak};
    line-height: 1.5;
  `,
    metadata: css`
    display: flex;
    gap: ${theme.spacing(2)};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
    metadataItem: css`
    display: flex;
    gap: ${theme.spacing(0.5)};
  `,
    label: css`
    color: ${theme.colors.text.disabled};
  `,
    badge: css`
    display: inline-block;
    padding: ${theme.spacing(0.25, 0.75)};
    background: ${theme.colors.primary.transparent};
    color: ${theme.colors.primary.text};
    border-radius: ${theme.shape.radius.default};
    font-size: ${theme.typography.bodySmall.fontSize};
    font-weight: ${theme.typography.fontWeightMedium};
    margin-left: ${theme.spacing(1)};
  `,
});

const BlockingNodeComponent: React.FC<{
    node: BlockingNode;
    currentSampleId?: string;
}> = ({ node, currentSampleId }) => {
    const theme = useTheme2();
    const styles = getStyles(theme);

    const isCurrentSample = currentSampleId && node.query_sample.sample_id === currentSampleId;
    const levelIndent = `${node.level * 24}px`;
    const formattedQuery = formatSQL(node.query_sample.query);

  const tooltipContent = (
    <div className={styles.tooltipContent}>
      <div className={styles.formattedSQL}>
        {formattedQuery}
      </div>
    </div>
  );
    return (
        <div
            className={styles.nodeWrapper}
            style={{ '--level-indent': levelIndent } as React.CSSProperties}
        >
            <div className={`${styles.nodeContent} ${isCurrentSample ? styles.currentNode : ''}`}>
                <div
                    className={`${styles.indicator} ${
                        node.query_sample.is_blocker ? styles.blockerIndicator : styles.blockedIndicator
                    }`}
                    title={node.query_sample.is_blocker ? 'Blocker' : 'Blocked'}
                />

                <div className={styles.nodeInfo}>
                    <div className={styles.sessionId}>
                        Session ID: {node.query_sample.sid}
                        {isCurrentSample && <span className={styles.badge}>Current</span>}
                    </div>
                  <Tooltip content={tooltipContent} placement="right">
                    <div className={styles.queryText}>
                      {node.query_sample.query}
                    </div>
                  </Tooltip>
                    <div className={styles.metadata}>
                        <div className={styles.metadataItem}>
                            <span className={styles.label}>Status:</span>
                            <span>{node.query_sample.status}</span>
                        </div>
                        <div className={styles.metadataItem}>
                            <span className={styles.label}>Time:</span>
                            <span>{node.query_sample.execution_time}</span>
                        </div>
                    </div>
                </div>
            </div>

            {node.child_nodes && node.child_nodes.length > 0 && (
                <>
                    {node.child_nodes.map((child, index) => (
                        <BlockingNodeComponent
                            key={`${child.query_sample.sid}-${index}`}
                            node={child}
                            currentSampleId={currentSampleId}
                        />
                    ))}
                </>
            )}
        </div>
    );
};

export const BlockingChainComponent: React.FC<BlockingChainProps> = ({ chain, currentSampleId }) => {
    const theme = useTheme2();
    const styles = getStyles(theme);

    if (!chain) {
        return (
            <div className={styles.container}>
                <h3 className={styles.title}>Blocking Chain</h3>
                <div style={{ color: theme.colors.text.secondary, padding: theme.spacing(2) }}>
                    No blocking detected
                </div>
            </div>
        );
    }

    return (
        <div className={styles.container}>
            <h3 className={styles.title}>Blocking Chain</h3>
            <div>
                {chain.roots.map((root, index) => (
                    <BlockingNodeComponent
                        key={`root-${root.query_sample.sid}-${index}`}
                        node={root}
                        currentSampleId={currentSampleId}
                    />
                ))}
            </div>
        </div>
    );
};
