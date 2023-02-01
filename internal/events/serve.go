package events

import "fmt"

type IMAPServerReady struct {
	eventBase

	Port int
}

func (event IMAPServerReady) String() string {
	return fmt.Sprintf("IMAPServerReady: Port %d", event.Port)
}

type IMAPServerStopped struct {
	eventBase
}

func (event IMAPServerStopped) String() string {
	return "IMAPServerStopped"
}

type IMAPServerError struct {
	eventBase

	Error error
}

func (event IMAPServerError) String() string {
	return fmt.Sprintf("IMAPServerError: %v", event.Error)
}

type SMTPServerReady struct {
	eventBase

	Port int
}

func (event SMTPServerReady) String() string {
	return fmt.Sprintf("SMTPServerReady: Port %d", event.Port)
}

type SMTPServerStopped struct {
	eventBase
}

func (event SMTPServerStopped) String() string {
	return "SMTPServerStopped"
}

type SMTPServerError struct {
	eventBase

	Error error
}

func (event SMTPServerError) String() string {
	return fmt.Sprintf("SMTPServerError: %v", event.Error)
}
