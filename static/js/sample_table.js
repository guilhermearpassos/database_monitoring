document.addEventListener('htmx:afterSettle', function () {
    const truncatedTexts = document.querySelectorAll('.truncated-sql-text');

    truncatedTexts.forEach((item) => {
        const textSpan = item.querySelector('.sql-text');
        const fullText = textSpan.textContent; // Get the full text from the span
        // Highlight the truncated text using Prism.js
        textSpan.innerHTML = Prism.highlight(fullText, Prism.languages.sql, 'tsql');

        // Create popover element
        const popover = document.createElement('div');
        popover.className = 'sql-popover';
        document.body.appendChild(popover);

        // Show popover on hover
        let isHovered = false;
        item.addEventListener('mouseenter', () => {
            isHovered = true;

            // Format the SQL query
            const formattedSQL = formatSQL(fullText);

            // Wrap the formatted SQL in a <pre><code> block for syntax highlighting
            popover.innerHTML = `
    <pre><code class="language-sql">${formattedSQL}</code></pre>
    <button class="sql-copy-button" title="Copy">
      <i class="fas fa-copy"></i>
    </button>
  `;

            // Highlight the SQL syntax using Prism.js
            Prism.highlightAllUnder(popover);

            // Copy text on button click
            const copyButton = popover.querySelector('.sql-copy-button');
            copyButton.addEventListener('click', () => {
                navigator.clipboard.writeText(formattedSQL).then(() => {
                    alert('Text copied to clipboard!');
                });
            });

            // Position the popover
            const rect = textSpan.getBoundingClientRect();
            popover.style.display = 'block';
            popover.style.top = `${rect.bottom + window.scrollY}px`;
            popover.style.left = `${rect.left + window.scrollX}px`;
        });

        // Hide popover when neither the cell nor the popover is hovered
        item.addEventListener('mouseleave', () => {
            isHovered = false;
            setTimeout(() => {
                if (!isHovered && !popover.matches(':hover')) {
                    popover.style.display = 'none';
                }
            }, 100); // Small delay to allow cursor to move to the popover
        });

        popover.addEventListener('mouseenter', () => {
            isHovered = true;
        });

        popover.addEventListener('mouseleave', () => {
            isHovered = false;
            setTimeout(() => {
                if (!isHovered && !item.matches(':hover')) {
                    popover.style.display = 'none';
                }
            }, 100); // Small delay to allow cursor to move back to the cell
        });
    });
});

// Function to format SQL for the popover
function formatSQL(sql) {
    try {
        // Preprocess the SQL text
        sql = preprocessSQL(sql);

        // Format the SQL using sql-formatter
        return sqlFormatter.format(sql, {
            language: 'tsql',
            indent: '  ',
            uppercase: true,
            expressionWidth: 120, // Increase the line width before breaking expressions
        });
    } catch (error) {
        console.error('SQL formatting failed:', error);
        console.error('Problematic SQL:', sql);
        return sql; // Return the original SQL if formatting fails
    }
}

// Function to preprocess SQL text
function preprocessSQL(sql) {
    // Remove extra spaces and line breaks
    sql = sql.replace(/\s+/g, ' ').trim();

    // Remove comments (optional)
    sql = sql.replace(/\/\*[\s\S]*?\*\/|--.*/g, '');

    return sql;
}