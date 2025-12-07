// samples_performance.js - Specific optimizations for samples table
document.addEventListener('DOMContentLoaded', initSamplesTableOptimizations);
document.addEventListener('htmx:afterSettle', initSamplesTableOptimizations);

function initSamplesTableOptimizations() {
    // Find all samples tables
    const samplesTables = document.querySelectorAll('.samples-table, table:has(th:contains("Query"))');

    samplesTables.forEach(table => {
        // 1. Optimize rendering of query columns
        const queryColumns = table.querySelectorAll('tr td:nth-child(2)');
        queryColumns.forEach(cell => {
            // Check if we need to process this cell
            if (cell.dataset.processed) {
                return;
            }

            const originalText = cell.textContent;

            // Only process if the text is long
            if (originalText.length > 100) {
                // Store full text in data attribute
                cell.dataset.fullText = originalText;

                // Truncate text
                cell.textContent = originalText.substring(0, 100) + '...';

                // Add tooltip behavior
                cell.classList.add('has-tooltip');
                cell.title = "Hover to see full text";

                // Mark as processed
                cell.dataset.processed = true;
            }
        });

        // 2. Add smart event delegation for tooltips and expansions
        if (!table.dataset.optimized) {
            table.addEventListener('mouseover', function(e) {
                const cell = e.target.closest('td.has-tooltip');
                if (cell && cell.dataset.fullText) {
                    // Show tooltip with full text on hover
                    showDetailedTooltip(cell, cell.dataset.fullText);
                }
            });

            // Mark table as optimized
            table.dataset.optimized = true;
        }
    });
}

function showDetailedTooltip(element, text) {
    // Reuse existing tooltip element
    let tooltip = document.getElementById('detailed-tooltip');
    if (!tooltip) {
        tooltip = document.createElement('div');
        tooltip.id = 'detailed-tooltip';
        tooltip.className = 'detailed-tooltip';
        document.body.appendChild(tooltip);

        // Add styles
        const style = document.createElement('style');
        style.textContent = `
            .detailed-tooltip {
                position: fixed;
                z-index: 9999;
                background: white;
                border: 1px solid #ccc;
                padding: 10px;
                border-radius: 4px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.15);
                max-width: 50vw;
                max-height: 30vh;
                overflow: auto;
                white-space: pre-wrap;
                font-family: monospace;
                font-size: 13px;
                display: none;
            }
        `;
        document.head.appendChild(style);
    }

    // Set content
    tooltip.textContent = text;
    tooltip.style.display = 'block';

    // Position tooltip near the element
    const rect = element.getBoundingClientRect();
    tooltip.style.left = `${Math.min(rect.left, window.innerWidth - 400)}px`;
    tooltip.style.top = `${rect.bottom + 5}px`;

    // Hide tooltip on mouse leave
    const hideTooltip = function() {
        tooltip.style.display = 'none';
        element.removeEventListener('mouseleave', hideTooltip);
    };

    element.addEventListener('mouseleave', hideTooltip);
}
