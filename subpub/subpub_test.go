package subpub_test

import (
	"context"
	"testing"
	"time"

	"awesomeProject1/subpub"
)

func TestBasicPubSub(t *testing.T) {
	bus := subpub.NewSubPub()
	got := ""

	sub, _ := bus.Subscribe("news", func(msg interface{}) {
		got = msg.(string)
	})

	_ = bus.Publish("news", "hello")
	time.Sleep(100 * time.Millisecond)

	if got != "hello" {
		t.Fatalf("хотели 'hello', но получили '%s'", got)
	}

	got = ""

	sub.Unsubscribe()

	_ = bus.Publish("news", "ignored")
	time.Sleep(200 * time.Millisecond)

	if got != "" {
		t.Errorf("После отписки не ожидали сообщений, но получили: %s", got)
	}
}

func TestSlowSubscriber(t *testing.T) {
	bus := subpub.NewSubPub()
	var fastReceived, slowReceived int

	bus.Subscribe("news", func(msg interface{}) {
		fastReceived++
	})

	bus.Subscribe("news", func(msg interface{}) {
		time.Sleep(200 * time.Millisecond)
		slowReceived++
	})

	start := time.Now()
	for i := 0; i < 5; i++ {
		_ = bus.Publish("news", i)
	}
	duration := time.Since(start)

	if duration > 200*time.Millisecond {
		t.Errorf("Publish block slow sub: %v", duration)
	}

	time.Sleep(300 * time.Millisecond)

	if fastReceived == 0 {
		t.Error("Быстрый подписчик не получил сообщения")
	}
	if slowReceived == 0 {
		t.Error("Публикация блокируется медленным подписчиком")
	}
}

func TestCloseStopsSubscribers(t *testing.T) {
	bus := subpub.NewSubPub()
	received := 0

	bus.Subscribe("news", func(msg interface{}) {
		received++
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := bus.Close(ctx)
	if err != nil {
		t.Fatalf("Close with error: %v", err)
	}

	_ = bus.Publish("news", "late")
	time.Sleep(50 * time.Millisecond)

	if received > 0 {
		t.Errorf("Sub received message after close: %v", received)
	}
}
