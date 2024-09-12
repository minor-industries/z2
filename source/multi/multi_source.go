package multi

import (
	"fmt"
	"github.com/minor-industries/z2/source"
)

type Source struct {
	handlers map[string]source.Source
}

func NewSource(sources []source.Source) (*Source, error) {
	handlers := map[string]source.Source{}

	for _, source := range sources {
		for _, svc := range source.Services() {
			for _, ch := range source.Characteristics() {
				key := fmt.Sprintf("%s::%s", svc, ch)

				if _, present := handlers[key]; present {
					return nil, fmt.Errorf("duplicate registration: %s", key)
				}
				handlers[key] = source
			}
		}
	}

	return &Source{handlers: handlers}, nil
}

func (s *Source) Convert(msg source.Message) []source.Value {
	key := fmt.Sprintf("%s::%s", msg.Service, msg.Characteristic)
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
