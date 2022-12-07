package proto

type Status interface {
	// Key return the key of this item
	Key() string
	// Compare compare if two IndexData is the same
	Compare(Status) bool
	// String stringify it
	String() string
	// Clone clone return a copy of this struct
	Clone() Status
}
type Report interface {
	// String stringify it
	String() string
	// Clone clone return a copy of this struct
	Clone() Report
}

type Request interface {
	isRequest()
}

type RequestStatus struct {
	Status Status
}

type RequestReport struct {
	Report Report
}

func (*RequestStatus) isRequest() {}

func (*RequestReport) isRequest() {}

// Algorithm is the node selection algorithm interface
type Algorithm interface {

	// UpdateStatus tell the algorithm that the SFU graph and the computation and communication quality has changed
	// `current` is the changed SFU's current status
	// `reports` is a series of Quality Report
	// `expected` is that which SFU's status should change
	UpdateStatus(current []Status, reports []Report) (expected []Status)
}
