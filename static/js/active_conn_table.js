
document.addEventListener('htmx:afterSwap', function (e) {
    // Toggle the visibility of the query sample table for the clicked snapshot
    const snapshotID = e.target.id.split('-')[1]; // Get snapshot ID from the ID of the expanded row
    const expandableRow = document.getElementById('queries-' + snapshotID);
    if (expandableRow) {
        expandableRow.style.display = (expandableRow.style.display === 'none' || expandableRow.style.display === '') ? 'table-row' : 'none';
    }
});