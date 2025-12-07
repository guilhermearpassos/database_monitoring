function toggleDropdown() {
    const dropdown = document.getElementById('timerange-dropdown');
    dropdown.classList.toggle('active');
}

// Improved event listener with proper delegation
function setupTimerangeListeners() {
    // Get all dropdown options and attach click handlers
    document.querySelectorAll('.dropdown-option').forEach(option => {
        option.addEventListener('click', function() {
            const value = this.dataset.value;
            document.getElementById('selected-timerange-label').innerText = this.innerText;
            document.getElementById("selected-timerange").value = this.innerText;
            document.getElementById('timerange-dropdown').classList.remove('active');

            // Update active state
            document.querySelectorAll('.dropdown-option').forEach(opt => opt.classList.remove('active'));
            this.classList.add('active');
            
            // Trigger change event
            const changeEvent = new Event('change');
            document.getElementById('selected-timerange').dispatchEvent(changeEvent);
        });
    });
}

// Ensure listeners are set up when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    setupTimerangeListeners();
});

// Add listener for dynamically loaded content
document.addEventListener('htmx:afterSettle', function() {
    setupTimerangeListeners();
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
        
        // Trigger change event
        const changeEvent = new Event('change');
        document.getElementById('selected-timerange').dispatchEvent(changeEvent);
    } else {
        alert('Please select both start and end times.');
    }
}

// Close dropdown when clicking outside
document.addEventListener('click', (e) => {
    const picker = document.getElementById('timerange-picker');
    if (picker && !picker.contains(e.target)) {
        const dropdown = document.getElementById('timerange-dropdown');
        if (dropdown) {
            dropdown.classList.remove('active');
        }
    }
});
