// Format SQL with sql-formatter library
import {format} from "sql-formatter";
import Prism from 'prismjs';
import 'prismjs/themes/prism-tomorrow.css';
import 'prismjs/components/prism-sql';

export const formatSQL = (sql: string): string => {
    if (!sql) {
        return '';
    }

    // Remove extra spaces and line breaks
    sql = sql.replace(/\s+/g, ' ').trim();

    // Remove comments (optional)
    sql = sql.replace(/\/\*[\s\S]*?\*\/|--.*/g, '');
    try {
        let s = format(sql, {
            language: 'tsql', // Use 'tsql' for SQL Server, 'sql' for generic, 'postgresql', 'mysql', etc.
            tabWidth: 2,
            keywordCase: 'upper',
            linesBetweenQueries: 2,
            expressionWidth: 120, // Increase the line width before breaking expressions
        });
        const highlighted = Prism.highlight(s, Prism.languages.sql, 'sql');
        return highlighted;
    } catch (error) {
        // If formatting fails, return original
        console.error('SQL formatting error:', error);
        return sql;
    }
};
