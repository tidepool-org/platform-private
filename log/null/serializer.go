package null

import (
	"io"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

// CONCURRENCY: SAFE

func NewSerializer() log.Serializer {
	return &serializer{}
}

type serializer struct{}

func (s *serializer) Serialize(fields log.Fields) error {
	if fields == nil {
		return errors.New("fields are missing")
	}

	return nil
}

func (s *serializer) SetWriter(w io.Writer) { /* no op */ }
