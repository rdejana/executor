package log_streamer

import (
	"io"
	"strconv"

	"github.com/cloudfoundry/dropsonde/emitter/logemitter"
	"github.com/cloudfoundry/dropsonde/events"
)

const MAX_MESSAGE_SIZE = 4096

type LogStreamer interface {
	Stdout() io.Writer
	Stderr() io.Writer

	Flush()

	WithSource(sourceName string) LogStreamer
}

type logStreamer struct {
	stdout *streamDestination
	stderr *streamDestination
}

func New(guid string, sourceName string, index *int, loggregatorEmitter logemitter.Emitter) LogStreamer {
	if guid == "" {
		return noopStreamer{}
	}

	sourceIndex := "0"
	if index != nil {
		sourceIndex = strconv.Itoa(*index)
	}

	return &logStreamer{
		stdout: newStreamDestination(
			guid,
			sourceName,
			sourceIndex,
			events.LogMessage_OUT,
			loggregatorEmitter,
		),

		stderr: newStreamDestination(
			guid,
			sourceName,
			sourceIndex,
			events.LogMessage_ERR,
			loggregatorEmitter,
		),
	}
}

func (e *logStreamer) Stdout() io.Writer {
	return e.stdout
}

func (e *logStreamer) Stderr() io.Writer {
	return e.stderr
}

func (e *logStreamer) Flush() {
	e.stdout.flush()
	e.stderr.flush()
}

func (e *logStreamer) WithSource(sourceName string) LogStreamer {
	if sourceName == "" {
		return e
	}

	return &logStreamer{
		stdout: e.stdout.withSource(sourceName),
		stderr: e.stderr.withSource(sourceName),
	}
}
