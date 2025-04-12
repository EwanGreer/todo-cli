package status

type Status uint

const (
	Done Status = iota
	InProgress
	Ready
)

func (s Status) Is(stat Status) bool {
	return s == stat
}
