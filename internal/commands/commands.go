package Commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/k3vwdd/aggreGATOR/internal/config"
	"github.com/k3vwdd/aggreGATOR/internal/database"
)

type State struct {
    // Before we can worry about command handlers,
    // we need to think about how we will give our handlers access to the application state
    // (later the database connection, but, for now, the config file).
    // Create a state struct that holds a pointer to a config.
    //     *config => /internal/config => (type Config struct {})
    Config *config.Config
    Db *database.Queries
}

type Command struct {
    // Create a command struct. A command contains a name AND A SLICE OF STRING ARGUMENTS.
    // For example, in the case of the login command,
    // the name would be "login" and the handler will expect the arguments
    // slice to contain one string, the username.
    Name string // Function name "login"
    Args []string // eg. "k3vwd" slice one username.
}

type Commands struct {
            // Command Name - the value
    Handlers map[string]func(*State, Command) error
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
        Name: userToRegister,
        ID: uuid.New(),
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

// This method registers a new handler function for a Command name.
func (c *Commands) Register(name string, f func(*State, Command) error)  {
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
