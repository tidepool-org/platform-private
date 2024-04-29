package log

import (
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoLogSink returns a Logger adapted to implement options.LogSink.
func NewMongoLogSink(l Logger) options.LogSink {
	return &MongoLogger{Logger: l}
}

// MongoLogger wraps a Logger to implement options.LogSink.
type MongoLogger struct {
	Logger
}

func (l *MongoLogger) Info(level int, message string, keysAndValues ...interface{}) {
	l.Logger.WithFields(l.fieldsFromArgs(keysAndValues...)).Info(message)
}

func (l *MongoLogger) Error(err error, message string, keysAndValues ...interface{}) {
	l.Logger.WithError(err).WithFields(l.fieldsFromArgs(keysAndValues...)).Error(message)
}

func (l MongoLogger) fieldsFromArgs(keysAndValues ...interface{}) Fields {
	if len(keysAndValues)%2 != 0 {
		l.Warnf("bad number of keys and values: %d", len(keysAndValues))
		return Fields{}
	}
	fields := make(Fields, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i].(string)
		v := keysAndValues[i+1]
		fields[k] = v
	}
	return fields
}
