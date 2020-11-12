// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subscription

import (
	"context"
	endpointapi "github.com/onosproject/onos-e2sub/api/e2/endpoint/v1beta1"
	subapi "github.com/onosproject/onos-e2sub/api/e2/subscription/v1beta1"
	subtaskapi "github.com/onosproject/onos-e2sub/api/e2/task/v1beta1"
	"github.com/onosproject/onos-e2t/pkg/southbound/e2/channel"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"io"
	"sync"
)

const queueSize = 100

// Watcher is a subscription watcher
type Watcher struct {
	endpointID endpointapi.ID
	tasks      subtaskapi.E2SubscriptionTaskServiceClient
	log        logging.Logger
	cancel     context.CancelFunc
	mu         sync.Mutex
}

// Start starts the subscription watcher
func (w *Watcher) Start(ch chan<- controller.ID) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	request := &subtaskapi.WatchSubscriptionTasksRequest{}
	stream, err := w.tasks.WatchSubscriptionTasks(ctx, request)
	if err != nil {
		cancel()
		return err
	}
	w.cancel = cancel

	go func() {
		for {
			response, err := stream.Recv()
			if err == io.EOF || err == context.Canceled {
				break
			}
			if err != nil {
				w.log.Error(err)
			} else if response.Event.Task.TerminationEndpointID == w.endpointID {
				ch <- controller.NewID(response.Event.Task.ID)
			}
		}
		close(ch)
	}()
	return nil
}

// Stop stops the subscription watcher
func (w *Watcher) Stop() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.mu.Unlock()
}

var _ controller.Watcher = &Watcher{}

// ChannelWatcher is a channel watcher
type ChannelWatcher struct {
	endpointID endpointapi.ID
	tasks      subtaskapi.E2SubscriptionTaskServiceClient
	subs       subapi.E2SubscriptionServiceClient
	channels   *channel.Manager
	log        logging.Logger
	cancel     context.CancelFunc
	mu         sync.Mutex
}

// Start starts the channel watcher
func (w *ChannelWatcher) Start(ch chan<- controller.ID) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		return nil
	}

	channelCh := make(chan channel.Channel, queueSize)
	ctx, cancel := context.WithCancel(context.Background())
	err := w.channels.Watch(ctx, channelCh)
	if err != nil {
		cancel()
		return err
	}
	w.cancel = cancel

	go func() {
		for c := range channelCh {
			request := &subtaskapi.ListSubscriptionTasksRequest{}
			response, err := w.tasks.ListSubscriptionTasks(ctx, request)
			if err != nil {
				w.log.Error(err)
			} else {
				for _, task := range response.Task {
					if task.TerminationEndpointID == w.endpointID {
						subRequest := &subapi.GetSubscriptionRequest{
							ID: task.SubscriptionID,
						}
						subResponse, err := w.subs.GetSubscription(ctx, subRequest)
						if err != nil {
							w.log.Error(err)
						} else if subResponse.Subscription.E2NodeID == subapi.E2NodeID(c.ID()) {
							ch <- controller.NewID(task.ID)
						}
					}
				}
			}
		}
		close(ch)
	}()
	return nil
}

// Stop stops the channel watcher
func (w *ChannelWatcher) Stop() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.mu.Unlock()
}