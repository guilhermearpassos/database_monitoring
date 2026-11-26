import React, { useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { Tooltip, useStyles2 } from '@grafana/ui';
import { Check, Copy, ExternalLink } from 'lucide-react';
import { useAppSettings } from '../context/AppSettingsContext';
import { copyToClipboard } from '../utils/sqlHelpers';

const getStyles = (theme: GrafanaTheme2) => ({
  wrapper: css`
    display: inline-flex;
    align-items: center;
    gap: ${theme.spacing(0.5)};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
  link: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.primary.transparent};
    color: ${theme.colors.primary.text};
    border: 1px solid ${theme.colors.primary.border};
    text-decoration: none;
    cursor: pointer;

    &:hover {
      background: ${theme.colors.primary.shade};
      color: ${theme.colors.primary.contrastText};
      text-decoration: none;
    }
  `,
  plainBadge: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.secondary};
    color: ${theme.colors.text.secondary};
    border: 1px solid ${theme.colors.border.weak};
  `,
  copyBtn: css`
    display: inline-flex;
    align-items: center;
    background: none;
    border: none;
    color: ${theme.colors.text.secondary};
    cursor: pointer;
    padding: 2px;
    border-radius: ${theme.shape.radius.default};

    &:hover {
      color: ${theme.colors.text.primary};
      background: ${theme.colors.action.hover};
    }
  `,
});

export const TempoTraceLink: React.FC<{
  traceId?: string;
  timeRange?: TimeRange;
  compact?: boolean;
}> = ({ traceId, timeRange, compact = false }) => {
  const styles = useStyles2(getStyles);
  const { buildTempoExploreUrl } = useAppSettings();
  const [copied, setCopied] = useState(false);

  if (!traceId) {
    return null;
  }

  const href = buildTempoExploreUrl(traceId, timeRange);
  const displayId = compact ? `${traceId.slice(0, 8)}…` : traceId;

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    e.preventDefault();
    const ok = await copyToClipboard(traceId);
    if (ok) {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    }
  };

  return (
    <span className={styles.wrapper} onClick={(e) => e.stopPropagation()}>
      {href ? (
        <Tooltip content={`Open trace ${traceId} in Tempo (Explore)`}>
          <a href={href} target="_blank" rel="noopener noreferrer" className={styles.link}>
            <span>{displayId}</span>
            <ExternalLink size={12} />
          </a>
        </Tooltip>
      ) : (
        <Tooltip content={`Trace ID: ${traceId} (configure a Tempo datasource in Plugin Config to enable direct links)`}>
          <span className={styles.plainBadge}>{displayId}</span>
        </Tooltip>
      )}
      <Tooltip content={copied ? 'Copied!' : 'Copy Trace ID'}>
        <button type="button" className={styles.copyBtn} onClick={handleCopy} aria-label="Copy Trace ID">
          {copied ? <Check size={12} /> : <Copy size={12} />}
        </button>
      </Tooltip>
    </span>
  );
};
