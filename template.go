package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strconv"
)

var templateHTML = `<html>
<head>
    <meta charset="UTF-8">
    <meta name="robots" content="noindex,nofollow,noarchive,nosnippet,noodp,noydir">
    <title>{{.Title}}</title>
    <link href="https://ajax.googleapis.com/ajax/static/modules/gviz/1.0/core/tooltip.css" rel="stylesheet" type="text/css">
</head>
<body>
    <script type="text/javascript" src="https://www.google.com/jsapi"></script>
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
    <script type="text/javascript">
        $(window).resize(function() {
            drawChart();
        });
        // Load the Visualization API and the chart package.
        google.load('visualization', '1.0', {
            'packages': ['corechart']
        });
        // Set a callback to run when the Google Visualization API is loaded.
        google.setOnLoadCallback(drawChart);
        // Callback that creates and populates a data table,
        // instantiates the chart, passes in the data and
        // draws it.
        function drawChart() {
			// Create the data table.
            var data = google.visualization.arrayToDataTable({{.Data}});
            // Set chart options
            var options = {
                title: '{{.Title}}',
                curveType: 'function',
                legend: {
                    position: 'bottom'
				},
				interpolateNulls: true,
				vAxis: {title: "Average in " + {{.Type}}},
    			hAxis: {title: "Nb Request"}
            };
            // Instantiate and draw our chart, passing in some options.
            var chart = new google.visualization.LineChart(document.getElementById('chart_div'));
            chart.draw(data, options);
        }
    </script>
    <script src="https://www.google.com/uds/?file=visualization&amp;v=1.0&amp;packages=corechart" type="text/javascript"></script>
    <link href="https://www.google.com/uds/api/visualization/1.0/40ff64b1d9d6b3213524485974f36cc0/ui+en.css" type="text/css" rel="stylesheet">
    <script src="https://www.google.com/uds/api/visualization/1.0/40ff64b1d9d6b3213524485974f36cc0/format+en,default+en,ui+en,corechart+en.I.js" type="text/javascript"></script>
    <div id="chart_div" style="height:600px;width:100%;"></div>
</body>
</html>`

func graphData(list []*perf, div float64, printlatency bool) []byte {
	data := new(bytes.Buffer)

	data.WriteString("[[")
	data.WriteString(`"Argument", "Average"`)
	if printlatency {
		data.WriteString(`,"50%", "75%", "90%", "95%", "99%"`)
	}
	data.WriteString("],")

	for i, p := range list {
		if i != 0 {
			data.WriteString(",")
		}

		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(p.reqbysec, 'f', 2, 64)) // NbRequest (x)
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat((p.avg / div), 'f', 2, 64)) // Avg (y)
		if printlatency {
			data.WriteString(",")
			data.WriteString(strconv.FormatFloat((p.perc50 / div), 'f', 2, 64)) // 50% (y)
			data.WriteString(",")
			data.WriteString(strconv.FormatFloat((p.perc75 / div), 'f', 2, 64)) // 75% (y)
			data.WriteString(",")
			data.WriteString(strconv.FormatFloat((p.perc90 / div), 'f', 2, 64)) // 90% (y)
			data.WriteString(",")
			data.WriteString(strconv.FormatFloat((p.perc95 / div), 'f', 2, 64)) // 95% (y)
			data.WriteString(",")
			data.WriteString(strconv.FormatFloat((p.perc99 / div), 'f', 2, 64)) // 99% (y)
		}
		data.WriteString("]")

	}
	data.WriteString("]")

	return data.Bytes()
}

func writeLocallyData(data, title, valueType string) (string, error) {
	t := template.New("limitri Template")
	t, err := t.Parse(templateHTML)
	if err != nil {
		return "", err
	}

	tmpfile, err := ioutil.TempFile("", "limitri")
	if err != nil {
		return "", err
	}

	model := struct {
		Title string
		Data  template.JS
		Type  string
	}{
		title,
		template.JS(data),
		valueType,
	}

	err = t.Execute(tmpfile, model)
	if err != nil {
		return "", err
	}

	tmpfile.Close()
	newName := fmt.Sprintf("%s.html", tmpfile.Name())
	err = os.Rename(tmpfile.Name(), newName)
	if err != nil {
		newName = tmpfile.Name()
	}

	return newName, nil

}
