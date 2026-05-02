package observe

import "context"

type Inhibitor struct {
	What string `json:"what"`
	Who  string `json:"who"`
	Why  string `json:"why,omitempty"`
	UID  int    `json:"uid"`
}

type Source interface {
	List(ctx context.Context) ([]Inhibitor, error)
}

type NoopSource struct{}

func (NoopSource) List(context.Context) ([]Inhibitor, error) {
	return nil, nil
}
