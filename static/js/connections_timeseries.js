
let chartInstance = null;

document.addEventListener('htmx:afterSettle', function(evt) {
    const chartContainer = document.querySelector('#chart-container');
    if (chartContainer) {
        const data = JSON.parse(chartContainer.getAttribute('data-chart'));
        const timeRange = JSON.parse(chartContainer.getAttribute('data-time-range'));
        createChart(data, timeRange);
    }
});

function createChart(data, timeRange) {
    // Sort data by timestamp to ensure correct order
    data.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
    console.log(new Date(timeRange.start))
    const waitGroups = Object.keys(data[0]?.wait_groups || {});
    const datasets = waitGroups.map(waitGroup => {
        const color = getRandomColor(waitGroup);
        return {
            label: waitGroup,
            data: data.map(item => ({
                x: new Date(item.timestamp),
                y: item.wait_groups[waitGroup]
            })),
            backgroundColor: color,
            borderColor: color,
            stack: 'stack1'
        };
    });

    const ctx = document.getElementById('chart').getContext('2d');

    if (chartInstance) {
        chartInstance.destroy();
    }

    chartInstance = new Chart(ctx, {
        type: 'bar',
        data: {
            datasets: datasets
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    type: 'time',
                    // time: {
                    //     unit: 'hour',
                    //     displayFormats: {
                    //         hour: 'MMM d, HH:mm'
                    //     }
                    // },
                    stacked: true,
                    min: new Date(timeRange.start),
                    max: new Date(timeRange.end),
                    grid: {
                        display: true
                    }
                },
                y: {
                    stacked: true,
                    beginAtZero: true
                }
            },
            plugins: {
                legend: {
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Number of Database Connections per Wait Group Over Time'
                }
            }
        }
    });
}

const colorMapping = {
    "PAGELATCH_EX": "#b91c1c",
    "PAGELATCH_SH": "#ef4444",
    "none":         "#6ee7b7",}

function getRandomColor(waitGroup) {
    if (waitGroup in colorMapping){
        return colorMapping[waitGroup];
    }
    return '#' + Math.floor(Math.random()*16777215).toString(16).padStart(6, '0');
}