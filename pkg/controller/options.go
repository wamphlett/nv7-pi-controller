package controller

type Opt func(*Controller)

func WithPublisher(p Publisher) Opt {
	return func(c *Controller) {
		c.publishers = append(c.publishers, p)
	}
}
