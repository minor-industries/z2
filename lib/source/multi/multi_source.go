package multi

import (
	"fmt"
	"github.com/minor-industries/z2/lib/source"
	"strings"
)

type Source struct {
	handlers map[string]source.Source
}

func NewSource() *Source {
	return &Source{handlers: map[string]source.Source{}}
}

func (s *Source) Add(src source.Source) error {
	for _, svc := range src.Services() {
		for _, ch := range src.Characteristics() {
			key := fmt.Sprintf("%s::%s", strings.ToLower(string(svc)), strings.ToLower(string(ch)))

			if _, present := s.handlers[key]; present {
				return fmt.Errorf("duplicate registration: %s", key)
			}
			s.handlers[key] = src
		}
	}

	return nil
}

func (s *Source) Convert(msg source.Message) []source.Value {
	key := fmt.Sprintf("%s::%s", strings.ToLower(string(msg.Service)), strings.ToLower(string(msg.Characteristic)))
	handler, ok := s.handlers[key]
	if !ok {
		return nil
	}
	return handler.Convert(msg)
}

func (s *Source) Services() []source.UUID {
	panic("this shouldn't be used for scanning")
}

func (s *Source) Characteristics() []source.UUID {
	panic("this shouldn't be used for scanning")
}
