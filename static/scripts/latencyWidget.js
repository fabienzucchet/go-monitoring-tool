/************************************************************/
/*              JAVASCRIPT FOR LATENCY WIDGET               */
/************************************************************/

const COLORS = [
    'rgb(255, 99, 132)',
    'rgb(255, 159, 64)',
    'rgb(255, 205, 86)',
    'rgb(75, 192, 192)',
    'rgb(54, 162, 235)',
    'rgb(153, 102, 255)',
    'rgb(179, 157, 207)',
    'rgb(201, 203, 207)'
]

var COLOR_IDX = 0;

/* Fetch the latency data */
function updateLatencyChart(chart, duration) {
    fetch("/metrics/latency?" + new URLSearchParams({duration: duration})).then(async (resp) => {
        const dataMap = await resp.json();
        const alreadyPlotted = chart.data.datasets.map(dataset => dataset.label);

        for (const target in dataMap) {
            // If the target is already plotted, we only update the data, else create a dataset object
            const idx = alreadyPlotted.indexOf(target);
            if (idx >= 0) {
                chart.data.datasets[idx].data = dataMap[target];
            } else {
                chart.data.datasets.push({
                    label: target,
                    data: dataMap[target],
                    showLine: true,
                    borderColor: COLORS[COLOR_IDX]
                });
                COLOR_IDX = COLOR_IDX + 1;
            }
        }

        // Update the chart
        chart.update();
    });
}

/* Init the latency chart */
function initLatencyChart() {
    var ctx = document.getElementById('latency-chart');
    var latencyChart = new Chart(ctx, {
        type: 'scatter',
        data: {
            datasets: []
        },
        options: {
            responsive: true,
            scales: {
                x: {
                    ticks: {
                        callback: function(value) {
                            const date = new Date(value * 1000);
                            return date.toUTCString().slice(-11, -4);
                        }
                    }
                },
                y: {
                    ticks: {
                        callback: function(value) {
                            return `${value / 1000} s`;
                        }
                    }
                }
            }
        }
    });
    
    return latencyChart;
}

/* Instantiate the latency chart */
var latencyChart = initLatencyChart();
updateLatencyChart(latencyChart, "-10m");

/* Refresh periodically the chart */
setInterval(() => {
    updateLatencyChart(latencyChart, "-10m");
}, 10 * 1000) // every 10 sec
