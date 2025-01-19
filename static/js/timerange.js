
document.addEventListener('htmx:afterSwap', function() {
    document.querySelectorAll('.bar-segment').forEach(function(segment) {
        const percent = segment.getAttribute('data-percent');
        if (percent) {
            segment.style.transition = 'width 0.5s ease-in-out';
        }
    });
});

function toggleDropdown() {
    const dropdown = document.getElementById('timerange-dropdown');
    dropdown.classList.toggle('active');
}

document.querySelectorAll('.dropdown-option').forEach(option => {
    option.addEventListener('click', () => {
        const value = option.dataset.value;
        document.getElementById('selected-timerange-label').innerText = option.innerText;
        document.getElementById("selected-timerange").value = option.innerText;
        document.getElementById('timerange-dropdown').classList.remove('active');

        // Update active state
        document.querySelectorAll('.dropdown-option').forEach(opt => opt.classList.remove('active'));
        option.classList.add('active');
        document.getElementById('selected-timerange').dispatchEvent(new Event('change'));
    });
});

function applyCustomRange() {
    const startTime = document.getElementById('start-time').value;
    const endTime = document.getElementById('end-time').value;

    if (startTime && endTime) {
        const formattedStart = new Date(startTime).toISOString();
        const formattedEnd = new Date(endTime).toISOString();
        const rangeText = formattedStart + ' - ' + formattedEnd;

        document.getElementById('selected-timerange-label').innerText = rangeText;
        document.getElementById("selected-timerange").value = rangeText;
        document.getElementById('timerange-dropdown').classList.remove('active');
        document.getElementById('selected-timerange').dispatchEvent(new Event('change'));

        // You can add HTMX trigger here if needed
        // document.getElementById('timerange-display').setAttribute('hx-vals',
        //     JSON.stringify({start: startTime, end: endTime}));
        // document.getElementById('timerange-display').click();
    } else {
        alert('Please select both start and end times.');
    }
}

// Close dropdown when clicking outside
document.addEventListener('click', (e) => {
    const picker = document.getElementById('timerange-picker');
    if (!picker.contains(e.target)) {
        document.getElementById('timerange-dropdown').classList.remove('active');
    }
})