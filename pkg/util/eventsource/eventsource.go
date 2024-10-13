package eventsource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lingticio/llmg/pkg/util/nanoid"
)

// Event represents Server-Sent Event.
// SSE explanation: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
type Event struct {
	// ID is used to set the EventSource object's last event ID value.
	ID []byte
	// Data field is for the message. When the EventSource receives multiple consecutive lines
	// that begin with data:, it concatenates them, inserting a newline character between each one.
	// Trailing newlines are removed.
	Data []byte
	// Event is a string identifying the type of event described. If this is specified, an event
	// will be dispatched on the browser to the listener for the specified event name; the website
	// source code should use addEventListener() to listen for named events. The onmessage handler
	// is called if no event name is specified for a message.
	Event []byte
	// Retry is the reconnection time. If the connection to the server is lost, the browser will
	// wait for the specified time before attempting to reconnect. This must be an integer, specifying
	// the reconnection time in milliseconds. If a non-integer value is specified, the field is ignored.
	Retry []byte
	// Comment line can be used to prevent connections from timing out; a server can send a comment
	// periodically to keep the connection alive.
	Comment []byte
}

// MarshalTo marshals Event to given Writer.
func (ev *Event) MarshalTo(w io.Writer) error {
	// Marshalling part is taken from: https://github.com/r3labs/sse/blob/c6d5381ee3ca63828b321c16baa008fd6c0b4564/http.go#L16
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 { //nolint
		if len(ev.ID) > 0 {
			if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
				return err
			}
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}

type ResponseAdapterType string

const (
	ResponseAdapterTypeEcho ResponseAdapterType = "echo"
)

type options struct {
	id                string
	idGenerateHandler func() string

	responseType ResponseAdapterType
	echoResponse *echo.Response
}

type CallOption func(*options)

func WithEchoResponse(response *echo.Response) CallOption {
	return func(o *options) {
		o.responseType = ResponseAdapterTypeEcho
		o.echoResponse = response
	}
}

func WithID(id string) CallOption {
	return func(o *options) {
		o.id = id
	}
}

func WithIDGenerator(handler func() string) CallOption {
	return func(o *options) {
		o.idGenerateHandler = handler
	}
}

type EventSource[D any] struct {
	options *options
}

func NewEventSource[D any](callOpts ...CallOption) *EventSource[D] {
	opts := &options{
		idGenerateHandler: func() string {
			return nanoid.New()
		},
	}

	for _, opt := range callOpts {
		opt(opts)
	}

	if opts.id == "" {
		opts.id = opts.idGenerateHandler()
	}

	es := &EventSource[D]{
		options: opts,
	}

	switch opts.responseType {
	case ResponseAdapterTypeEcho:
		ApplyHeaders(opts.echoResponse.Header())
	}

	return es
}

func ApplyHeaders(header http.Header) {
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
}

func (e *EventSource[D]) SendJSON(message D) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	event := Event{
		Data: jsonData,
	}

	switch e.options.responseType {
	case ResponseAdapterTypeEcho:
		err = event.MarshalTo(e.options.echoResponse.Writer)
		if err != nil {
			return err
		}

		e.options.echoResponse.Flush()
	}

	return nil
}

func (e *EventSource[D]) SendRaw(message []byte) error {
	event := Event{
		Data: message,
	}

	switch e.options.responseType {
	case ResponseAdapterTypeEcho:
		err := event.MarshalTo(e.options.echoResponse.Writer)
		if err != nil {
			return err
		}

		e.options.echoResponse.Flush()
	}

	return nil
}

func (e *EventSource[D]) SendWithID(id string, message D) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	event := Event{
		ID:   []byte(id),
		Data: jsonData,
	}

	switch e.options.responseType {
	case ResponseAdapterTypeEcho:
		err = event.MarshalTo(e.options.echoResponse.Writer)
		if err != nil {
			return err
		}

		e.options.echoResponse.Flush()
	}

	return nil
}
