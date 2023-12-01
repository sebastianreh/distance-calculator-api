package httpserver

// Routes build the routes of the server
func (s *Server) Routes() {
	root := s.Server.Group(s.dependencies.Config.Prefix)
	s.Server.GET("/ping", s.dependencies.PingHandler.Ping)

	calculatorGroup := root.Group("/calculate")

	calculatorGroup.POST("/preprocess", s.dependencies.CalculatorHandler.PreprocessRestaurants)
	calculatorGroup.GET("/restaurants", s.dependencies.CalculatorHandler.Calculate)
}
