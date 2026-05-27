package report

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteHTML renders r as a self-contained HTML file at path, creating parent
// directories as needed. The file contains inline CSS, inline SVG charts, and
// ~50 lines of vanilla JS for sortable table columns — no external dependencies.
func WriteHTML(r *Report, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	buf, err := RenderHTML(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}

// RenderHTML returns the fully rendered HTML bytes for a report.
func RenderHTML(r *Report) ([]byte, error) {
	vm := buildViewModel(r)
	tmpl, err := template.New("report").Funcs(htmlFuncs()).Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse html template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vm); err != nil {
		return nil, fmt.Errorf("render html: %w", err)
	}
	return buf.Bytes(), nil
}

// ─── view model ─────────────────────────────────────────────────────────────

type htmlReport struct {
	Title     string
	RunID     string
	Tool      string
	Generated string
	SpecID    string
	Sources   *htmlSources
	Jobs      []htmlJob
}

type htmlSources struct {
	Spec, Suite, Pool, Judgments string
}

type htmlJob struct {
	Name         string
	Aggregated   []htmlAggRow
	Latency      []htmlLatRow
	Significance []htmlSigRow
	PerQuery     []htmlQueryRow
	NDCGChart    template.HTML
	KValues      []int
}

type htmlAggRow struct {
	Engine string
	NDCG   []htmlCell
	MAP    htmlCell
	MRR    htmlCell
	Bpref  htmlCell
	Judged string
	Errors string
}

type htmlCell struct {
	Val    string
	Stddev string // optional ±stddev suffix
}

type htmlLatRow struct {
	Engine                                          string
	Min, P50, P75, P90, P95, P99, Max, Mean, Stddev string
	Samples                                         int
}

type htmlSigRow struct {
	EngineA, EngineB, Metric string
	W                        string
	P                        string
	Stars                    string
	NS                       bool
}

type htmlQueryRow struct {
	Query, Engine, NDCG, Precision, AP, RR, Bpref string
	Hits                                          int64
	P50, P95                                      string
	Status                                        string
	IsErr                                         bool
}

func buildViewModel(r *Report) htmlReport {
	title := r.Provenance.SpecID
	if title == "" {
		title = "Benchmark Report"
	}

	vm := htmlReport{
		Title:     title,
		RunID:     r.Provenance.RunID,
		Tool:      r.Provenance.Tool,
		Generated: r.Provenance.GeneratedAt.Format("2006-01-02 15:04:05 UTC"),
		SpecID:    r.Provenance.SpecID,
	}
	if s := r.Provenance.Sources; s != nil {
		vm.Sources = &htmlSources{
			Spec:      s.Spec,
			Suite:     s.Suite,
			Pool:      s.Pool,
			Judgments: s.Judgments,
		}
	}

	kVals := r.Config.KValues

	for _, jr := range r.Jobs {
		job := htmlJob{Name: jr.JobName, KValues: kVals}

		for _, agg := range jr.Aggregated {
			row := htmlAggRow{Engine: agg.EngineName}
			for _, k := range kVals {
				c := htmlCell{Val: fmt.Sprintf("%.4f", agg.NDCG[k])}
				if sd, ok := agg.NDCGStddev[k]; ok && sd > 0 {
					c.Stddev = fmt.Sprintf("%.4f", sd)
				}
				row.NDCG = append(row.NDCG, c)
			}
			row.MAP = htmlCell{Val: fmt.Sprintf("%.4f", agg.MAP)}
			row.MRR = htmlCell{Val: fmt.Sprintf("%.4f", agg.MRR)}
			row.Bpref = htmlCell{Val: fmt.Sprintf("%.4f", agg.MBpref)}
			row.Judged = fmt.Sprintf("%d/%d", agg.JudgedCount, agg.QueryCount)
			row.Errors = fmt.Sprintf("%d/%d", agg.ErrorCount, agg.QueryCount)
			job.Aggregated = append(job.Aggregated, row)
		}

		for _, agg := range jr.Aggregated {
			s := agg.Latency
			job.Latency = append(job.Latency, htmlLatRow{
				Engine:  agg.EngineName,
				Min:     fmtDuration(s.Min),
				P50:     fmtDuration(s.P50()),
				P75:     fmtDuration(s.P75()),
				P90:     fmtDuration(s.P90()),
				P95:     fmtDuration(s.P95()),
				P99:     fmtDuration(s.P99()),
				Max:     fmtDuration(s.Max),
				Mean:    fmtDuration(s.Mean),
				Stddev:  fmtDuration(s.Stddev),
				Samples: s.SampleCount,
			})
		}

		for _, sig := range jr.Significance {
			row := htmlSigRow{
				EngineA: sig.EngineA,
				EngineB: sig.EngineB,
				Metric:  sig.Metric,
				W:       fmt.Sprintf("%.1f", sig.W),
				P:       fmt.Sprintf("%.4f", sig.P),
				Stars:   sig.Stars,
				NS:      sig.Stars == "",
			}
			job.Significance = append(job.Significance, row)
		}

		k := primaryK(kVals)
		for _, e := range jr.PerQuery {
			status := "OK"
			isErr := false
			if e.Error != "" {
				status = "ERR"
				isErr = true
			}
			apStr, rrStr, bpStr := "—", "—", "—"
			if e.Judged {
				apStr = fmt.Sprintf("%.4f", e.AP)
				rrStr = fmt.Sprintf("%.4f", e.RR)
				bpStr = fmt.Sprintf("%.4f", e.Bpref)
			}
			job.PerQuery = append(job.PerQuery, htmlQueryRow{
				Query:     e.QueryID,
				Engine:    e.EngineName,
				NDCG:      fmtScore(e.NDCG, k),
				Precision: fmtScore(e.Precision, k),
				AP:        apStr,
				RR:        rrStr,
				Bpref:     bpStr,
				Hits:      e.TotalMatches,
				P50:       fmtDuration(e.Latency.P50()),
				P95:       fmtDuration(e.Latency.P95()),
				Status:    status,
				IsErr:     isErr,
			})
		}

		job.NDCGChart = buildNDCGChart(jr.Aggregated, kVals)
		vm.Jobs = append(vm.Jobs, job)
	}

	return vm
}

// ─── SVG chart ───────────────────────────────────────────────────────────────

var engineColors = []string{
	"#4C8EDA", "#E05C5C", "#52C878", "#F5A623", "#9B59B6", "#1ABC9C",
}

func buildNDCGChart(aggregated []AggregatedEntry, kVals []int) template.HTML {
	if len(aggregated) == 0 || len(kVals) == 0 {
		return ""
	}

	const (
		svgW      = 620
		svgH      = 200
		padLeft   = 50
		padRight  = 20
		padTop    = 20
		padBottom = 40
		maxVal    = 1.0
	)

	plotW := svgW - padLeft - padRight
	plotH := svgH - padTop - padBottom

	nGroups := len(kVals)
	nEngines := len(aggregated)
	groupW := float64(plotW) / float64(nGroups)
	barW := groupW / float64(nEngines+1) // +1 for gap

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" class="ndcg-chart">`, svgW, svgH))

	// Y-axis gridlines + labels.
	for _, tick := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
		y := padTop + plotH - int(tick*float64(plotH)/maxVal)
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#e5e7eb" stroke-width="1"/>`,
			padLeft, y, svgW-padRight, y))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="end" font-size="11" fill="#6b7280">%.2f</text>`,
			padLeft-4, y+4, tick))
	}

	// Bars.
	for gi, k := range kVals {
		groupX := padLeft + int(float64(gi)*groupW)
		for ei, agg := range aggregated {
			val := math.Min(agg.NDCG[k], maxVal)
			barH := int(val * float64(plotH) / maxVal)
			x := groupX + int(float64(ei)*barW) + int(barW*0.2)
			y := padTop + plotH - barH
			color := engineColors[ei%len(engineColors)]
			sb.WriteString(fmt.Sprintf(
				`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="2"/>`,
				x, y, int(barW*0.8), barH, color))
			// Value label above bar.
			sb.WriteString(fmt.Sprintf(
				`<text x="%d" y="%d" text-anchor="middle" font-size="9" fill="%s">%.2f</text>`,
				x+int(barW*0.4), y-3, color, val))
		}
		// Group label (K value).
		labelX := groupX + int(groupW/2)
		labelY := padTop + plotH + 16
		sb.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d" text-anchor="middle" font-size="12" fill="#374151">@%d</text>`,
			labelX, labelY, k))
	}

	// Legend.
	legendY := padTop + plotH + 32
	legendX := padLeft
	for ei, agg := range aggregated {
		color := engineColors[ei%len(engineColors)]
		sb.WriteString(fmt.Sprintf(
			`<rect x="%d" y="%d" width="10" height="10" fill="%s" rx="1"/>`, legendX, legendY-9, color))
		sb.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d" font-size="11" fill="#374151">%s</text>`,
			legendX+14, legendY, template.HTMLEscapeString(agg.EngineName)))
		legendX += 14 + len(agg.EngineName)*7 + 16
	}

	// X-axis label.
	sb.WriteString(fmt.Sprintf(
		`<text x="%d" y="%d" text-anchor="middle" font-size="12" fill="#6b7280">NDCG@K</text>`,
		padLeft+plotW/2, svgH-4))

	sb.WriteString(`</svg>`)
	return template.HTML(sb.String())
}

// ─── HTML template ───────────────────────────────────────────────────────────

func htmlFuncs() template.FuncMap {
	return template.FuncMap{
		"fmtTime": func(t time.Time) string { return t.Format("2006-01-02 15:04:05 UTC") },
		"hasStars": func(rows []htmlSigRow) bool {
			for _, r := range rows {
				if r.Stars != "" {
					return true
				}
			}
			return false
		},
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Bench Report: {{.Title}}</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;font-size:14px;color:#111827;background:#f9fafb;padding:24px}
.container{max-width:1200px;margin:0 auto}
header{background:#fff;border:1px solid #e5e7eb;border-radius:8px;padding:20px 24px;margin-bottom:24px}
header h1{font-size:20px;font-weight:600;color:#111827;margin-bottom:12px}
.provenance{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:6px 24px}
.provenance .row{display:flex;gap:8px}
.provenance .label{color:#6b7280;min-width:80px}
.provenance .val{color:#111827;font-weight:500;word-break:break-all}
.sources{margin-top:12px;padding-top:12px;border-top:1px solid #e5e7eb}
.sources h3{font-size:12px;text-transform:uppercase;letter-spacing:.05em;color:#6b7280;margin-bottom:8px}
.sources .row .label{min-width:70px;color:#9ca3af}
.job{background:#fff;border:1px solid #e5e7eb;border-radius:8px;margin-bottom:24px;overflow:hidden}
.job-header{padding:16px 20px;background:#f8fafc;border-bottom:1px solid #e5e7eb}
.job-header h2{font-size:16px;font-weight:600;color:#111827}
.section{padding:20px}
.section+.section{border-top:1px solid #f3f4f6}
.section h3{font-size:13px;font-weight:600;text-transform:uppercase;letter-spacing:.05em;color:#6b7280;margin-bottom:12px}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:6px 10px;background:#f3f4f6;border-bottom:2px solid #e5e7eb;white-space:nowrap;cursor:pointer;user-select:none}
th:hover{background:#e9ecef}
th.sorted-asc::after{content:" ↑"}
th.sorted-desc::after{content:" ↓"}
td{padding:6px 10px;border-bottom:1px solid #f3f4f6;white-space:nowrap}
tr:last-child td{border-bottom:none}
tr:hover td{background:#fafafa}
.cell-val{font-weight:500}
.cell-sd{color:#9ca3af;font-size:11px;margin-left:4px}
.sig-stars{font-weight:700;color:#059669}
.sig-ns{color:#9ca3af}
.status-ok{color:#059669}
.status-err{color:#dc2626;font-weight:600}
.chart-section{padding:20px}
.chart-section h3{font-size:13px;font-weight:600;text-transform:uppercase;letter-spacing:.05em;color:#6b7280;margin-bottom:12px}
.ndcg-chart{display:block;max-width:100%;overflow:visible}
footer{text-align:center;color:#9ca3af;font-size:12px;margin-top:24px}
</style>
</head>
<body>
<div class="container">

<header>
  <h1>{{.Title}}</h1>
  <div class="provenance">
    <div class="row"><span class="label">Run ID</span><span class="val">{{.RunID}}</span></div>
    <div class="row"><span class="label">Tool</span><span class="val">{{.Tool}}</span></div>
    <div class="row"><span class="label">Generated</span><span class="val">{{.Generated}}</span></div>
    {{if .SpecID}}<div class="row"><span class="label">Spec</span><span class="val">{{.SpecID}}</span></div>{{end}}
  </div>
  {{if .Sources}}
  <div class="sources">
    <h3>Sources</h3>
    {{if .Sources.Spec}}<div class="row"><span class="label">spec</span><span class="val">{{.Sources.Spec}}</span></div>{{end}}
    {{if .Sources.Suite}}<div class="row"><span class="label">suite</span><span class="val">{{.Sources.Suite}}</span></div>{{end}}
    {{if .Sources.Pool}}<div class="row"><span class="label">pool</span><span class="val">{{.Sources.Pool}}</span></div>{{end}}
    {{if .Sources.Judgments}}<div class="row"><span class="label">judgments</span><span class="val">{{.Sources.Judgments}}</span></div>{{end}}
  </div>
  {{end}}
</header>

{{range .Jobs}}
<div class="job">
  <div class="job-header"><h2>{{.Name}}</h2></div>

  {{if .NDCGChart}}
  <div class="chart-section">
    <h3>NDCG@K</h3>
    {{.NDCGChart}}
  </div>
  {{end}}

  <div class="section">
    <h3>Aggregated Results</h3>
    <table class="sortable" id="agg-{{.Name}}">
      <thead><tr>
        <th>Engine</th>
        {{range .KValues}}<th>NDCG@{{.}}</th>{{end}}
        <th>MAP</th><th>MRR</th><th>Bpref</th><th>Judged</th><th>Errors</th>
      </tr></thead>
      <tbody>
      {{range .Aggregated}}
      <tr>
        <td>{{.Engine}}</td>
        {{range .NDCG}}
        <td><span class="cell-val">{{.Val}}</span>{{if .Stddev}}<span class="cell-sd">±{{.Stddev}}</span>{{end}}</td>
        {{end}}
        <td class="cell-val">{{.MAP.Val}}</td>
        <td class="cell-val">{{.MRR.Val}}</td>
        <td class="cell-val">{{.Bpref.Val}}</td>
        <td>{{.Judged}}</td>
        <td>{{.Errors}}</td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </div>

  <div class="section">
    <h3>Latency Statistics</h3>
    <table class="sortable">
      <thead><tr>
        <th>Engine</th><th>Min</th><th>p50</th><th>p75</th><th>p90</th><th>p95</th><th>p99</th><th>Max</th><th>Mean</th><th>Stddev</th><th>Samples</th>
      </tr></thead>
      <tbody>
      {{range .Latency}}
      <tr>
        <td>{{.Engine}}</td>
        <td>{{.Min}}</td><td>{{.P50}}</td><td>{{.P75}}</td><td>{{.P90}}</td>
        <td>{{.P95}}</td><td>{{.P99}}</td><td>{{.Max}}</td>
        <td>{{.Mean}}</td><td>{{.Stddev}}</td><td>{{.Samples}}</td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </div>

  {{if .Significance}}
  <div class="section">
    <h3>Statistical Significance <small style="font-weight:400;text-transform:none;letter-spacing:0">(Wilcoxon signed-rank, two-tailed — * p&lt;0.05 &nbsp; ** p&lt;0.01)</small></h3>
    <table>
      <thead><tr><th>Engine A</th><th>Engine B</th><th>Metric</th><th>W</th><th>p-value</th><th>Sig</th></tr></thead>
      <tbody>
      {{range .Significance}}
      <tr>
        <td>{{.EngineA}}</td><td>{{.EngineB}}</td><td>{{.Metric}}</td>
        <td>{{.W}}</td><td>{{.P}}</td>
        <td>{{if .NS}}<span class="sig-ns">ns</span>{{else}}<span class="sig-stars">{{.Stars}}</span>{{end}}</td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </div>
  {{end}}

  <div class="section">
    <h3>Per-Query Results</h3>
    <table class="sortable">
      <thead><tr>
        <th>Query</th><th>Engine</th><th>NDCG@K</th><th>P@K</th><th>AP</th><th>RR</th><th>Bpref</th><th>Hits</th><th>p50</th><th>p95</th><th>Status</th>
      </tr></thead>
      <tbody>
      {{range .PerQuery}}
      <tr>
        <td>{{.Query}}</td>
        <td>{{.Engine}}</td>
        <td>{{.NDCG}}</td>
        <td>{{.Precision}}</td>
        <td>{{.AP}}</td>
        <td>{{.RR}}</td>
        <td>{{.Bpref}}</td>
        <td>{{.Hits}}</td>
        <td>{{.P50}}</td>
        <td>{{.P95}}</td>
        <td>{{if .IsErr}}<span class="status-err">ERR</span>{{else}}<span class="status-ok">OK</span>{{end}}</td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </div>
</div>
{{end}}

<footer>Generated by {{.Tool}}</footer>
</div>

<script>
(function(){
  document.querySelectorAll('table.sortable').forEach(function(tbl){
    var ths = tbl.querySelectorAll('thead th');
    ths.forEach(function(th, col){
      th.addEventListener('click', function(){
        var asc = !th.classList.contains('sorted-asc');
        ths.forEach(function(h){ h.classList.remove('sorted-asc','sorted-desc'); });
        th.classList.add(asc ? 'sorted-asc' : 'sorted-desc');
        var tbody = tbl.querySelector('tbody');
        var rows = Array.from(tbody.querySelectorAll('tr'));
        rows.sort(function(a,b){
          var av = a.cells[col] ? a.cells[col].innerText.trim() : '';
          var bv = b.cells[col] ? b.cells[col].innerText.trim() : '';
          var an = parseFloat(av), bn = parseFloat(bv);
          var cmp = (!isNaN(an) && !isNaN(bn)) ? an-bn : av.localeCompare(bv);
          return asc ? cmp : -cmp;
        });
        rows.forEach(function(r){ tbody.appendChild(r); });
      });
    });
  });
})();
</script>
</body>
</html>`
