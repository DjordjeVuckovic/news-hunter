package judgment

import (
	"fmt"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"gopkg.in/yaml.v3"
)

func ExportForAnnotation(poolFile *pool.PoolFile, outputPath string) error {
	jf := JudgmentFile{
		Strategy: "manual",
		Queries:  make([]JudgmentEntry, 0, len(poolFile.Queries)),
	}

	for _, pe := range poolFile.Queries {
		entry := JudgmentEntry{
			QueryID: pe.QueryID,
			Docs:    make([]GradedDoc, 0, len(pe.Docs)),
		}
		for _, doc := range pe.Docs {
			entry.Docs = append(entry.Docs, GradedDoc{
				DocID: doc.DocID,
				Grade: -1,
			})
		}
		jf.Queries = append(jf.Queries, entry)
	}

	data, err := yaml.Marshal(&jf)
	if err != nil {
		return fmt.Errorf("marshal judgment template: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write judgment template: %w", err)
	}
	return nil
}

func ImportAnnotations(path string) (*JudgmentFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read judgment file: %w", err)
	}
	var jf JudgmentFile
	if err := yaml.Unmarshal(data, &jf); err != nil {
		return nil, fmt.Errorf("parse judgment file: %w", err)
	}
	return &jf, nil
}
