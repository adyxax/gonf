package gonf

var customPromises []*CustomPromise

func init() {
	customPromises = make([]*CustomPromise, 0)
}

type CustomPromiseInterface interface {
	Promise
	Status() Status
}

type CustomPromise struct {
	promise CustomPromiseInterface
}

func MakeCustomPromise(p CustomPromiseInterface) *CustomPromise {
	return &CustomPromise{
		promise: p,
	}
}

func (c *CustomPromise) IfRepaired(p ...Promise) Promise {
	c.promise.IfRepaired(p...)
	return c
}

func (c *CustomPromise) Promise() Promise {
	customPromises = append(customPromises, c)
	return c
}

func (c *CustomPromise) Resolve() {
	c.promise.Resolve()
}

func (c CustomPromise) Status() Status {
	return c.Status()
}

func resolveCustomPromises() (status Status) {
	status = KEPT
	for _, c := range customPromises {
		if c.promise.Status() == PROMISED {
			c.Resolve()
			switch c.promise.Status() {
			case BROKEN:
				return BROKEN
			case REPAIRED:
				status = REPAIRED
			}
		}
	}
	return
}
