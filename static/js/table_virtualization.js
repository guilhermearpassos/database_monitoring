// table_virtualization.js - Add progressive loading for large tables
document.addEventListener('DOMContentLoaded', initializeTableVirtualization);
document.addEventListener('htmx:afterSettle', initializeTableVirtualization);

function initializeTableVirtualization() {
    const samplesTable = document.querySelector('#snapshots-table, .samples-table');
    if (!samplesTable) {
        return;
    }

    const tbody = samplesTable.querySelector('tbody');
    if (!tbody) {
        return;
    }

    const rows = Array.from(tbody.querySelectorAll('tr:not(.expandable-row)'));
    if (rows.length <= 50) {
        return; // Only apply for larger tables
    }

    // Find the container or create a wrapper
    let container = samplesTable.closest('.table-container');
    if (!container) {
        // Create a container
        container = document.createElement('div');
        container.className = 'table-container';
        container.style.maxHeight = '600px';
        container.style.overflowY = 'auto';
        samplesTable.parentNode.insertBefore(container, samplesTable);
        container.appendChild(samplesTable);
    }

    // Setup scroll optimization
    let scrollTimeout;
    container.addEventListener('scroll', function () {
        if (!container.classList.contains('is-scrolling')) {
            container.classList.add('is-scrolling');
        }

        clearTimeout(scrollTimeout);
        scrollTimeout = setTimeout(() => {
            container.classList.remove('is-scrolling');
        }, 150);

        const visibleTop = container.scrollTop;
        const visibleBottom = visibleTop + container.clientHeight;

        // Mark rows as visible or hidden based on their position
        let rowTop = 0;
        rows.forEach((row, index) => {
            const rowHeight = row.clientHeight || 53; // Default height if not available
            const rowBottom = rowTop + rowHeight;

            // Add buffer rows above and below viewport
            const isVisible = (rowBottom >= visibleTop - 300) && (rowTop <= visibleBottom + 300);

            // Handle expanded rows associated with this row
            const rowId = row.getAttribute('data-id') || row.id;
            const expandedRows = rowId ? tbody.querySelectorAll(`.expandable-row[data-parent="${rowId}"]`) : [];

            if (isVisible) {
                row.classList.remove('virtualized-hidden');
                Array.from(expandedRows).forEach(expRow => {
                    if (expRow.classList.contains('show')) {
                        expRow.classList.remove('virtualized-hidden');
                    }
                });
            } else {
                row.classList.add('virtualized-hidden');
                Array.from(expandedRows).forEach(expRow => {
                    expRow.classList.add('virtualized-hidden');
                });
            }

            rowTop += rowHeight;
            // Add height of any expanded rows
            Array.from(expandedRows).forEach(expRow => {
                if (expRow.classList.contains('show')) {
                    rowTop += expRow.clientHeight || 100;
                }
            });
        });
    });

    // Add needed CSS class
    const style = document.createElement('style');
    style.textContent = `
        .virtualized-hidden {
            display: none !important;
        }
        .table-container {
            will-change: transform;
            transform: translateZ(0);
        }
    `;
    document.head.appendChild(style);

    // Initial scroll trigger
    setTimeout(() => container.dispatchEvent(new Event('scroll')), 100);
}
