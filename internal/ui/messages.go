package ui

type stepCompleteMsg struct {
	stepName string
	nextStep int
}

type errMsg error
type doneMsg struct{}
type upToDateMsg struct {
	version string
}

type upToDateError struct {
	version string
}

func (e *upToDateError) Error() string {
	return "already up to date"
}
