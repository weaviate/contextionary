package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

func serveResults(index, total int, corpus, oldPrediction, control string,
	ranked []wordWithDistance, percentiles []percentile, tfIdf []TermWithTfIdf) {
	var (
		pageNext string
		pagePrev string
	)

	if index > 0 {
		pagePrev = fmt.Sprintf("/%d", index-1)
	}
	if index < total-1 {
		pageNext = fmt.Sprintf("/%d", index+1)
	}

	data := struct {
		Input       string
		OldMatch    bool
		OldResult   string
		Control     string
		List        []wordWithDistance
		Percentiles []percentile
		PageNext    string
		PagePrev    string
		TfIdfList   []TermWithTfIdf
	}{
		Input:       corpus,
		List:        ranked,
		OldResult:   oldPrediction,
		Control:     control,
		Percentiles: percentiles,
		PageNext:    pageNext,
		PagePrev:    pagePrev,
		OldMatch:    oldPrediction == control,
		TfIdfList:   tfIdf,
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
			  <div class="level">
				  <div class="level-left">
					  {{ if .PagePrev }}
						 <a class="button" href="{{ .PagePrev }}">
							<span>Previous</span>
						</a>
						{{ end }}
					</div>
				  <div class="level-right">
					  {{ if .PageNext }}
						 <a class="button" href="{{ .PageNext }}">
							<span>Next</span>
						</a>
						{{ end }}
					</div>
				</div>
			  <div class="columns">
				  <div class="column is-7">
						<h2 class="subtitle">Input</h2>
							<table class="table">
								<thead>
								  <tr>
										<th>Old Mechanism</th>
										<th>Control</th>
									</tr>
								</thead>
								<tbody>
										{{ if .OldMatch }}
											<tr class="has-background-success has-text-white">
										{{ else }}
											<tr class="has-background-danger has-text-white">
										{{ end }}
										<td>{{ .OldResult }}</td>
										<td>{{ .Control }}</td>
									</tr>
								</tbody>
							</table>
							<pre style="max-height:250px; overflow:scroll">
							 {{ .Input }}
							</pre>
						</div>
						<div class="column is-5">
							<h2 class="subtitle">New Predictions</h2>
								<table class="table is-full-width">
									<thead>
										<tr>
											<th>Percentile</th>
											<th>Prediction</th>
										</tr>
									</thead>
									<tbody>
									  {{ range .Percentiles }}
										{{ if .Match }}
											<tr class="has-background-success has-text-white">
										{{ else }}
											<tr class="has-background-danger has-text-white">
										{{ end }}
											<td>{{ .Percentile }}</td>
											<td>{{ .Prediction }}</td>
										</tr>
										{{ end }}
									</tbody>
								</table>
						  


						</div>
					</div>
					<div class="columns" style="margin-top: 50px">
						<div class="column is-6">
							<h2 class="subtitle">Ranked by Information Gain</h2>
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
							<div class="column is-6">
							<h2 class="subtitle">Ranked by tf-idf</h2>
								<table class="table">
									<thead>
										<tr>
											<th>Score</th>
											<th>Word</th>
											<th>Relative Score</th>
										</tr>
									</thead>
									<tbody>
										{{range .TfIdfList}}
											<tr>
												<td>{{ .TfIdf }}</td>
												<td>{{ .Term }}</td>
												<td>{{ .RelativeScore }}</td>
											</tr>
										{{end}}
									</tbody>
								</table>
							</div>

							</div>
					</div>
				</div>
			</div>
		</section>
	</body>
</html>`

func serveSuccess(docs []doc, total, newSuccessRate, previousSuccessRate, absoluteImprovement,
	relativeImprovement float32) {

	data := struct {
		Docs                []doc
		Total               float32
		NewSuccessRate      string
		PreviousSuccessRate string
		AbsoluteImprovement string
		RelativeImprovement string
	}{
		Docs:                docs,
		Total:               total,
		NewSuccessRate:      formatAsPercentage(newSuccessRate),
		PreviousSuccessRate: formatAsPercentage(previousSuccessRate),
		AbsoluteImprovement: formatAsPercentage(absoluteImprovement),
		RelativeImprovement: formatAsPercentage(relativeImprovement),
	}

	t, err := template.New("webpage").Parse(succssTpl)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, data)
	})

}

const succssTpl = `
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
			  <h2 class="title">Overall Results Aggregation</h2>
				<table class="table">
				  <tbody>
					  <tr>
						  <th>Total</th>
							<td>{{ .Total }}</td>
						</tr>
					  <tr>
						  <th>New Success Rate</th>
							<td>{{ .NewSuccessRate }}</td>
						</tr>
					  <tr>
						  <th>Old Mechanism Success Rate</th>
							<td>{{ .PreviousSuccessRate }}</td>
						</tr>
					  <tr>
						  <th>Absolute Improvment</th>
							<td>{{ .AbsoluteImprovement }}</td>
						</tr>
					  <tr>
						  <th>Relative Improvment</th>
							<td>{{ .RelativeImprovement }}</td>
						</tr>
				  </tbody>
				</table>
			</div>
		</section>
	</body>
</html>`

func formatAsPercentage(in float32) string {
	return fmt.Sprintf("%.2f %%", in*100)
}
