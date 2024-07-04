package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"strings"
	"time"
	"time-tracker/internal/config"
	"time-tracker/internal/logger"
	"time-tracker/internal/models"
)

var db *sql.DB

func InitDB(config *config.Config) error {
	logger.Logger.Info("Initializing database connection")
	defer logger.Logger.Info("Database connection initialized")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBName)
	var err error

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrations: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://internal/database/migrations", config.DBName, driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	}
	return nil
}

func CheckUserByPassport(passportNumber string) (bool, error) {
	logger.Logger.Info("Checking user by passport number")
	defer logger.Logger.Info("Checked user by passport number")

	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE passport_number = $1)`

	err := db.QueryRow(query, passportNumber).Scan(&exist)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func GetUsers(pageSize, offset int, name, surname, patronymic, passportNumber, address string) ([]models.User, error) {
	logger.Logger.Info("Getting users")
	defer logger.Logger.Info("Done getting users")
	var users []models.User
	query := `SELECT id, surname, name, patronymic, passport_number, address FROM users`
	var conditions []string
	var args []interface{}
	argCount := 1

	if passportNumber != "" {
		conditions = append(conditions, fmt.Sprintf("passport_number = $%d", argCount))
		args = append(args, passportNumber)
		argCount++
	}

	if surname != "" {
		conditions = append(conditions, fmt.Sprintf("surname = $%d", argCount))
		args = append(args, surname)
		argCount++
	}

	if name != "" {
		conditions = append(conditions, fmt.Sprintf("name = $%d", argCount))
		args = append(args, name)
		argCount++
	}

	if patronymic != "" {
		conditions = append(conditions, fmt.Sprintf("patronymic = $%d", argCount))
		args = append(args, patronymic)
		argCount++
	}

	if address != "" {
		conditions = append(conditions, fmt.Sprintf("address = $%d", argCount))
		args = append(args, address)
		argCount++
	}

	if len(conditions) > 0 {
		query += "WHERE" + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users1: %v", err)
	}

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Surname, &user.Name, &user.Patronymic, &user.PassportNumber, &user.Address)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to retrieve users2: %v", err)
	}
	return users, nil

}

func CheckTaskExist(taskId int) (bool, error) {
	logger.Logger.Info("Checking task exists")
	defer logger.Logger.Info("Done checking task exists")
	query := `SELECT task_id FROM tasks 
              WHERE task_id = $1
			  AND end_time IS NULL`

	row := db.QueryRow(query, taskId)

	err := row.Scan(&taskId)
	if err != nil {
		return false, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return true, nil
}

func CheckUserExist(userID int) (bool, error) {
	logger.Logger.Info("Checking user exists")
	defer logger.Logger.Info("Done checking user exists")
	query := `SELECT id FROM users WHERE id = $1`

	row := db.QueryRow(query, userID)

	err := row.Scan(&userID)
	if err != nil {
		return false, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, nil
}

func StartTaskTimer(userId int) error {
	logger.Logger.Info("Starting task timer")
	defer logger.Logger.Info("Done starting task timer")
	query := `INSERT INTO tasks (user_id, start_time) VALUES ($1, $2)`

	_, err := db.Exec(query, userId, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func StopTaskTimer(taskID int) error {
	logger.Logger.Info("Stopping task timer")
	defer logger.Logger.Info("Done stopping task timer")

	query := `UPDATE tasks SET end_time = $1 WHERE task_id = $2`

	_, err := db.Exec(query, time.Now(), taskID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(userId int) error {
	logger.Logger.Info("Deleting user")
	defer logger.Logger.Info("Done deleting user")

	query := `DELETE FROM tasks WHERE user_id = $1`
	_, err := db.Exec(query, userId)
	if err != nil {
		return err
	}

	query = `DELETE FROM users WHERE id = $1`
	_, err = db.Exec(query, userId)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(userId int, surname, name, patronymic, passportNumber, address string) error {
	logger.Logger.Info("Updating user")
	defer logger.Logger.Info("Done updating user")
	query := `UPDATE users SET `

	var conditions []string
	var args []interface{}
	var argCount = 1

	if name != "" {
		conditions = append(conditions, fmt.Sprintf("name = $%d", argCount))
		args = append(args, name)
		argCount++
	}
	if surname != "" {
		conditions = append(conditions, fmt.Sprintf("surname = $%d", argCount))
		args = append(args, surname)
		argCount++
	}

	if patronymic != "" {
		conditions = append(conditions, fmt.Sprintf("patronymic = $%d", argCount))
		args = append(args, patronymic)
		argCount++
	}
	if passportNumber != "" {
		conditions = append(conditions, fmt.Sprintf("passport_number = $%d", argCount))
		args = append(args, passportNumber)
		argCount++
	}
	if address != "" {
		conditions = append(conditions, fmt.Sprintf("address = $%d", argCount))
		args = append(args, address)
		argCount++
	}

	query += strings.Join(conditions, ", ")
	query += fmt.Sprintf(" WHERE id = %d", userId)

	fmt.Println(query, args)

	_, err := db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func GetTasks(userId int) ([]models.Task, error) {
	logger.Logger.Info("Getting tasks")
	defer logger.Logger.Info("Done getting tasks")

	query := `SELECT user_id, task_id, start_time, end_time
 			  FROM tasks
 			  WHERE end_time IS NOT NULL
 			  AND user_id = $1`

	rows, err := db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.UserID, &task.TaskID, &task.StartTime, &task.EndTime); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil

}

func GetTasksByPeriod(userId int, startPeriod, endPeriod time.Time) ([]models.Task, error) {
	logger.Logger.Info("Getting tasks by period")
	defer logger.Logger.Info("Done getting tasks by period")
	query := `SELECT user_id, task_id, start_time, end_time
 			  FROM tasks
 			  WHERE end_time IS NOT NULL
 			  AND user_id = $3
 			  AND start_time >= $1 AND start_time <= $2`

	rows, err := db.Query(query, startPeriod, endPeriod, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.UserID, &task.TaskID, &task.StartTime, &task.EndTime); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func SaveUser(user models.User) error {
	logger.Logger.Info("Saving user")
	defer logger.Logger.Info("Done saving user")

	query := `INSERT INTO users (surname, name, patronymic, address, passport_number)
 			  VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(query, user.Surname, user.Name, user.Patronymic, user.Address, user.PassportNumber)
	if err != nil {
		return fmt.Errorf("failed to save user: %v", err)
	}
	return nil
}
