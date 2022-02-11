package server

const (
	metricsTemplate = `<!DOCTYPE html>
<html lang="en">
<body>
<table>
    <tr>
        <th>Type</th>
        <th>Name</th>
        <th>Value</th>
    </tr>
    {{ range $key, $value := .GaugeMetrics -}}
    <tr>
        <td style='text-align:center; vertical-align:middle'>Gauge</td>
        <td style='text-align:center; vertical-align:middle'>{{ $key }}</td>
        <td style='text-align:center; vertical-align:middle'>{{ $value }}</td>
    </tr>
    {{ end -}}
    {{ range $key, $value := .CounterMetrics -}}
    <tr>
        <td style='text-align:center; vertical-align:middle'>Counter</td>
        <td style='text-align:center; vertical-align:middle'>{{ $key }}</td>
        <td style='text-align:center; vertical-align:middle'>{{ $value }}</td>
    </tr>
    {{ end -}}
</table>
</body>
</html>
`
)
