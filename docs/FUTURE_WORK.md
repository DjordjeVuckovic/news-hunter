# FUTURE_WORK.md

## Search Results Analysis & Unification Roadmap

### Current Issues Identified

#### Search Result Differences Between PostgreSQL and Elasticsearch

After analyzing search results from both storage backends, several critical discrepancies have been identified:

**Data Integrity Issues:**
- **Date Fields**: PostgreSQL returns proper timestamps (`2025-08-07T01:48:44.945761+02:00`), while Elasticsearch shows zero dates (`0001-01-01T00:00:00Z`)
- **Missing Fields**: Elasticsearch results lack `language` field and some `sourceId` values
- **Field Completeness**: PostgreSQL shows more complete metadata across all fields

**Ranking Inconsistencies:**
- **Scale Differences**: PostgreSQL uses 0-1 normalized scale (e.g., `0.091906235`), Elasticsearch uses 1-30+ scale (e.g., `23.70625`)
- **Ranking Logic**: Different algorithms produce different result ordering for identical queries
- **Score Interpretation**: No unified way to compare relevance across storage backends

**Result Quality Comparison:**

**PostgreSQL Strengths:**
- Complete data integrity with all timestamps and fields populated
- Consistent metadata structure
- More diverse result sets mixing different topics
- Reliable date-based filtering and sorting
- Normalized ranking scores

**Elasticsearch Strengths:**
- Superior semantic clustering (related articles grouped together)
- Better content relevance matching
- Advanced full-text search capabilities
- More sophisticated ranking algorithms
- Better handling of complex queries

**Current Recommendation:** PostgreSQL produces more **reliable and complete** results, while Elasticsearch provides **better search relevance**. The ideal solution requires unifying both strengths.

---

## Implementation Roadmap

### Phase 1: Search Comparison Infrastructure

#### 1.1 Create CLI Comparison Tool
**Location:** `cmd/search_compare/`

**Features:**
- Query both backends simultaneously with identical parameters
- Generate side-by-side JSON comparison output
- Highlight data discrepancies and missing fields
- Support batch testing with multiple queries from config file
- Export comparison reports in multiple formats (JSON, CSV, HTML)

**Sample Usage:**
```bash
./bin/search-compare --query "Trump" --backends "pg,es" --output comparison-report.json
./bin/search-compare --config test-queries.yaml --report-format html
```

#### 1.2 Automated Regression Testing
- Integration with CI/CD pipeline
- Automated alerts for ranking drift between backends
- Historical comparison tracking

### Phase 2: Data Consistency Resolution

#### 2.1 Data Import Audit
**Investigation Areas:**
- Verify identical source data is imported to both backends
- Check field mapping configurations for consistency
- Identify root cause of zero dates in Elasticsearch
- Ensure character encoding and data type consistency

**Action Items:**
- Review `internal/storage/es/` and `internal/storage/pg/` implementations
- Compare mapping configurations in `configs/mappings/`
- Add data validation during import process
- Implement consistency checks between storage backends

#### 2.2 Schema/Index Alignment
**Elasticsearch Index Mapping Review:**
```json
{
  "mappings": {
    "properties": {
      "createdAt": {"type": "date", "format": "strict_date_time"},
      "metadata.importedAt": {"type": "date", "format": "strict_date_time"},
      "metadata.publishedAt": {"type": "date", "format": "strict_date_time"}
    }
  }
}
```

**PostgreSQL Schema Verification:**
- Ensure timestamp fields use consistent timezone handling
- Verify full-text search configuration matches ES capabilities
- Align indexing strategies for optimal performance

### Phase 3: Unified Ranking System

#### 3.1 Ranking Abstraction Layer
**New Package:** `internal/ranking/`

**Interface Definition:**
```go
type RankingService interface {
    NormalizeScore(engineScore float64, engine storage.Type) float64
    CalculateCompositeScore(textRelevance, dateRecency, sourceAuthority float64) float64
    ExplainRanking(result SearchResult) RankingExplanation
}

type RankingExplanation struct {
    TotalScore      float64 `json:"total_score"`
    TextRelevance   float64 `json:"text_relevance"`
    DateRecency     float64 `json:"date_recency"`
    SourceAuthority float64 `json:"source_authority"`
    Explanation     string  `json:"explanation"`
}
```

**Components:**
- **Score Normalization**: Convert engine-specific scores to 0-1 scale
- **Composite Ranking**: Combine multiple relevance factors
- **Configurable Weights**: Adjustable importance of different ranking factors
- **Ranking Explanation**: Transparent scoring breakdown for debugging

#### 3.2 Enhanced Search Result Structure
```go
type UnifiedSearchResult struct {
    Article           dto.Article         `json:"article"`
    Score             float64            `json:"score"`           // Normalized 0-1
    RankingDetails    RankingExplanation `json:"ranking_details"`
    StorageEngine     string             `json:"storage_engine"`  // For debugging
}
```

### Phase 4: Search Quality Framework

#### 4.1 Evaluation Metrics
**Query Categories:**
- **Exact Match**: Entity names, specific phrases
- **Semantic Search**: Conceptual queries requiring understanding
- **Temporal Queries**: Date-based filtering and recency
- **Multi-faceted**: Complex queries with multiple criteria

**Quality Metrics:**
- **Relevance**: Manual evaluation of top-10 results
- **Diversity**: Topic coverage in result sets
- **Freshness**: Appropriate weighting of recent content
- **Completeness**: Data field coverage and accuracy

#### 4.2 Benchmark Suite
**Test Query Examples:**
```yaml
test_queries:
  - query: "Donald Trump fraud trial"
    intent: "legal_proceedings"
    expected_topics: ["legal", "politics", "business"]
    time_sensitivity: "high"
    
  - query: "climate change renewable energy"
    intent: "environmental_policy"
    expected_topics: ["environment", "technology", "policy"]
    time_sensitivity: "medium"
```

**Automated Scoring:**
- NDCG (Normalized Discounted Cumulative Gain) for ranking quality
- Topic diversity measurements
- Date distribution analysis
- Source authority scoring

### Phase 5: Production Optimization

#### 5.1 Performance Monitoring
- Query performance comparison between backends
- Index optimization based on search patterns
- Cache warming strategies for consistent performance

#### 5.2 Hybrid Search Strategy
**Intelligent Backend Selection:**
- Route queries to optimal storage backend based on query characteristics
- Combine results from multiple backends for complex queries
- Fallback mechanisms for backend failures

#### 5.3 Real-time Quality Assurance
- Monitor ranking drift over time
- Alert on significant result quality degradation
- A/B testing framework for ranking algorithm improvements

---

## Success Criteria

### Short-term (Next Sprint)
- [ ] Search comparison CLI tool operational
- [ ] Data consistency issues identified and documented
- [ ] Basic ranking normalization implemented

### Medium-term (Next Quarter)
- [ ] Unified ranking system deployed
- [ ] Data import issues resolved
- [ ] Automated quality monitoring in place

### Long-term (Next Release)
- [ ] Search quality matches or exceeds single-backend performance
- [ ] Transparent backend switching without user impact
- [ ] Comprehensive quality assurance framework operational

---

## Technical Debt & Considerations

### Current Technical Debt
1. **Hardcoded Ranking Logic**: Engine-specific scoring embedded in storage implementations
2. **Inconsistent Error Handling**: Different error responses between storage backends
3. **Missing Integration Tests**: No comprehensive cross-backend testing
4. **Configuration Drift**: Storage-specific settings not centrally managed

### Architecture Improvements
1. **Ranking Service**: Extract ranking logic into dedicated service
2. **Result Transformation Pipeline**: Standardize result processing across backends
3. **Configuration Management**: Unified config for all storage backend settings
4. **Observability**: Enhanced logging and metrics for search operations

### Performance Considerations
- Impact of additional ranking normalization on response times
- Memory usage of storing ranking explanations
- Network overhead of cross-backend comparisons
- Cache invalidation strategies for unified results

---

## Risk Assessment

### High Priority Risks
- **Data Loss**: Incorrect field mapping during unification
- **Performance Degradation**: Additional processing overhead
- **Ranking Quality**: Normalization may reduce relevance quality

### Mitigation Strategies
- Comprehensive backup strategy before major changes
- Gradual rollout with feature flags
- A/B testing to validate ranking improvements
- Rollback procedures for each implementation phase

---

## Resource Requirements

### Development Effort
- **Phase 1**: 1-2 weeks (Comparison tool)
- **Phase 2**: 2-3 weeks (Data consistency)
- **Phase 3**: 3-4 weeks (Ranking unification)
- **Phase 4**: 2-3 weeks (Quality framework)
- **Phase 5**: Ongoing (Optimization)

### Infrastructure
- Additional monitoring and alerting setup
- Test data sets for comprehensive evaluation
- Performance testing environment
- Documentation and training materials

---

*Document created: 2025-08-31*  
*Last updated: 2025-08-31*  
*Status: Planning Phase*