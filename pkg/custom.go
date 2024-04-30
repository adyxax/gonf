package gonf

var customPromises []*CustomPromise

func init() {
	customPromises = make([]*CustomPromise, 0)
}

type CustomPromise struct {
	promise Promise
}

func MakeCustomPromise(p Promise) *CustomPromise {
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
	return c.promise.Status()
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
