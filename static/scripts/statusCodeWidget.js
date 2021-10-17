/************************************************************/
/*            JAVASCRIPT FOR STATUS CODE WIDGET             */
/************************************************************/

/* Group the status codes in categories */
function groupStatusCodes(dataMap) {

    const data = [0, 0, 0, 0];

    for (const target in dataMap) {
        const targetData = dataMap[target];
        for (const statusCode in targetData) {
            if (statusCode >= 200 && statusCode < 300) {
                data[0] = data[0] + targetData[statusCode];
            } else if (statusCode >= 400 && statusCode < 500) {
                data[1] = data[1] + targetData[statusCode];
            } else if (statusCode >= 500 && statusCode < 600) {
                data[2] = data[2] + targetData[statusCode];
            } else {
                data[3] = data[3] + targetData[statusCode];
            }
        }
    }

    return data;
}

/* Fetch the status code data */
function updateStatusCodeChart(chart, duration) {
    fetch("/metrics/httpstatus?" + new URLSearchParams({duration: duration})).then(async (resp) => {
        const dataMap = await resp.json();
        
        chart.data.datasets[0].data = groupStatusCodes(dataMap);

        // Update the chart
        chart.update();
    })
}

/* Init the status code chart */
function initStatusCodeChart() {
    var ctx = document.getElementById('status-code-chart');
    var statusCodeChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: ['2xx', '4xx', '5xx', 'Other'],
            datasets: [{
                label: 'HTTP Errors',
                data: [0, 0, 0, 0],
                backgroundColor: [
                'rgb(54, 162, 235)',
                'rgb(255, 205, 86)',
                'rgb(255, 99, 132)',
                'rgb(255, 159, 64)',
                ],
                hoverOffset: 4
            }]
        },
        options: {
            responsive: true,
        }
    });

    return statusCodeChart;
}

/* Instantiate the status code chart */
var statusCodeChart = initStatusCodeChart();
updateStatusCodeChart(statusCodeChart, "-10m");

/* Refresh periodically the chart */
setInterval(() => {
    updateStatusCodeChart(statusCodeChart, "-10m");
}, 10 * 1000) // every 10 sec
