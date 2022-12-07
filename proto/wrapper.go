package proto

type ProtobufAlgorithm interface {
	UpdateSFUStatus(current []*SFUStatus, reports []*QualityReport) (expected []*SFUStatus)
}

// AlgorithmWrapper wrap the ProtobufAlgorithm
type AlgorithmWrapper struct {
	ProtobufAlgorithm
}

func (w AlgorithmWrapper) UpdateStatus(current []Status, reports []Report) (expected []Status) {
	current_ := make([]*SFUStatus, len(current))
	for i, c := range current {
		current_[i] = c.(*SFUStatus)
	}
	reports_ := make([]*QualityReport, len(reports))
	for i, r := range reports {
		reports_[i] = r.(*QualityReport)
	}
	expected_ := w.ProtobufAlgorithm.UpdateSFUStatus(current_, reports_)
	expected = make([]Status, len(expected_))
	for i, e := range expected_ {
		expected[i] = e
	}
	return expected
}
