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
				interpolateNulls: true
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

func graphData(list []*perf, printlatency bool) []byte {
	data := new(bytes.Buffer)

	data.WriteString("[[")
	data.WriteString(`"Argument", "Average"`)
	data.WriteString("],")

	for i, p := range list {
		if i != 0 {
			data.WriteString(",")
		}

		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(p.reqbysec, 'f', 2, 64)) // X
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat(p.avg, 'f', 2, 64)) // Y
		data.WriteString("]")

	}
	data.WriteString("]")

	return data.Bytes()
}

// graphData translate bench results to Google graph JSON structure
func graphData2(list []*perf, printlatency bool) []byte {
	data := new(bytes.Buffer)

	data.WriteString("[[")
	data.WriteString(`"Argument", "Average"`)
	if printlatency {
		data.WriteString(`,"50%", "75%", "90%"`)
		if list[0].perc95 > 0 {
			data.WriteString(`,"95%"`)
		}
		data.WriteString(`,"99%"`)
	}
	data.WriteString("],")

	for i, p := range list {
		if i != 0 {
			data.WriteString(",")
		}
		writeAvgCoordinates(data, p.avg, p.reqbysec, printlatency, p.perc95 != 0)
		if printlatency {
			data.WriteString(",")
			write5OpercentCoordinates(data, p.perc50, p.reqbysec, printlatency, p.perc95 != 0)
			data.WriteString(",")
			write75percentCoordinates(data, p.perc75, p.reqbysec, printlatency, p.perc95 != 0)
			data.WriteString(",")
			write90percentCoordinates(data, p.perc90, p.reqbysec, printlatency, p.perc95 != 0)
			data.WriteString(",")
			write95percentCoordinates(data, p.perc95, p.reqbysec, printlatency, p.perc95 != 0)
			data.WriteString(",")
			write99percentCoordinates(data, p.perc99, p.reqbysec, printlatency, p.perc95 != 0)
		}

	}

	data.WriteString("]")

	return data.Bytes()
}

// writeAvgCoordinates should write [avg, nbRequest ... ]
func writeAvgCoordinates(data *bytes.Buffer, avg, nbRequest float64, printlatency, is95 bool) {
	data.WriteString("[")
	data.WriteString(strconv.FormatFloat(avg, 'f', 2, 64)) // X
	data.WriteString(",")
	data.WriteString(strconv.FormatFloat(nbRequest, 'f', 2, 64)) // Y

	if printlatency {
		data.WriteString(",null") // 50%
		data.WriteString(",null") // 75%
		data.WriteString(",null") // 90%
		if is95 {
			data.WriteString(",null") // 95%
		}
		data.WriteString(",null") // 99%
	}

	data.WriteString("]")
}

// write5OpercentCoordinates should write [avg, null, nbRequest, null ... ]
func write5OpercentCoordinates(data *bytes.Buffer, avg, nbRequest float64, printlatency, is95 bool) {
	if printlatency {
		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(avg, 'f', 2, 64)) // X
		data.WriteString(",null")                              // AVG
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat(nbRequest, 'f', 2, 64)) // 50%
		data.WriteString(",null")                                    // 75%
		data.WriteString(",null")                                    // 90%
		if is95 {
			data.WriteString(",null") // 95%
		}
		data.WriteString(",null") // 99%
		data.WriteString("]")
	}
}

// write75percentCoordinates should write [avg, null, null, nbRequest, null ... ]
func write75percentCoordinates(data *bytes.Buffer, avg, nbRequest float64, printlatency, is95 bool) {
	if printlatency {
		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(avg, 'f', 2, 64)) // X
		data.WriteString(",null")                              // AVG
		data.WriteString(",null")                              // 50%
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat(nbRequest, 'f', 2, 64)) // 75%
		data.WriteString(",null")                                    // 90%
		if is95 {
			data.WriteString(",null") // 95%
		}
		data.WriteString(",null") // 99%
		data.WriteString("]")
	}
}

// write90percentCoordinates should write [avg, null, null, null, nbRequest,  ... ]
func write90percentCoordinates(data *bytes.Buffer, avg, nbRequest float64, printlatency, is95 bool) {
	if printlatency {
		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(avg, 'f', 2, 64)) // X
		data.WriteString(",null")                              // AVG
		data.WriteString(",null")                              // 50%
		data.WriteString(",null")                              // 75%
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat(nbRequest, 'f', 2, 64)) // 90%
		if is95 {
			data.WriteString(",null") // 95%
		}
		data.WriteString(",null") // 99%
		data.WriteString("]")
	}
}

// write95percentCoordinates should write [avg, null, null, null, null, nbRequest,  ... ]
func write95percentCoordinates(data *bytes.Buffer, avg, nbRequest float64, printlatency, is95 bool) {
	if printlatency && is95 {
		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(avg, 'f', 2, 64)) // X
		data.WriteString(",null")                              // AVG
		data.WriteString(",null")                              // 50%
		data.WriteString(",null")                              // 75%
		data.WriteString(",null")                              // 90%
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat(nbRequest, 'f', 2, 64)) // 95%
		data.WriteString(",null")                                    // 99%
		data.WriteString("]")
	}
}

// write99percentCoordinates should write [avg, null, ..., nbRequest]
func write99percentCoordinates(data *bytes.Buffer, avg, nbRequest float64, printlatency, is95 bool) {
	if printlatency {
		data.WriteString("[")
		data.WriteString(strconv.FormatFloat(avg, 'f', 2, 64)) // X
		data.WriteString(",null")                              // AVG
		data.WriteString(",null")                              // 50%
		data.WriteString(",null")                              // 75%
		data.WriteString(",null")                              // 90%
		if is95 {
			data.WriteString(",null") // 95%
		}
		data.WriteString(",")
		data.WriteString(strconv.FormatFloat(nbRequest, 'f', 2, 64)) // 99%
		data.WriteString("]")
	}
}

func writeLocallyData(data, title string) (string, error) {
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
	}{
		title,
		template.JS(data),
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
