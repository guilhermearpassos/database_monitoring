import React from 'react';
import {Tooltip, useTheme2} from '@grafana/ui';
import {QuerySample} from "./types";
import {formatSQL} from "../../utils/formatters";
import {getStyles} from "../../utils/styles";

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
      <div className={styles.formattedSQL}  dangerouslySetInnerHTML={{ __html: formattedQuery }}>
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
                  <Tooltip interactive={true} content={tooltipContent} placement="right">
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
