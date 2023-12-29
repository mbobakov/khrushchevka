package flow

import (
	"context"
	"fmt"
)

// Flow is the interface for implementation of ligths patterns
type Flow interface {
	Start(ctx context.Context) error
	Stop()
	Name() string
}

// Controller executes the selected flow
type Controller struct {
	registry    []Flow
	currentFlow Flow
	errorChan   chan error
}

func NewController(flows ...Flow) *Controller {
	return &Controller{
		registry:  flows,
		errorChan: make(chan error),
	}
}

// SubscribeToErrors returns the channel to subscribe to errors
func (c *Controller) SubscribeToErrors() <-chan error {
	return c.errorChan
}

// SelectFlow selects the flow to be executed
// previous flow will be stopped
func (c *Controller) SelectFlow(ctx context.Context, name string) error {
	if c.currentFlow != nil {
		c.currentFlow.Stop()
	}

	for _, flow := range c.registry {
		if flow.Name() == name {
			c.currentFlow = flow
			go func() {
				c.errorChan <- flow.Start(ctx)
			}()

			return nil
		}
	}
	return fmt.Errorf("flow %s not found", name)
}

// ListFlows returns the list of available flows
func (c *Controller) ListFlows() []string {
	var flows []string
	for _, flow := range c.registry {
		flows = append(flows, flow.Name())
	}
	return flows
}
