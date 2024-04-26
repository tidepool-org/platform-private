package json

import (
	"encoding/json"
	"io"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

// CONCURRENCY: SAFE IFF writer is safe

func NewSerializer(writer io.Writer) (log.Serializer, error) {
	if writer == nil {
		return nil, errors.New("writer is missing")
	}

	return &serializer{
		writer: writer,
	}, nil
}

type serializer struct {
	writer io.Writer
}

func (s *serializer) Serialize(fields log.Fields) error {
	if fields == nil {
		return errors.New("fields are missing")
	}

	bites, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrapf(err, "unable to serialize fields")
	}

	bites = append(bites, []byte("\n")...)

	_, err = s.writer.Write(bites)
	if err != nil {
		return errors.Wrapf(err, "unable to write serialized field")
	}

	return nil
}

func (s *serializer) SetWriter(w io.Writer) {
	s.writer = w
}
