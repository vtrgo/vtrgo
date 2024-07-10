// metricsChart.js

document.addEventListener('DOMContentLoaded', function() {
    const ctx = document.getElementById('metricsChart').getContext('2d');

    const metricsChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'PLC Metrics',
                data: [],
                borderColor: 'rgba(75, 192, 192, 1)',
                borderWidth: 2,
                fill: false
            }]
        },
        options: {
            responsive: true,
            scales: {
                x: {
                    display: true,
                    title: {
                        display: true,
                        text: 'Time'
                    }
                },
                y: {
                    display: true,
                    title: {
                        display: true,
                        text: 'Value'
                    }
                }
            }
        }
    });

    function updateChartData(data) {
        metricsChart.data.labels.push(new Date().toLocaleTimeString());
        metricsChart.data.datasets[0].data.push(data);
        metricsChart.update();
    }

    setInterval(() => {
        fetch('/metrics/latest')
            .then(response => response.json())
            .then(data => {
                updateChartData(data.value);
            })
            .catch(error => console.error('Error fetching metrics data:', error));
    }, 5000);
});
