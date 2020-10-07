package worker

func (s *server) SendCommand(cmd string) (err error) {
	err = s.container.Exec(cmd)

	return
}
