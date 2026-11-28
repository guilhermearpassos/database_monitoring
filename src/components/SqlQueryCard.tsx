import React, { useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { Button, Card, useStyles2 } from '@grafana/ui';
import { Check, Copy } from 'lucide-react';
import { formatSQL } from '../utils/formatters';
import { copyToClipboard, formatSQLPlain } from '../utils/sqlHelpers';

const getStyles = (theme: GrafanaTheme2) => ({
  sqlBox: css`
    position: relative;
    background: ${theme.colors.background.canvas};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.5)};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 320px;
    overflow: auto;
    line-height: 1.5;
  `,
});

export const SqlQueryCard: React.FC<{
  sql: string;
  heading?: string;
  className?: string;
}> = ({ sql, heading = 'SQL Query', className }) => {
  const styles = useStyles2(getStyles);
  const [copied, setCopied] = useState(false);

  const copyTimerRef = React.useRef<number | null>(null);
  React.useEffect(() => {
    return () => {
      if (copyTimerRef.current !== null) {
        window.clearTimeout(copyTimerRef.current);
      }
    };
  }, []);

  const handleCopy = async () => {
    const ok = await copyToClipboard(formatSQLPlain(sql));
    if (ok) {
      setCopied(true);
      if (copyTimerRef.current !== null) {
        window.clearTimeout(copyTimerRef.current);
      }
      copyTimerRef.current = window.setTimeout(() => setCopied(false), 1800);
    }
  };

  return (
    <Card className={className}>
      <Card.Heading>{heading}</Card.Heading>
      <Card.Actions>
        <Button size="sm" variant="secondary" fill="outline" onClick={handleCopy}>
          {copied ? <Check size={13} /> : <Copy size={13} />}
          <span style={{ marginLeft: 4 }}>{copied ? 'Copied' : 'Copy Formatted SQL'}</span>
        </Button>
      </Card.Actions>
      <Card.Description>
        <div className={styles.sqlBox} dangerouslySetInnerHTML={{ __html: formatSQL(sql) }} />
      </Card.Description>
    </Card>
  );
};
