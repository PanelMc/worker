package worker

func (s *server) Start() (err error) {
	err = s.container.Start()
	
	return
}
