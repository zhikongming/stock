package utils

import "sync"

type WrapData struct {
	Data interface{}
	Err  error
}

func ConcurrentActuator(f []func() (interface{}, error), limit int) ([]interface{}, error) {
	wg := sync.WaitGroup{}
	ch := make(chan struct{}, limit)
	dataCh := make(chan WrapData, len(f))
	for _, act := range f {
		wg.Add(1)
		ch <- struct{}{}
		go func(act func() (interface{}, error)) {
			defer wg.Done()
			defer func() { <-ch }()
			data, err := act()
			dataCh <- WrapData{
				Data: data,
				Err:  err,
			}
		}(act)
	}
	wg.Wait()
	close(dataCh)
	ret := make([]interface{}, 0)
	for data := range dataCh {
		if data.Err != nil {
			return nil, data.Err
		}
		ret = append(ret, data.Data)
	}
	return ret, nil
}
