package storage

import "github.com/DjordjeVuckovic/news-hunter/internal/types/query"

type CapabilityProvider interface {
	GetCapabilities() query.Capabilities
}
