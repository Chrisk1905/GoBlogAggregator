package main

import (
	"GoBlogAggregator/internal/config"
	"GoBlogAggregator/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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

// prints all users
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return nil
	}

	for _, user := range users {
		if s.config.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

// prints a blog feed https://www.wagslane.dev/index.xml
func handlerAgg(s *state, cmd command) error {
	rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", rssFeed)
	return nil
}

// Get current user from the database, and make a new feed row
// args{
// name: name of feed
// url: url of feed }
func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("must provide args and url")
	}
	var feedName string = cmd.args[0]
	var feedUrl string = cmd.args[1]

	_, err := url.Parse(feedUrl)
	if err != nil {
		return fmt.Errorf("invalid URL provided: %v", err)
	}

	// find the user
	currentUserName := s.config.CurrentUserName
	user, err := s.db.GetUser(context.Background(), currentUserName)
	if err != nil {
		return err
	}
	///
	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: feedName, Valid: true},
		Url:       sql.NullString{String: feedUrl, Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	}
	// add new feed row
	feed, err := s.db.CreateFeed(context.Background(), feedParams)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", feed)
	return nil

}

// prints all feeds: name, url, name of the user who created the feed
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		feedUser, err := s.db.GetUserByID(context.Background(), feed.UserID.UUID)
		if err != nil {
			return err
		}

		fmt.Printf("{ name: %s url: %s user: %s }\n", feed.Name.String, feed.Url.String, feedUser.Name)
	}

	return nil
}

// fetch a feed from the given URL, return an RSSFeed struct
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req.Header.Add("User-Agent", "Gator")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	xmlBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	rssFeed := &RSSFeed{}
	err = xml.Unmarshal(xmlBytes, rssFeed)
	if err != nil {
		return nil, err
	}
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.EscapeString(rssFeed.Channel.Description)
	for _, item := range rssFeed.Channel.Item {
		item.Description = item.Description
		item.Title = item.Title

	}
	return rssFeed, nil
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
	commands.registerHandler("users", handlerUsers)
	commands.registerHandler("agg", handlerAgg)
	commands.registerHandler("addfeed", handlerAddFeed)
	commands.registerHandler("feeds", handlerFeeds)
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
