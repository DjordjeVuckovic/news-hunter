package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/ingest/reader"
)

type preprocessorConfig struct {
	InputPath   string
	OutputDir   string
	MappingPath string
	Workers     int
	WriteReport bool
}

type PreprocessReport struct {
	TotalRecords      int       `json:"total_records"`
	ProcessedRecords  int       `json:"processed_records"`
	DuplicatesRemoved int       `json:"duplicates_removed"`
	ProcessingTime    float64   `json:"processing_time_seconds"`
	OutputFile        string    `json:"output_file"`
	Timestamp         time.Time `json:"timestamp"`
}

func parseFlags() preprocessorConfig {
	var cfg preprocessorConfig
	flag.StringVar(&cfg.InputPath, "input", "", "Path to the input CSV file")
	flag.StringVar(&cfg.OutputDir, "output", "", "Output directory for canonical dataset")
	flag.StringVar(&cfg.MappingPath, "mapping", "", "Path to the YAML field-mapping config")
	flag.IntVar(&cfg.Workers, "workers", 16, "Number of parallel workers")
	flag.BoolVar(&cfg.WriteReport, "report", false, "Write validation report")
	flag.Parse()
	return cfg
}

func main() {
	cfg := parseFlags()
	if cfg.InputPath == "" || cfg.OutputDir == "" || cfg.MappingPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	if err := runPreprocessor(ctx, cfg); err != nil {
		slog.Error("preprocessing failed", "error", err)
		os.Exit(1)
	}

	slog.Info("preprocessing completed successfully")
}

func runPreprocessor(ctx context.Context, cfg preprocessorConfig) error {
	start := time.Now()

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	inputBasename := strings.TrimSuffix(filepath.Base(cfg.InputPath), filepath.Ext(cfg.InputPath))
	outputFilename := fmt.Sprintf("%s_canonical.jsonl", inputBasename)
	outputPath := filepath.Join(cfg.OutputDir, outputFilename)

	mappingFile, err := os.Open(cfg.MappingPath)
	if err != nil {
		return fmt.Errorf("failed to open mapping config: %w", err)
	}
	defer mappingFile.Close()

	mappingCfg, err := reader.NewYAMLConfigLoader(mappingFile).Load(true)
	if err != nil {
		return fmt.Errorf("failed to load mapping config: %w", err)
	}
	mapper := reader.NewArticleMapper(mappingCfg)

	dataFile, err := os.Open(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer dataFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	report := &PreprocessReport{
		Timestamp:  time.Now(),
		OutputFile: outputFilename,
	}

	csvReader := reader.NewCSVReader(dataFile)
	resultsChan, err := csvReader.ReadParallel(ctx, cfg.Workers)
	if err != nil {
		return fmt.Errorf("failed to create parallel reader: %w", err)
	}

	encoder := json.NewEncoder(outFile)

	for result := range resultsChan {
		report.TotalRecords++

		if result.Err != nil {
			slog.Warn("failed to read record", "error", result.Err)
			continue
		}

		article, err := mapper.Map(result.Record)
		if err != nil {
			slog.Warn("failed to map record", "error", err)
			continue
		}

		if err := encoder.Encode(reader.ToCanonicalRecord(article)); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}

		report.ProcessedRecords++
	}

	report.ProcessingTime = time.Since(start).Seconds()

	if cfg.WriteReport {
		if err := writeReport(cfg.OutputDir, inputBasename, report); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
	}

	logSummary(report)

	return nil
}

func writeReport(outputDir, basename string, report *PreprocessReport) error {
	reportPath := filepath.Join(outputDir, fmt.Sprintf("%s_report.json", basename))

	reportFile, err := os.Create(reportPath)
	if err != nil {
		return err
	}
	defer reportFile.Close()

	encoder := json.NewEncoder(reportFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return err
	}

	slog.Info("report written", "path", reportPath)
	return nil
}

func logSummary(report *PreprocessReport) {
	slog.Info("preprocessing summary",
		"total_records", report.TotalRecords,
		"processed_records", report.ProcessedRecords,
		"duplicates_removed", report.DuplicatesRemoved,
		"processing_time", fmt.Sprintf("%.2fs", report.ProcessingTime),
	)
}
