import React, {useState} from "react";
import {ChevronDown, ChevronRight, AlertTriangle, Info} from 'lucide-react';
import {PlanNodeComponent, PlanNode, getStyles} from "./planNode"
import {useStyles2, useTheme2} from "@grafana/ui";

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
    database: string
    schema: string
    table: string
    impact: number
    equality_columns?: string[]
    inequality_columns?: string[]
    include_columns?: string[]
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

// Statistics Info Component
const StatisticsInfoComponent: React.FC<{ stats: StatisticsInfo }> = ({stats}) => {
    const styles = useStyles2(getStyles);
    return (
        <div className={styles.statsCard}>
            <div className={styles.statsHeader}>
                <div className={styles.statsTitle}>{stats.table}</div>
                <div className={styles.statsDate}>Last Update: {stats.last_update}</div>
            </div>
            <div className={styles.statsBody}>
                <div className={styles.statsText}>Modifications: {stats.modification_count}</div>
                <div className={styles.statsText}>Sampling: {stats.sampling_percent.toFixed(2)}%</div>
                <div className={styles.statsDetail}>{stats.statistics}</div>
            </div>
        </div>
    );
};

// Warning Component
const WarningComponent: React.FC<{ warning: PlanWarning }> = ({warning}) => {
    const styles = useStyles2(getStyles);
    if ((!warning.convert) && !warning.missing_index) {
        return null;
    }

    return (
        <div className={styles.warningCard}>
            <div className={styles.warningIcon}>
                <AlertTriangle size={20}/>
            </div>
            {warning.convert &&
                <div className={styles.warningContent}>
                    <div className={styles.warningTitle}>Convert Issue: {warning.convert.convert_issue}</div>
                    <div className={styles.warningDetail}>Expression: {warning.convert.expression}</div>
                </div>}
            {warning.missing_index &&
                <div className={styles.warningContent}>
                    <div className={styles.warningTitle}>Missing index: {warning.missing_index.table}</div>
                    <div className={styles.warningDetail}>Expression: create index idx01 on {warning.missing_index.database}.{warning.missing_index.schema}.{warning.missing_index.table}({((warning.missing_index.equality_columns??[]).concat(warning.missing_index.inequality_columns??[])).join(",")}) include ({(warning.missing_index.include_columns??[]).join(",")})</div>
                </div>
            }
        </div>
    );
};


// Main Execution Plan Viewer Component
export const ExecutionPlanViewer: React.FC<{ executionPlan: ParsedExecutionPlan }> = ({executionPlan}) => {
    const [showRawXml, setShowRawXml] = useState(false);
    const styles = useStyles2(getStyles);
    const theme = useTheme2();

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <Info size={24} color={theme.colors.primary.text}/>
                <h3 className={styles.title}>Execution Plan</h3>
            </div>

            {/* Raw XML Plan Toggle */}
            <div className={styles.section}>
                <button
                    className={styles.toggleButton}
                    onClick={() => setShowRawXml(!showRawXml)}
                >
                    {showRawXml ? <ChevronDown size={16}/> : <ChevronRight size={16}/>}
                    Toggle Raw XML Plan
                </button>
                {showRawXml && (
                    <pre className={styles.xmlPre}>
            {executionPlan.plan.xml_plan}
          </pre>
                )}
            </div>

            {/* Plan Warnings */}
            {executionPlan.warnings && executionPlan.warnings.length > 0 && (
                <div className={styles.section}>
                    <h4 className={styles.sectionTitle}>
                        <AlertTriangle size={20} color={theme.colors.warning.text}/>
                        Warnings
                    </h4>
                    <div className={styles.warningsList}>
                        {executionPlan.warnings.map((warning, idx) => (
                            <WarningComponent key={idx} warning={warning}/>
                        ))}
                    </div>
                </div>
            )}

            {/* Statistics Usage */}
            {executionPlan.stats_usage && executionPlan.stats_usage.length > 0 && (
                <div className={styles.section}>
                    <h4 className={styles.sectionTitle}>Statistics Usage</h4>
                    <div className="space-y-2">
                        {executionPlan.stats_usage.map((stats, idx) => (
                            <StatisticsInfoComponent key={idx} stats={stats}/>
                        ))}
                    </div>
                </div>
            )}

            {/* Plan Tree */}
            <div className={styles.section}>
                <h4 className={styles.sectionTitle}>Plan Tree</h4>
                <div className={styles.planTree}>
                    {executionPlan.nodes && executionPlan.nodes.map((node, idx) => (
                        <PlanNodeComponent key={idx} node={node} level={1}/>
                    ))}
                </div>
            </div>
        </div>
    );
};
