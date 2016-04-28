package manager

import "fmt"

// waitableEvent provides a way to wait for event's processing to complete
// and return the event's processing status.
// This can be useful for generating responses to a UI event.
// Note that an event processing may itself generate more events and it is up to
// the processing logic of the event to handle waits internally.
type waitableEvent struct {
	inEvent  event
	statusCh chan error
}

// newWaitableEvent creates and returns waitableEvent event
func newWaitableEvent(e event) *waitableEvent {
	return &waitableEvent{
		inEvent:  e,
		statusCh: make(chan error),
	}
}

func (e *waitableEvent) String() string {
	return fmt.Sprintf("waitableEvent: %s", e.inEvent)
}

func (e *waitableEvent) process() error {
	// run the contained event's processing
	err := e.inEvent.process()
	// signal it's status
	e.statusCh <- err
	//return the status to event loop
	return err
}

func (e *waitableEvent) waitForCompletion() error {
	select {
	case err := <-e.statusCh:
		return err
	}
}
