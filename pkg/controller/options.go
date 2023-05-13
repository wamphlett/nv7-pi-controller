package controller

type Opt func(*Controller)

func WithPublisher() Opt {
	return func(c *Controller) {}
}
