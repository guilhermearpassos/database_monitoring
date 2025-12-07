// Initialize Lucide icons
lucide.createIcons();

// Refresh control variables
let refreshInterval = 5;
let refreshTimer = null;
let isPlaying = true;

// Function to refresh the data
function refreshData() {
    document.getElementById('server-table-body').setAttribute('hx-get', '/servers');
    document.getElementById('server-table-body').setAttribute('hx-trigger', 'load');
}

// Function to start/stop auto-refresh
function startAutoRefresh() {
    if (refreshTimer) {
        clearInterval(refreshTimer);
    }
    if (isPlaying && refreshInterval > 0) {
        refreshTimer = setInterval(refreshData, refreshInterval * 1000);
    }
}

// Toggle play/pause
function togglePlayPause() {
    isPlaying = !isPlaying;
    const icon = document.getElementById('playPauseIcon');
    icon.setAttribute('data-lucide', isPlaying ? 'pause' : 'play');
    lucide.createIcons();

    if (isPlaying) {
        startAutoRefresh();
    } else {
        clearInterval(refreshTimer);
    }
}

// Toggle refresh dropdown
function toggleRefreshDropdown() {
    const dropdown = document.getElementById('refreshDropdown');
    dropdown.classList.toggle('active');
}

// Handle refresh interval selection
document.querySelectorAll('.refresh-option').forEach(option => {
    option.addEventListener('click', () => {
        const interval = parseInt(option.dataset.interval);
        refreshInterval = interval;

        document.getElementById('refreshInterval').textContent = interval === 0 ? 'Auto (Off)' :
            interval >= 60 ? 'Auto (' + (interval / 60) + 'm)' :
                'Auto (' + interval + 's)';

        document.querySelectorAll('.refresh-option').forEach(opt => opt.classList.remove('active'));
        option.classList.add('active');

        document.getElementById('refreshDropdown').classList.remove('active');

        startAutoRefresh();
    });
});

// Start auto-refresh on page load
startAutoRefresh();
