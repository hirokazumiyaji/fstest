package fstest

import (
	"context"
	"runtime"
	"sync"

	"cloud.google.com/go/firestore"
)

var defaultContext *Context

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defaultContext = &Context{mutex: new(sync.Mutex), instances: make(map[string]*instanceContext)}
}

type instanceContext struct {
	counter  int
	instance Instance
	client   *firestore.Client
}

type Context struct {
	mutex     *sync.Mutex
	instances map[string]*instanceContext
}

func Setup(options *Options) (*firestore.Client, error) {
	background := context.Background()
	defaultContext.mutex.Lock()
	defer defaultContext.mutex.Unlock()
	ctx, ok := defaultContext.instances[options.ProjectId]
	if ok {
		ctx.counter++
		return ctx.client, nil
	}
	instance, err := NewInstance(options)
	if err != nil {
		return nil, err
	}
	client, err := firestore.NewClient(background, options.ProjectId)
	if err != nil {
		return nil, err
	}
	c := &instanceContext{
		counter:  1,
		instance: instance,
		client:   client,
	}
	defaultContext.instances[options.ProjectId] = c
	return c.client, nil
}

func Teardown(projectId string) error {
	defaultContext.mutex.Lock()
	defer defaultContext.mutex.Unlock()
	ctx, ok := defaultContext.instances[projectId]
	if !ok {
		return nil
	}
	ctx.counter--
	if ctx.counter > 0 {
		return nil
	}
	err := ctx.instance.Close()
	if err != nil {
		return err
	}
	delete(defaultContext.instances, projectId)
	return nil
}
