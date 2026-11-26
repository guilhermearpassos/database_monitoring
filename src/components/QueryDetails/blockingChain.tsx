import React from 'react';
import { TimeRange } from '@grafana/data';
import { Tooltip, useTheme2 } from '@grafana/ui';
import { QuerySample } from './types';
import { formatSQL } from '../../utils/formatters';
import { getStyles } from '../../utils/styles';
import { TempoTraceLink } from '../TempoTraceLink';

export interface BlockingNode {
  level: number;
  query_sample: QuerySample;
  child_nodes?: BlockingNode[];
}

export interface BlockingChain {
  roots: BlockingNode[];
}

interface BlockingChainProps {
  chain: BlockingChain;
  currentSampleId?: string;
  timeRange?: TimeRange;
  onSelectSample?: (snapId: string, sampleId: string) => void;
}

const BlockingNodeComponent: React.FC<{
  node: BlockingNode;
  currentSampleId?: string;
  timeRange?: TimeRange;
  onSelectSample?: (snapId: string, sampleId: string) => void;
}> = ({ node, currentSampleId, timeRange, onSelectSample }) => {
  const theme = useTheme2();
  const styles = getStyles(theme);

  const isCurrentSample = Boolean(currentSampleId && node.query_sample.sample_id === currentSampleId);
  const levelIndent = `${node.level * 24}px`;
  const formattedQuery = formatSQL(node.query_sample.query);
  const clickable = Boolean(onSelectSample && !isCurrentSample && node.query_sample.sample_id && node.query_sample.snap_id);

  const tooltipContent = (
    <div className={styles.tooltipContent}>
      <div className={styles.formattedSQL} dangerouslySetInnerHTML={{ __html: formattedQuery }} />
    </div>
  );

  return (
    <div className={styles.nodeWrapper} style={{ '--level-indent': levelIndent } as React.CSSProperties}>
      <div
        className={`${styles.nodeContent} ${isCurrentSample ? styles.currentNode : ''}`}
        style={clickable ? { cursor: 'pointer' } : undefined}
        onClick={() => {
          if (clickable && onSelectSample) {
            onSelectSample(node.query_sample.snap_id, node.query_sample.sample_id);
          }
        }}
        title={clickable ? 'Click to inspect this session' : undefined}
      >
        <div
          className={`${styles.indicator} ${
            node.query_sample.is_blocker ? styles.blockerIndicator : styles.blockedIndicator
          }`}
          title={node.query_sample.is_blocker ? 'Blocker' : 'Blocked'}
        />

        <div className={styles.nodeInfo}>
          <div className={styles.sessionId} style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <span>Session ID: {node.query_sample.sid}</span>
            {isCurrentSample && <span className={styles.badge}>Current</span>}
            {node.query_sample.is_blocker && (
              <span
                className={styles.badge}
                style={{ background: theme.colors.error.transparent, color: theme.colors.error.text }}
              >
                Blocker
              </span>
            )}
            {node.query_sample.database && (
              <span style={{ fontSize: 12, color: theme.colors.text.secondary }}>{node.query_sample.database}</span>
            )}
            {node.query_sample.trace_id && (
              <TempoTraceLink traceId={node.query_sample.trace_id} timeRange={timeRange} compact />
            )}
          </div>

          <Tooltip interactive={true} content={tooltipContent} placement="right">
            <div className={styles.queryText}>{node.query_sample.query}</div>
          </Tooltip>

          <div className={styles.metadata}>
            <div className={styles.metadataItem}>
              <span className={styles.label}>Status:</span>
              <span>{node.query_sample.status}</span>
            </div>
            <div className={styles.metadataItem}>
              <span className={styles.label}>Elapsed:</span>
              <span>{node.query_sample.execution_time}</span>
            </div>
            {node.query_sample.wait_event && (
              <div className={styles.metadataItem}>
                <span className={styles.label}>Wait:</span>
                <span>{node.query_sample.wait_event}</span>
              </div>
            )}
            {node.query_sample.user && (
              <div className={styles.metadataItem}>
                <span className={styles.label}>User:</span>
                <span>{node.query_sample.user}</span>
              </div>
            )}
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
              timeRange={timeRange}
              onSelectSample={onSelectSample}
            />
          ))}
        </>
      )}
    </div>
  );
};

export const BlockingChainComponent: React.FC<BlockingChainProps> = ({
  chain,
  currentSampleId,
  timeRange,
  onSelectSample,
}) => {
  const theme = useTheme2();
  const styles = getStyles(theme);

  if (!chain || !chain.roots || chain.roots.length === 0) {
    return (
      <div className={styles.container}>
        <div style={{ color: theme.colors.text.secondary, padding: theme.spacing(2) }}>
          No blocking chain detected for this sample.
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container} style={{ maxHeight: '60vh' }}>
      <div>
        {chain.roots.map((root, index) => (
          <BlockingNodeComponent
            key={`root-${root.query_sample.sid}-${index}`}
            node={root}
            currentSampleId={currentSampleId}
            timeRange={timeRange}
            onSelectSample={onSelectSample}
          />
        ))}
      </div>
    </div>
  );
};
