package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v3"
)

const todoDbFile = "./todosDB.db?_journal_mode=WAL&_busy_timeout=5000"

type Todo struct {
	ID            uint
	Description   string
	CompletedDate *string
	CreatedDate   string
}

func createDb() *sql.DB {
	db, err := sql.Open("sqlite3", todoDbFile)
	if err != nil {
		panic("Unable to access/create a database file")
	}
	queryCreate := `
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			todo TEXT,
			completed DATETIME,
			created DATETIME default current_timestamp
		);
	`
	_, err = db.Exec(queryCreate)
	if err != nil {
		panic("Unable to create a table in the database file")
	}
	return db
}

func addTodo(db *sql.DB, todo string) {
	_, err := db.Exec("INSERT INTO todos (todo, created) VALUES (?, DATETIME('now', 'localtime'))", todo)

	if err != nil {
		panic("Unable to add a todo")
	}
	fmt.Println("\n📖 Todo successfully added\n")
}

func listTodos(db *sql.DB) {
	// Nice-to-have : an option to list all including completed
	rows, err := db.Query("SELECT id, todo, DATE(DATETIME(created, 'localtime')) FROM todos WHERE completed IS NULL ORDER BY created DESC")
	if err != nil {
		panic("Unable to select rows from database")
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Description, &todo.CreatedDate); err != nil {
			fmt.Println("Unable to read db row")
			continue
		}
		todos = append(todos, todo)
	}
	t := table.New(os.Stdout)
	t.SetHeaders("ID", "Todo", "Created date")

	for _, todo := range todos {
		t.AddRow(strconv.FormatUint(uint64(todo.ID), 10), todo.Description, todo.CreatedDate)
	}
	t.Render()
}

func completeTodo(db *sql.DB, id uint) {
	_, err := db.Exec("UPDATE todos SET completed = DATETIME('now', 'localtime') WHERE id = ?", id)
	if err != nil {
		panic("Unable to update a todo in DB")
	}
}

func deleteTodo(db *sql.DB, id uint) {
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		panic("Unable to delete a todo in DB")
	}
}

func main() {
	var completedId uint
	var deletedId uint

	db := createDb()
	defer db.Close()
	cmd := &cli.Command{
		Name:  "todo",
		Usage: "Manage your todo list",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("Hello World!")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a todo",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					addTodo(db, cmd.Args().First())
					listTodos(db)
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "List all todos",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					listTodos(db)
					return nil
				},
			},
			{
				Name:    "complete",
				Aliases: []string{"c"},
				Arguments: []cli.Argument{
					&cli.UintArg{
						Name:        "id",
						Destination: &completedId,
					},
				},
				Usage: "Complete a todo",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					completeTodo(db, completedId)
					listTodos(db)
					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a todo",
				Arguments: []cli.Argument{
					&cli.UintArg{
						Name:        "id",
						Destination: &deletedId,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					deleteTodo(db, deletedId)
					listTodos(db)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
