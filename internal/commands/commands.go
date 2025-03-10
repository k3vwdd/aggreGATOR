package Commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/k3vwdd/aggreGATOR/internal/config"
	"github.com/k3vwdd/aggreGATOR/internal/database"
	"github.com/k3vwdd/aggreGATOR/internal/rss"
)

type State struct {
	// Before we can worry about command handlers,
	// we need to think about how we will give our handlers access to the application state
	// (later the database connection, but, for now, the config file).
	// Create a state struct that holds a pointer to a config.
	//     *config => /internal/config => (type Config struct {})
	Config *config.Config
	Db     *database.Queries
}

type Command struct {
	// Create a command struct. A command contains a name AND A SLICE OF STRING ARGUMENTS.
	// For example, in the case of the login command,
	// the name would be "login" and the handler will expect the arguments
	// slice to contain one string, the username.
	Name string   // Function name "login"
	Args []string // eg. "k3vwd" slice one username.
}

type Commands struct {
	// Command Name - the value
	Handlers map[string]func(*State, Command) error
}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Unable to get user: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("Login Command requires a single argument: gator login <username>")
	}

	loggedInUser := cmd.Args[0]
	user, err := s.Db.GetUser(context.Background(), loggedInUser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "User doesn't exists: %v\n", err)
		os.Exit(1)
	}

	s.Config.SetUser(user.Name)
	fmt.Printf("%v has been set:", s.Config.CurrentUserName)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("Register requires a single arument")
	}

	userToRegister := cmd.Args[0]
	createUser, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		Name:      userToRegister,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error registering user: %v\n", err)
		os.Exit(1)
	}
	s.Config.SetUser(createUser.Name)
	fmt.Printf("User %s created:", userToRegister)
	fmt.Println(createUser.UpdatedAt)
	fmt.Println(createUser.CreatedAt)
	fmt.Println(createUser.ID)

	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.Db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error trying to delete users")
	}
	fmt.Println("All users removed")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	dbusers, err := s.Db.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list users")
	}
	loggedInUser := s.Config.CurrentUserName
	for _, user := range dbusers {
		if user.Name == loggedInUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)

	}

	timeBetweenRequest, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Unable to parse string into a Duration: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ticker := time.NewTicker(timeBetweenRequest)
	defer ticker.Stop()
	for ctx.Err() == nil {
		err := HandlerScrapeFeeds(s, cmd, database.User{})
		if err != nil {
			return fmt.Errorf("Unable to scrape Feed: %w", err)
		}
		//urlsToFetch, err := s.Db.GetFeeds(ctx)
		//if err != nil {
		//    return fmt.Errorf("Error GetFeeds(): %w", err)
		//}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			fmt.Println("Recieved Interrupt stop")
			return nil

		}
		//for _, url := range urlsToFetch {
		//    rssData, err := rss.FetchFeed(ctx, url.Url)
		//    if err != nil {
		//        fmt.Printf("Error fetching URL %s: %v\n", url.Url, err)
		//        continue
		//    }
		//    fmt.Printf("fetching feed %s\n", url.Url)
		//    fmt.Println(rssData)
		//}
	}
	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("addfeed commands needs an argument")
	}

	feedName := cmd.Args[0]
	url := cmd.Args[1]

	createFeed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		Name:      feedName,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Url:       url,
		UserID:    user.ID,
	})

	if err != nil {
		return fmt.Errorf("Error trying to createFeed in DB")
	}

	// After successfully creating the feed...
	_, err = s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: createFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}

	fmt.Println("Feed created")
	fmt.Println(createFeed.Name)
	fmt.Println(createFeed.ID)
	fmt.Println(createFeed.CreatedAt)
	fmt.Println(createFeed.UpdatedAt)
	fmt.Println(createFeed.Url)
	fmt.Println(createFeed.UserID)

	return nil
}

func HandlerListFeeds(s *State, cmd Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Unable to get Feeds")
	}
	for _, val := range feeds {
		fmt.Println(val.FeedsName)
		fmt.Println(val.Url)
		fmt.Println(val.UsersName)
	}
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unfollow needs a url")
	}

	unfollow := s.Db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    cmd.Args[0],
	})

	if unfollow != nil {
		return fmt.Errorf("Unable to delete follow feed")
	}

	return nil
}

func HandlerScrapeFeeds(s *State, cmd Command, user database.User) error {
	nextFeed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Unable to getNextFeed: %w", err)
	}
	scrapeFeed(s.Db, nextFeed)

	return nil
}

func scrapeFeed(db *database.Queries, feed database.Feed) {

	err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Println("Unable to markFeedFetched")
	}

	rssData, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("Failed fetching feed at %s: %v", feed.Url, err)
		return
	}

	const layout = time.RFC1123
	for _, title := range rssData.Channel.Item {
		description := title.Description
		if description == "" {
			description = "No description"
		}
		pubDate, err := time.Parse(layout, title.PubDate)
		if err != nil {
			log.Printf("Failed to parse publish date (%s), skipping post: %v", title.PubDate, err)
			continue
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       title.Title,
			Url:         title.Link,
			Description: title.Description,
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				log.Println("Duplicate post URL detected, skipping insertion.")
				continue
			}
			log.Println("Error creating post:", err)
		}
	}
}

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 2 {
		if cmd.Args[0] != "limit" && cmd.Args[0] != "LIMIT" {
			return fmt.Errorf("need 'limit' or 'LIMIT'")
		}

		var err error
		limit, err = strconv.Atoi(cmd.Args[1])
		if err != nil {
			return fmt.Errorf("Error: invalid limit value: %w", err)
		}
	} else if len(cmd.Args) != 0 {
		return fmt.Errorf("invalid command usage: expected 'browse' or 'browse limit <number>'")
	}

	posts, err := s.Db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}


    fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("from %s\n", post.PublishedAt.Format("Mon Jan 2"))
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("follow needs a url")
	}

	url := cmd.Args[0]
	feedUrl, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Error trying to fetch feed url from db: => %w", err)
	}

	newFeedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedUrl.ID,
	})

	if err != nil {
		return fmt.Errorf("Error %w", err)
	}

	for _, val := range newFeedFollow {
		fmt.Println(val.FeedName)
		fmt.Println(val.UserName)
	}
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	feedsFollower, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Error trying to get feed follwers for user: %w", err)
	}

	for _, feed := range feedsFollower {
		fmt.Printf("%s\n", feed.FeedName)
	}

	return nil
}

// This method registers a new handler function for a Command name.
func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Handlers[name] = f
}

// This method runs a given Command with the provided State if it exists.
func (c *Commands) Run(s *State, cmd Command) error {
	handlers, exists := c.Handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("uknown command: %v", cmd.Name)
	}
	return handlers(s, cmd)
}
