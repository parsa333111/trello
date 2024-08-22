package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DatabaseInfo struct {
	host          string
	port          string
	user          string
	password      string
	database_name string
}

var DB *sql.DB
var DBInfo *DatabaseInfo

func loadDatabaseConnectionConfig() {
	host, ok := os.LookupEnv("PQ_HOST")
	if !ok {
		log.Fatal("Missing enviroment variable PQ_HOST.")
	}

	port, ok := os.LookupEnv("PQ_PORT")
	if !ok {
		log.Fatal("Missing enviroment variable PQ_PORT.")
	}

	user, ok := os.LookupEnv("PQ_USER")
	if !ok {
		log.Fatal("Missing enviroment variable PQ_USER.")
	}

	password, ok := os.LookupEnv("PQ_PASSWORD")
	if !ok {
		log.Fatal("Missing enviroment variable PQ_PASSWORD.")
	}

	database_name, ok := os.LookupEnv("PQ_DBNAME")
	if !ok {
		log.Fatal("Missing enviroment variable PQ_DBNAME.")
	}

	DBInfo = &DatabaseInfo{
		host,
		port,
		user,
		password,
		database_name,
	}
}

func connectToDatabase() {
	var info_string string = fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		DBInfo.host, DBInfo.port, DBInfo.user, DBInfo.password, DBInfo.database_name)

	var err error
	DB, err = sql.Open("postgres", info_string)
	if err != nil {
		log.Fatalf("Failed to connect to the database %s with error: %s", info_string, err)
	}
}

func createWorkspaceTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS Workspace (
								id integer PRIMARY KEY generated always as identity, 
								name varchar(30) not null,
								description TEXT,
								created_at timestamp,
								updated_at timestamp
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()
}

func createTaskTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS Task (
								id integer PRIMARY KEY generated always as identity, 
								title varchar(30) not null,
								description TEXT,
								status varchar(30),
								estimated_time integer,
								actual_time integer,
								due_date timestamp,
								priority integer,
								workspace_id integer not null,
								assignee_id integer,
								created_at timestamp,
								updated_at timestamp,
								image_url varchar(100),
								FOREIGN KEY(workspace_id) REFERENCES Workspace(id) ON DELETE CASCADE,
								FOREIGN KEY(assignee_id) REFERENCES Users(id) ON DELETE CASCADE
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()

}

func createSubtaskTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS Subtask (
								id integer PRIMARY KEY generated always as identity, 
								task_id integer not null,
								title varchar(30) not null,
								is_completed varchar(30),
								assignee_id integer,
								created_at timestamp,
								updated_at timestamp,
								FOREIGN KEY(task_id) REFERENCES Task(id) ON DELETE CASCADE,
								FOREIGN KEY(assignee_id) REFERENCES Users(id) ON DELETE CASCADE							
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()

}

func createUserTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS Users (
								id integer PRIMARY KEY generated always as identity, 
								username varchar(30) not null,
								email varchar(30) not null,
								password_hash BYTEA,
								created_at timestamp,
								updated_at timestamp,
								UNIQUE (username),
								UNIQUE (email)
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()

}

func createUserWorkspaceRoleTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS UserWorkspaceRole (
								id integer PRIMARY KEY generated always as identity, 
								user_id integer not null,
								workspace_id integer not null,
								role varchar(30),
								created_at timestamp,
								updated_at timestamp,
								FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE,
								FOREIGN KEY(workspace_id) REFERENCES Workspace(id) ON DELETE CASCADE	
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()

}

func createCommentTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS Comment (
								id integer PRIMARY KEY generated always as identity, 
								task_id integer not null,
								user_id integer not null,
								text TEXT,
								FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE,
								FOREIGN KEY(task_id) REFERENCES Task(id) ON DELETE CASCADE
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()

}

func createWatchTable() {
	query, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS Watch (
								task_id integer not null,
								user_id integer not null,
								PRIMARY KEY (task_id, user_id),
								FOREIGN KEY(user_id) REFERENCES Users(id) ON DELETE CASCADE,
								FOREIGN KEY(task_id) REFERENCES Task(id) ON DELETE CASCADE
							)`)
	if err != nil {
		log.Fatal("Error:", err)
	}
	query.Exec()

}

func createTables() {
	createWorkspaceTable()
	createUserTable()
	createTaskTable()
	createSubtaskTable()
	createUserWorkspaceRoleTable()
	createCommentTable()
	createWatchTable()
}

func InitializeDatabase() {
	loadDatabaseConnectionConfig()
	connectToDatabase()
	createTables()
	RedisInitial()
}
