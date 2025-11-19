package grpcio

import (
	"fmt"
	"net/url"
	"regexp"
)

type Route interface {
	SetClusterName(name string)
	Name() string
	MatchInfo(method, path string) (InfoRequest, error)
	RequestCov(info InfoRequest, jsonBody []byte) ([]byte, error)
	ResponseCov(info InfoRequest, grpcBody []byte) ([]byte, error)
}

type RequestMapper struct {
	name        string
	clusterName string
	entries     []routeEntry
	path2Cov    map[string]BodyCov
}

type routeEntry struct {
	method   string
	pattern  string
	regex    *regexp.Regexp
	keys     []string
	grpcPath string
	cov      BodyCov
}

func NewRequestMapper(name string) *RequestMapper {
	return &RequestMapper{
		name:     name,
		entries:  []routeEntry{},
		path2Cov: map[string]BodyCov{},
	}
}

// implement by hoster
func (s *RequestMapper) RegisterRoutes() {
	// s.Register("GET /v1/user/{id}", "/user.UserService/GetUser")
	// s.Register("POST /v1/user", "/user.UserService/CreateUser")
}

// shared method
func (s *RequestMapper) SetClusterName(name string) {
	s.clusterName = name
}

func (s *RequestMapper) Name() string {
	return s.name
}

func (s *RequestMapper) Register(pattern, grpcPath string, converter BodyCov) {
	// pattern can be like "GET /v1/user/{id}" or just "/v1/user/{id}"
	method := ""
	pathPattern := pattern
	if p := regexp.MustCompile(`^[A-Z]+\s+`).FindString(pattern); p != "" {
		method = p[:len(p)-1]
		pathPattern = pattern[len(p):]
	}

	keys := extractParamkeys(pathPattern)
	re := compilePatternToRegex(pathPattern)

	s.entries = append(s.entries, routeEntry{
		method:   method,
		pattern:  pathPattern,
		regex:    re,
		keys:     keys,
		grpcPath: grpcPath,
		cov:      converter,
	})
	s.path2Cov[grpcPath] = converter
}

func (s *RequestMapper) MatchInfo(method, path string) (InfoRequest, error) {
	// path may include query string; parse it
	u, err := url.Parse(path)
	if err != nil {
		return InfoRequest{}, fmt.Errorf("invalid path: %w", err)
	}

	for _, e := range s.entries {
		if e.method != "" && e.method != method {
			continue
		}
		if !e.regex.MatchString(u.Path) {
			continue
		}
		matches := e.regex.FindStringSubmatch(u.Path)
		params := map[string]string{}
		if len(e.keys) > 0 {
			for i, k := range e.keys {
				// first submatch is at index 1
				if i+1 < len(matches) {
					params[k] = matches[i+1]
				} else {
					params[k] = ""
				}
			}
		}

		info := InfoRequest{
			ClusterName: s.clusterName,
			PathGrpc:    e.grpcPath,
			PathParams:  params,
			Query:       u.Query(),
		}
		return info, nil
	}
	return InfoRequest{}, fmt.Errorf("no route match for %s %s", method, path)
}

func (s *RequestMapper) RequestCov(info InfoRequest, jsonBody []byte) ([]byte, error) {
	converter, found := s.path2Cov[info.PathGrpc]
	if !found {
		return nil, fmt.Errorf("no path match for %s", info.PathGrpc)
	}
	return converter.Json2Grpc(info, jsonBody)
}

func (s *RequestMapper) ResponseCov(info InfoRequest, grpcBody []byte) ([]byte, error) {
	converter, found := s.path2Cov[info.PathGrpc]
	if !found {
		return nil, fmt.Errorf("no path match for %s", info.PathGrpc)
	}
	return converter.Grpc2Json(info, grpcBody)
}

// -- private method --  //
// compilePatternToRegex transforms patterns like "/v1/user/{id}" into
// regular expressions capturing param values in the same order as extractParamkeys.
func compilePatternToRegex(pattern string) *regexp.Regexp {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	// replace each {param} with a capture group matching non-slash chars
	regexStr := re.ReplaceAllString(pattern, `([^/]+)`)
	// ensure full match
	regexStr = "^" + regexStr + "$"
	return regexp.MustCompile(regexStr)
}
