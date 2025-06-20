package domain

type ServerSummary struct {
	Name             string
	Type             string
	Connections      int
	RequestRate      float64
	ConnsByWaitGroup map[string]int32
}
