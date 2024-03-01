package system

type ReadyChecker interface {
	IsReady() bool
}

type ReadinessService struct {
	toCheck []ReadyChecker
}

func NewReadinessService(toCheck ...ReadyChecker) *ReadinessService {
	return &ReadinessService{
		toCheck: toCheck,
	}
}

func (s *ReadinessService) IsReady() bool {
	for _, c := range s.toCheck {
		if !c.IsReady() {
			return false
		}
	}
	return true
}
