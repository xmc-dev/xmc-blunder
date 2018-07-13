package config

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

var Consul *api.Client

var wg sync.WaitGroup

func startAgent(quit, started chan bool) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		cmd := exec.Command("consul", "agent", "-dev")
		err := cmd.Start()
		if err != nil {
			log.Fatal("Couldn't start consul agent:", err)
		}
		log.Printf("Consul listening")

		// Give it some sleep to be certain that consul is listening
		time.Sleep(1 * time.Second)
		started <- true

		<-quit
		log.Print("Stopping consul")
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("Couldn't kill consul agent:", err)
		}
	}()
}

func putKV(key, value string, ignore bool) {
	pair := &api.KVPair{Key: key, Value: []byte(value), Flags: 1}
	if ignore {
		pair.Flags = 0
	}
	_, err := Consul.KV().Put(pair, nil)
	if err != nil {
		log.Fatal("Couldn't put key "+key+" into consul:", err)
	}
}

func initClient(bogus bool) *api.Client {
	conf := api.DefaultConfig()
	conf.Address = "127.0.0.1:8500"
	if bogus {
		conf.Address = "127.0.0.1:8420"
	}
	c, err := api.NewClient(conf)
	if err != nil {
		log.Fatal("Couldn't init consul client:", err)
	}

	return c
}

func setupCleanup(q chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		q <- true
		wg.Wait()
		os.Exit(1)
	}()
}

func TestMain(m *testing.M) {
	q := make(chan bool, 1)
	setupCleanup(q)

	s := make(chan bool, 1)
	startAgent(q, s)

	// wait for the agent to start
	<-s

	Consul = initClient(false)

	exit := m.Run()
	q <- true
	wg.Wait()
	os.Exit(exit)
}

func TestReadConfig(t *testing.T) {
	cfg := []struct {
		Key    string
		Env    string
		Value  string
		Ignore bool
	}{
		{Key: "xmc/core/db/user", Env: "CFG_DB_USER", Value: "u s e r"},
		{Key: "xmc/core/db/pass", Env: "CFG_DB_PASS", Value: "p a s s"},
		{Key: "cfg/val", Env: "CFG_VAL", Value: "v a l", Ignore: true},
	}

	for _, c := range cfg {
		putKV(c.Key, c.Value, c.Ignore)
	}

	r := NewReader(Consul, "xmc/core", "cfg")
	err := r.ReadConfig()
	if err != nil {
		t.Fatal("Error while reading config:", err)
	}

	for _, c := range cfg {
		e := os.Getenv(c.Env)
		if c.Ignore {
			if len(e) > 0 {
				t.Fatalf("Config key '%s' (%s) marked as ignored was not ignored.", c.Key, c.Env)
			}
		} else {
			if e != c.Value {
				t.Fatalf("Wrong env value for %s. Expected '%s', got '%s'", c.Env, c.Value, e)
			}
		}
	}
}

func TestError(t *testing.T) {
	c := initClient(true)
	r := NewReader(c, "a", "b")
	err := r.ReadConfig()

	if err == nil {
		t.Fatal("Expected error")
	}
	t.Log(err)
}

func TestDoesntPanic(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("It panicked", r)
			}
		}()
		r := NewReader(Consul, "a", "b")
		r.MustReadConfig()
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("It didn't panick")
			} else {
				t.Log("Panic:", r)
			}
		}()
		c := initClient(true)
		r := NewReader(c, "a", "b")
		r.MustReadConfig()
	}()
}
