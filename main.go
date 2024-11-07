package main

import (
	"GoBlogAggregator/internal/config"
	"GoBlogAggregator/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type Config = config.Config

type state struct {
	config *Config
	db     *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

// registers a new handler function for a command name
func (c *commands) registerHandler(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("no handler %s", cmd.name)
	}
	return handler(s, cmd)
}

// login as a user
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username given")
	}
	newUsername := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), newUsername)
	if err != nil {
		fmt.Printf("no user: %s\n", newUsername)
		os.Exit(1)
		return err
	}
	//set username
	err = s.config.SetUser(newUsername)
	if err != nil {
		return err
	}
	fmt.Println("username has been set")
	return nil
}

// register a new user
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no name given")
	}
	name := cmd.args[0]

	//existing user check
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Printf("user %v already exists\n", name)
		os.Exit(1)
	}
	if err != sql.ErrNoRows {
		return err
	}

	args := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}
	user, err := s.db.CreateUser(context.Background(), args)
	if err != nil {
		return err
	}
	err = s.config.SetUser(name)
	if err != nil {
		return err
	}
	userJSON, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println(string(userJSON))
	return nil
}

// delete all users
func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("All data in users table deleted\n")
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
	//db connection
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("Failed to open connection to db: %s", err)
	}
	dbQueries := database.New(db)
	state.db = dbQueries

	//register commands
	commands.registerHandler("login", handlerLogin)
	commands.registerHandler("register", handlerRegister)
	commands.registerHandler("reset", handlerReset)
	if len(os.Args) < 2 {
		log.Fatalf("no command given")
	}

	//run commands
	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	err = commands.run(state, cmd)
	if err != nil {
		log.Fatalf("%s", err)
	}
	os.Exit(0)

}
