package common_domain

type ExecutionPlan struct {
	PlanHandle []byte
	Server     ServerMeta
	XmlData    string
}
