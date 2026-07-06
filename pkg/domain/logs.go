package domain

// LogOptions are the framework-free knobs the logs view passes through to the
// engine; Follow is always true for the live tail, so it is not exposed here.
type LogOptions struct {
	Timestamps bool
	Since      string
	Tail       string
}
