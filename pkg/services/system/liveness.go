package system

type LivenessService struct {
}

func NewLivenessService() *LivenessService {
	return &LivenessService{}
}

func (s *LivenessService) IsLive() bool {
	return true
}
