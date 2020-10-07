package worker

func (s *server) Stop() (err error) {
	err = s.container.Stop()

	return
}
