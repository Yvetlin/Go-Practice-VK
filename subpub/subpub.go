// Author: Yvetlin [Gloria]
// Date: 2025
package subpub

import (
	"context"
	"log"
	"sync"
)

type MessageHandler func(msg interface{}) //кастом тип

type Subscription interface { //интерфейс
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subscriber struct { //структура
	ch      chan interface{}
	handle  MessageHandler
	stop    chan struct{}
	subject string
	parent  *subPub
	once    sync.Once
}

type subPub struct {
	sync.Mutex
	subscribers map[string][]*subscriber
}

func NewSubPub() SubPub {
	return &subPub{
		subscribers: make(map[string][]*subscriber),
	}
}

func (s *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) { //метод
	sub := &subscriber{
		ch:      make(chan interface{}, 16),
		handle:  cb,
		stop:    make(chan struct{}),
		subject: subject,
		parent:  s,
	}

	s.Lock()
	s.subscribers[subject] = append(s.subscribers[subject], sub)
	s.Unlock()

	// Вот это запускает обработку входящих сообщений:
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[INFO][subpub] recovered from panic in handler %v", r)
			}
		}()

		for {
			select {
			case msg := <-sub.ch:
				sub.handle(msg)
			case <-sub.stop:
				return
			}
		}
	}()

	return sub, nil
}

func (s *subscriber) Unsubscribe() {
	s.once.Do(func() {
		close(s.stop)

		s.parent.Lock()
		defer s.parent.Unlock()

		subs := s.parent.subscribers[s.subject]
		for i, sub := range subs {
			//Delete from slice
			if sub == s {
				s.parent.subscribers[s.subject] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	})
}

func (s *subPub) Publish(subject string, msg interface{}) error {
	s.Lock()
	subs := append([]*subscriber{}, s.subscribers[subject]...)
	s.Unlock()

	for _, sub := range subs {
		select {
		case sub.ch <- msg:
			// успешно отправлено
		default:
			// канал переполнен, лог или дроп
		}
	}
	return nil
}

func (s *subPub) Close(ctx context.Context) error {
	s.Lock()
	var wg sync.WaitGroup

	for _, subs := range s.subscribers {
		for _, sub := range subs {
			wg.Add(1)
			go func(sub *subscriber) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				default:
					sub.Unsubscribe()
				}
			}(sub)
		}
	}
	s.Unlock()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
