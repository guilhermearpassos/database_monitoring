let chartInstance = null;

document.addEventListener('DOMContentLoaded', function() {
    initializeChart();
});

// Only initialize the chart when the slideover is first loaded
function initializeChart() {
    const chartContainer = document.querySelector('#chart-container');
    if (chartContainer) {
        const data = JSON.parse(chartContainer.getAttribute('data-chart') || '[]');
        const timeRange = JSON.parse(chartContainer.getAttribute('data-time-range') || '{}');
        const colormap = JSON.parse(chartContainer.getAttribute("data-color-mapping") || '{}');
        
        createChart(data, timeRange, colormap);
    }
}

// This will ONLY respond to slideover being loaded initially, not to other HTMX events
document.addEventListener('htmx:afterSettle', function(evt) {
    // Only create chart when the slideover is first loaded
    if (evt.detail.target.id === 'slideover-wrapper') {
        initializeChart();
    }
});

// Preserve your existing color mapping function
function getRandomColor(waitGroup, colormap) {
    if (waitGroup in colormap){
        return colormap[waitGroup];
    }
    return '#' + Math.floor(Math.random()*16777215).toString(16).padStart(6, '0');
}

function createChart(data, timeRange, colormap) {
    // Handle empty data
    if (!data || data.length === 0) {
        document.getElementById('chart').innerHTML = '<div class="no-data-message">No data available for the selected time range</div>';
        return;
    }
    
    // Sort data by timestamp to ensure correct order
    data.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
    
    // Get all unique wait groups across all data points
    const allWaitGroups = new Set();
    data.forEach(item => {
        if (item.wait_groups) {
            Object.keys(item.wait_groups).forEach(group => allWaitGroups.add(group));
        }
    });
    const waitGroups = Array.from(allWaitGroups);
    
    // Create series for ApexCharts - Each series represents a wait group
    const series = waitGroups.map(waitGroup => ({
        name: waitGroup,
        data: data.map(item => ({
            x: new Date(item.timestamp).getTime(),
            y: item.wait_groups && item.wait_groups[waitGroup] ? item.wait_groups[waitGroup] : 0
        }))
    }));
    
    // Get colors for each wait group
    const colors = waitGroups.map(waitGroup => getRandomColor(waitGroup, colormap));
    
    // Chart options for stacked bar chart
    const options = {
        series: series,
        colors: colors,
        chart: {
            type: 'bar', // Changed to bar chart
            height: 350,
            stacked: true,
            toolbar: {
                show: true,
                tools: {
                    download: true,
                    selection: true,
                    zoom: true,
                    zoomin: true,
                    zoomout: true,
                    pan: true,
                    reset: true
                }
            },
            animations: {
                enabled: true,
                speed: 500,
                dynamicAnimation: {
                    enabled: true
                }
            },
            zoom: {
                enabled: true
            }
        },
        plotOptions: {
            bar: {
                horizontal: false,
                columnWidth: '70%', // Adjust bar width
                borderRadius: 2,    // Slightly rounded corners
                dataLabels: {
                    position: 'center',
                    maxItems: 100
                }
            }
        },
        dataLabels: {
            enabled: false
        },
        stroke: {
            width: 1,
            colors: ['#fff']
        },
        grid: {
            borderColor: '#e0e0e0',
            row: {
                colors: ['transparent']
            }
        },
        legend: {
            position: 'top',
            horizontalAlign: 'left',
            offsetY: 10,
            fontSize: '13px',
            markers: {
                radius: 2
            }
        },
        xaxis: {
            type: 'datetime',
            min: new Date(timeRange.start).getTime(),
            max: new Date(timeRange.end).getTime(),
            labels: {
                datetimeUTC: false,
                format: 'HH:mm'
            },
            axisTicks: {
                show: true
            },
            axisBorder: {
                show: true
            }
        },
        yaxis: {
            labels: {
                formatter: function(val) {
                    return Math.round(val);
                }
            },
            title: {
                text: 'Number of Connections'
            }
        },
        tooltip: {
            shared: true,
            intersect: false,
            y: {
                formatter: function(val) {
                    return val + " connections";
                }
            },
            x: {
                format: 'MMM dd, yyyy HH:mm:ss'
            }
        },
        title: {
            text: 'Database Connections by Wait Group',
            align: 'left',
            style: {
                fontSize: '16px',
                fontWeight: 'bold'
            }
        },
        responsive: [{
            breakpoint: 768,
            options: {
                legend: {
                    position: 'bottom',
                    offsetY: 0
                }
            }
        }]
    };

    // Destroy existing chart if it exists
    if (chartInstance) {
        chartInstance.destroy();
    }

    // Create new ApexCharts instance
    chartInstance = new ApexCharts(document.querySelector("#chart"), options);
    chartInstance.render();
}