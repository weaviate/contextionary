package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

func serveResults(index int, corpus, oldPrediction, control string, ranked []wordWithDistance) {
	data := struct {
		Input     string
		OldResult string
		Control   string
		List      []wordWithDistance
	}{
		Input:     corpus,
		List:      ranked,
		OldResult: oldPrediction,
		Control:   control,
	}

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc(fmt.Sprintf("/%d", index), func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, data)
	})

}

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Noise Spike</title>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.8.1/css/bulma.css" />
	</head>
	<body>
	  <section class="section">
			<div class="container">
			  <h2 class="subtitle">Input</h2>
				  <table class="table">
						<thead>
							<tr>
								<th>Old Mechanism</th>
								<th>Control</th>
							</tr>
						</thead>
						<tbody>
							<tr>
							  <td>{{ .OldResult }}</td>
							  <td>{{ .Control }}</td>
							</tr>
						</tbody>
					</table>
					<pre style="max-height:250px; overflow:scroll">
					 {{ .Input }}
				  </pre>
					<div class="columns" style="margin-top: 50px">
						<div class="column">
			  <h2 class="subtitle">List</h2>
					<table class="table">
						<thead>
							<tr>
								<th>Dist</th>
								<th>Word</th>
								<th>Predicted Category</th>
								<th>Information Gain</th>
							</tr>
						</thead>
						<tbody>
							{{range .List}}
								<tr>
									<td>{{ .Distance }}</td>
									<td>{{ .Word }}</td>
									<td>{{ .Prediction }}</td>
									<td>{{ .InformationGain }}</td>
								</tr>
							{{end}}
						</tbody>
					</table>
					</div>
				</div>
			</div>
		</section>
	</body>
</html>`
