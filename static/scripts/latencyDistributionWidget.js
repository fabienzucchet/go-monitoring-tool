/************************************************************/
/*        JAVASCRIPT FOR LATENCY DISTRIBUTION WIDGET        */
/************************************************************/

/* function to generate the histogram data */
function generateHistogramData(responseTimeList) {
    data = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];

    for (const el of responseTimeList) {
        if (el.y < 50) {
            data[0] = data[0] + 1;
        } else if (el.y < 100) {
            data[1] = data[1] + 1;
        } else if (el.y < 200) {
            data[2] = data[2] + 1;
        } else if (el.y < 300) {
            data[3] = data[3] + 1;
        } else if (el.y < 500) {
            data[4] = data[4] + 1;
        } else if (el.y < 700) {
            data[5] = data[5] + 1;
        } else if (el.y < 1000) {
            data[6] = data[6] + 1;
        } else if (el.y < 2000) {
            data[7] = data[7] + 1;
        } else if (el.y < 3000) {
            data[8] = data[8] + 1;
        } else if (el.y < 5000) {
            data[9] = data[9] + 1;
        } else {
            data[10] = data[10] + 1;
        }
    }

    return data;
}

/* Fetch the chart data */
function updateLatencyDistributionChart(chart, duration) {
    fetch("/metrics/latency?" + new URLSearchParams({duration: duration})).then(async (resp) => {
        const dataMap = await resp.json();
        const alreadyPlotted = chart.data.datasets.map(dataset => dataset.label);

        for (const target in dataMap) {
            const idx = alreadyPlotted.indexOf(target);
            if (idx >= 0) {
                chart.data.datasets[idx].data = generateHistogramData(dataMap[target]);
            } else {
                chart.data.datasets.push({
                    label: target,
                    data: generateHistogramData(dataMap[target]),
                    backgroundColor: [
                        'rgba(255, 99, 132, 0.2)',
                        'rgba(255, 159, 64, 0.2)',
                        'rgba(255, 205, 86, 0.2)',
                        'rgba(75, 192, 192, 0.2)',
                        'rgba(54, 162, 235, 0.2)',
                        'rgba(153, 102, 255, 0.2)',
                        'rgba(179, 157, 207, 0.2)',
                        'rgba(201, 203, 207, 0.2)',
                        'rgba(53, 203, 64, 0.2)',
                        'rgba(153, 64, 86, 0.2)',
                        'rgba(179, 99, 86, 0.2)',
                    ],
                    borderColor: [
                        'rgb(255, 99, 132)',
                        'rgb(255, 159, 64)',
                        'rgb(255, 205, 86)',
                        'rgb(75, 192, 192)',
                        'rgb(54, 162, 235)',
                        'rgb(153, 102, 255)',
                        'rgb(179, 157, 207)',
                        'rgb(201, 203, 207)',
                        'rgb(53, 203, 64)',
                        'rgb(153, 64, 86)',
                        'rgb(179, 99, 86)',
                        ],
                    borderWidth: 1
                });
            }
        }

        // Update the chart
        chart.update();

    })
}

/* Init the latency distribution chart */
function initLatencyDistributionChart() {
    var ctx = document.getElementById('latency-distribution-chart');
    var latencyDistributionChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: ['<50ms', '50-100ms', '100-200ms', '200-300ms', '300-500ms', '500-700ms', '0.7-1s', '1-2s','2-3s', '3-5s', '>5s'],
            datasets: []
        },
        options: {
            scales: {
                y: {
                    beginAtZero: true
                }
            },
            responsive: true,
        },
    });

    return latencyDistributionChart;
}

/* Instantiate the latency distribution chart */
var latencyDistributionChart = initLatencyDistributionChart();
updateLatencyDistributionChart(latencyDistributionChart, "-10m");

/* Refresh periodically the chart */
setInterval(() => {
    updateLatencyDistributionChart(latencyDistributionChart, "-10m");
}, 10 * 1000) // every 10 sec
