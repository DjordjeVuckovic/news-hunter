package storage

import "github.com/DjordjeVuckovic/news-hunter/internal/domain/query"

type CapabilityProvider interface {
	GetCapabilities() query.Capabilities
}
