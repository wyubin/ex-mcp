package grpcio

type Routes struct {
	routes []Route
}

func NewRoutes() *Routes {
	return &Routes{routes: []Route{}}
}

func (s *Routes) Add(route Route) {
	s.routes = append(s.routes, route)
}

func (s *Routes) Match(method, path string) (Route, InfoRequest) {
	var (
		info InfoRequest
		err  error
	)

	for _, route := range s.routes {
		info, err = route.MatchInfo(method, path)
		if err == nil {
			return route, info
		}
	}
	return nil, info
}
