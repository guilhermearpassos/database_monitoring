import React, {useState} from 'react';
import {ChevronDown, ChevronRight} from 'lucide-react';
import {GrafanaTheme2} from "@grafana/data";
import {css} from '@emotion/css';
import {useStyles2} from "@grafana/ui";

// Type definitions based on your Go structs
export interface PlanNodeHeader {
    PhysicalOp: string;
    LogicalOp: string;
    EstimateCpu: number;
    EstimateIO: number;
    EstimateRows: number;
    EstimatedCost: number;
    Parallel: string;
}

export interface PlanNode {
    name: string;
    estimated_rows: number;
    subtree_cost: number;
    node_cost: number;
    header: PlanNodeHeader;
    nodes: PlanNode[];
}

export const getStyles = (theme: GrafanaTheme2) => ({
    container: css`
        background: ${theme.colors.background.primary};
        border: 1px solid ${theme.colors.border.weak};
        border-radius: ${theme.shape.radius.default};
        padding: ${theme.spacing(3)};
        max-width: 100%;
        overflow: auto;
    `,
    header: css`
        display: flex;
        align-items: center;
        gap: ${theme.spacing(1)};
        margin-bottom: ${theme.spacing(3)};
    `,
    title: css`
        font-size: ${theme.typography.h3.fontSize};
        font-weight: ${theme.typography.h3.fontWeight};
        color: ${theme.colors.text.primary};
        margin: 0;
    `,
    section: css`
        margin-bottom: ${theme.spacing(3)};
    `,
    sectionTitle: css`
        font-size: ${theme.typography.h5.fontSize};
        font-weight: ${theme.typography.h5.fontWeight};
        color: ${theme.colors.text.primary};
        margin-bottom: ${theme.spacing(2)};
        display: flex;
        align-items: center;
        gap: ${theme.spacing(1)};
    `,
    toggleButton: css`
        background: none;
        border: none;
        color: ${theme.colors.primary.text};
        cursor: pointer;
        font-weight: 500;
        display: flex;
        align-items: center;
        gap: ${theme.spacing(1)};
        padding: ${theme.spacing(1)};
        border-radius: ${theme.shape.radius.default};

        &:hover {
            background: ${theme.colors.action.hover};
        }
    `,
    xmlPre: css`
        background: ${theme.colors.background.secondary};
        border: 1px solid ${theme.colors.border.weak};
        border-radius: ${theme.shape.radius.default};
        padding: ${theme.spacing(2)};
        font-size: ${theme.typography.bodySmall.fontSize};
        font-family: ${theme.typography.fontFamilyMonospace};
        overflow: auto;
        max-height: 400px;
        margin-top: ${theme.spacing(2)};
    `,
    planNode: css`
        border-left: 2px solid ${theme.colors.primary.border};
        padding-left: ${theme.spacing(2)};
        padding-top: ${theme.spacing(1)};
        padding-bottom: ${theme.spacing(1)};
    `,
    planNodeCard: css`
        background: ${theme.colors.background.primary};
        border: 1px solid ${theme.colors.border.weak};
        border-radius: ${theme.shape.radius.default};
        padding: ${theme.spacing(1.5)};
        box-shadow: ${theme.shadows.z1};
        transition: box-shadow 0.2s;

        &:hover {
            box-shadow: ${theme.shadows.z2};
        }
    `,
    planNodeHeader: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
    `,
    planNodeName: css`
        display: flex;
        align-items: center;
        gap: ${theme.spacing(1)};
        font-weight: 500;
        color: ${theme.colors.primary.text};
    `,
    planNodeCost: css`
        font-size: ${theme.typography.bodySmall.fontSize};
        color: ${theme.colors.text.secondary};
    `,
    planNodeGrid: css`
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: ${theme.spacing(1)};
        margin-top: ${theme.spacing(1)};
        font-size: ${theme.typography.bodySmall.fontSize};
    `,
    gridLabel: css`
        color: ${theme.colors.text.secondary};
    `,
    gridValue: css`
        color: ${theme.colors.text.primary};
        margin-left: ${theme.spacing(0.5)};
    `,
    expandButton: css`
        background: none;
        border: none;
        padding: ${theme.spacing(0.5)};
        cursor: pointer;
        display: flex;
        align-items: center;
        border-radius: ${theme.shape.radius.default};

        &:hover {
            background: ${theme.colors.action.hover};
        }
    `,
    statsCard: css`
        background: ${theme.colors.background.secondary};
        border: 1px solid ${theme.colors.border.weak};
        border-radius: ${theme.shape.radius.default};
        padding: ${theme.spacing(2)};
        margin-bottom: ${theme.spacing(1)};
    `,
    statsHeader: css`
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: ${theme.spacing(1)};
    `,
    statsTitle: css`
        font-weight: 500;
        color: ${theme.colors.text.primary};
    `,
    statsDate: css`
        font-size: ${theme.typography.bodySmall.fontSize};
        color: ${theme.colors.text.secondary};
    `,
    statsBody: css`
        font-size: ${theme.typography.bodySmall.fontSize};
        display: flex;
        flex-direction: column;
        gap: ${theme.spacing(0.5)};
    `,
    statsText: css`
        color: ${theme.colors.text.primary};
    `,
    statsDetail: css`
        color: ${theme.colors.text.secondary};
        font-size: ${theme.typography.bodySmall.fontSize};
    `,
    warningCard: css`
        background: ${theme.colors.warning.transparent};
        border: 1px solid ${theme.colors.warning.border};
        border-radius: ${theme.shape.radius.default};
        padding: ${theme.spacing(2)};
        display: flex;
        gap: ${theme.spacing(1)};
    `,
    warningIcon: css`
        color: ${theme.colors.warning.text};
        flex-shrink: 0;
        margin-top: 2px;
    `,
    warningContent: css`
        flex: 1;
    `,
    warningTitle: css`
        color: ${theme.colors.warning.text};
        font-weight: 500;
    `,
    warningDetail: css`
        color: ${theme.colors.warning.text};
        font-size: ${theme.typography.bodySmall.fontSize};
        margin-top: ${theme.spacing(0.5)};
    `,
    planTree: css`
        background: ${theme.colors.background.secondary};
        border: 1px solid ${theme.colors.border.weak};
        border-radius: ${theme.shape.radius.default};
        padding: ${theme.spacing(2)};
    `,
    childNodes: css`
        margin-top: ${theme.spacing(1)};
    `,
    warningsList: css`
        display: flex;
        flex-direction: column;
        gap: ${theme.spacing(1)};
    `,
});

// Recursive Plan Node Component
export const PlanNodeComponent: React.FC<{ node: PlanNode, level: number }> = ({node, level}) => {
    const [isExpanded, setIsExpanded] = useState(true);
    const styles = useStyles2(getStyles);

    const hasChildren = node.nodes && node.nodes.length > 0;
    const nodeStyle = {
        marginLeft: `${level * 16}px`,
    };
    return (
        <div style={nodeStyle} className={styles.planNode}>
            <div className={styles.planNodeCard}>
                <div className={styles.planNodeHeader}>
                    <div className={styles.planNodeName}>
                        {hasChildren && (
                            <button
                                onClick={() => setIsExpanded(!isExpanded)}
                                className={styles.expandButton}
                            >
                                {isExpanded ? (
                                    <ChevronDown size={16}/>
                                ) : (
                                    <ChevronRight size={16}/>
                                )}
                            </button>
                        )}
                        <div className="font-medium text-blue-800">{node.name}</div>
                    </div>
                    <div className={styles.planNodeCost}>
                        Cost: {node.node_cost.toFixed(2)}
                    </div>
                </div>

                <div className={styles.planNodeGrid}>
                    <div>
                        <span className={styles.gridLabel}>Physical Op:</span>
                        <span className={styles.gridValue}>{node.header.PhysicalOp}</span>
                    </div>
                    <div>
                        <span className={styles.gridLabel}>Logical Op:</span>
                        <span className={styles.gridValue}>{node.header.LogicalOp}</span>
                    </div>
                    <div>
                        <span className={styles.gridLabel}>Est. Rows:</span>
                        <span className={styles.gridValue}>{node.estimated_rows.toFixed(0)}</span>
                    </div>
                    <div>
                        <span className={styles.gridLabel}>Subtree Cost:</span>
                        <span className={styles.gridValue}>{node.subtree_cost.toFixed(2)}</span>
                    </div>
                    {node.header.Parallel && (
                        <div style={{gridColumn: '1 / -1'}}>
                            <span className={styles.gridLabel}>Parallel:</span>
                            <span className={styles.gridValue}>{node.header.Parallel}</span>
                        </div>
                    )}
                </div>
            </div>

            {hasChildren && isExpanded && (
                <div className={styles.childNodes}>
                    {node.nodes.map((childNode, idx) => (
                        <PlanNodeComponent key={idx} node={childNode} level={level + 1}/>
                    ))}
                </div>
            )}
        </div>
    );
};
