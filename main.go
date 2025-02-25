package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/k3vwdd/aggreGATOR/internal/commands"
	"github.com/k3vwdd/aggreGATOR/internal/config"
	"github.com/k3vwdd/aggreGATOR/internal/database"
    _ "github.com/lib/pq"
)

func main() {

    cfg, err := config.Read()
    if err != nil {
        fmt.Println("error reading config")
        os.Exit(1)
    }

    dbURL := cfg.DBUrl
    db, err := sql.Open("postgres", dbURL)
    dbQueries := database.New(db)

    programState := Commands.State{
        Config: &cfg,
        Db: dbQueries,
    }

    cmds := Commands.Commands{
        Handlers: map[string]func(*Commands.State, Commands.Command) error{},
    }

    cmds.Register("login", Commands.HandlerLogin)
    cmds.Register("register", Commands.HandlerRegister)
    cmds.Register("reset", Commands.HandlerReset)
    cmds.Register("users", Commands.HandlerUsers)

    if len(os.Args) < 2 {
        fmt.Println("not enough args")
        os.Exit(1)
    }

    cmd := Commands.Command{
        Name: os.Args[1],
        Args: os.Args[2:],
    }

    if err := cmds.Run(&programState, cmd); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
