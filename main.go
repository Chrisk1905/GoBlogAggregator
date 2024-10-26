package main

import (
	"GoBlogAggregator/internal/config"
	"fmt"
	"log"
	"os"
)

type Config = config.Config

type state struct {
	config *Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

// registers a new handler function for a command name
func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("no handler %s", cmd.name)
	}
	return handler(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username given")
	}
	//set username
	newUsername := cmd.args[0]
	err := s.config.SetUser(newUsername)
	if err != nil {
		return err
	}
	fmt.Println("username has been set")
	return nil
}

func main() {
	//config
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	commands := &commands{
		handlers: make(map[string]func(*state, command) error),
	}
	state := &state{
		config: &cfg,
	}
	commands.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatalf("no command given")
	}
	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	err = commands.run(state, cmd)
	if err != nil {
		log.Fatalf("%s", err)
	}
}
