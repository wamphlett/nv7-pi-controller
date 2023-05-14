package controller

// Opt defines a controller option
type Opt func(*Controller)

// WithPublisher configures a publisher for the controller
func WithPublisher(p Publisher) Opt {
	return func(c *Controller) {
		c.publishers = append(c.publishers, p)
	}
}
