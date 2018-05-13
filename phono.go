package phono

import (
	"context"
	"sync"
)

// Pipe transport types
type (
	// Samples represent a sample data sliced per channel
	Samples [][]float64

	// Message is a main structure for pipe transport
	Message struct {
		// Samples of message
		Samples
		// Pulse
		*Params
		*sync.WaitGroup
	}

	// NewMessageFunc is a message-producer function
	NewMessageFunc func() Message
)

// Pipe function types
type (
	// PumpFunc is a function to pump sound data to pipe
	PumpFunc func(context.Context, NewMessageFunc) (out <-chan Message, errc <-chan error, err error)

	// ProcessFunc is a function to process sound data in pipe
	ProcessFunc func(ctx context.Context, in <-chan Message) (out <-chan Message, errc <-chan error, err error)

	// SinkFunc is a function to sink data from pipe
	SinkFunc func(ctx context.Context, in <-chan Message) (errc <-chan error, err error)
)

// Params support types
type (
	// ParamFunc represents a function which applies the param
	ParamFunc func()

	// Param is a structure for delayed parameters apply
	Param struct {
		Consumer interface{}
		Apply    ParamFunc
	}

	// Params represents current track attributes: time signature, bpm e.t.c.
	Params struct {
		private map[interface{}][]ParamFunc
	}
)

// Small types for common params
type (
	// BufferSize represents a buffer size value
	BufferSize int
	// NumChannels represents a number of channels
	NumChannels int
	// SampleRate represents a sample rate value
	SampleRate int
	// Tempo represents a tempo value
	Tempo float32
	// TimeSignature represents a time signature
	TimeSignature struct {
		NotesPerBar int // 3 in 3/4
		NoteValue   int // 4 in 3/4
	}
	// SamplePosition represents a position in samples measure
	SamplePosition int64
)

// NewParams returns a new params instance with initialised map inside
func NewParams(params ...Param) (result *Params) {
	result = &Params{
		private: make(map[interface{}][]ParamFunc),
	}
	result.Add(params...)
	return
}

// Add accepts a slice of params
func (p *Params) Add(params ...Param) *Params {
	if p == nil {
		return nil
	}
	for _, param := range params {
		private, ok := p.private[param.Consumer]
		if !ok {
			private = make([]ParamFunc, 0, len(params))
		}
		private = append(private, param.Apply)

		p.private[param.Consumer] = private
	}

	return p
}

// ApplyTo consumes params defined for consumer in this param set
func (p *Params) ApplyTo(consumer interface{}) {
	if p == nil {
		return
	}
	if params, ok := p.private[consumer]; ok {
		for _, param := range params {
			param()
		}
	}
}

// Join two param sets into one
func (p *Params) Join(source *Params) *Params {
	if p == nil || p.Empty() {
		return source
	}
	for newKey, newValues := range source.private {
		if _, ok := p.private[newKey]; ok {
			p.private[newKey] = append(p.private[newKey], newValues...)
		} else {
			p.private[newKey] = newValues
		}
	}
	return p
}

// Empty returns true if params are empty
func (p Params) Empty() bool {
	if p.private == nil || len(p.private) == 0 {
		return true
	}
	return false
}

// RecievedBy should be called by sink once message is recieved
func (m *Message) RecievedBy(reciever interface{}) {
	// TODO check in map of receivers
	if m.WaitGroup != nil {
		m.Done()
	}
}
