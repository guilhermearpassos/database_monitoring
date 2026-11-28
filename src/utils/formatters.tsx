// Format SQL with sql-formatter library
import { format } from 'sql-formatter';
import Prism from 'prismjs';
import 'prismjs/themes/prism-tomorrow.css';
import 'prismjs/components/prism-sql';

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

export const formatSQL = (sql: string): string => {
  if (!sql) {
    return '';
  }

  const trimmed = sql.trim();
  try {
    const s = format(trimmed, {
      language: 'tsql',
      tabWidth: 2,
      keywordCase: 'upper',
      linesBetweenQueries: 2,
      expressionWidth: 120,
    });
    return Prism.highlight(s, Prism.languages.sql, 'sql');
  } catch (error) {
    // If formatting fails, safely escape HTML so raw SQL can never inject markup/scripts.
    console.error('SQL formatting error:', error);
    return escapeHtml(trimmed);
  }
};
