package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"time-tracker/internal/database"
	"time-tracker/internal/models"
)

var usersTemplate = template.Must(template.ParseFiles("internal/templates/user.html"))
var usersEffortTemplate = template.Must(template.ParseFiles("internal/templates/user_efforts.html"))

func AddUser(w http.ResponseWriter, r *http.Request) {
	var userReq models.UserRequest
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body"), http.StatusBadRequest)
		return
	}

	passport := userReq.PassportNumber

	parts := strings.Fields(passport)

	if len(parts) != 2 {
		http.Error(w, fmt.Sprintf("Invalid passport number format: %s", passport), http.StatusBadRequest)
		return
	}

	exist, err := database.CheckUserByPassport(passport)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking user: %s", err), http.StatusInternalServerError)
		return
	}
	if exist {
		http.Error(w, fmt.Sprintf("User with passport %s already exists", passport), http.StatusConflict)
		return
	}

	apiURL := os.Getenv("API_URL")
	params := url.Values{}
	params.Add("passportSerie", parts[0])
	params.Add("passportNumber", parts[1])

	resp, err := http.Get(fmt.Sprintf("%s?%s", apiURL, params.Encode()))
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Failed to get user info from externalAPI: %s", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	apiUser := models.User{PassportNumber: passport}
	err = json.NewDecoder(resp.Body).Decode(&apiUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed parsing user info from externalAPI: %s", err), http.StatusInternalServerError)
		return
	}

	err = database.SaveUser(apiUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed saving user: %s", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("User added successfully")))
}

func GetUsers(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")

	surname := query.Get("surname")
	name := query.Get("name")
	patronymic := query.Get("patronymic")
	address := query.Get("address")
	passportNumber := query.Get("passportNumber")

	page, err := strconv.Atoi(pageStr)
	if page < 1 || err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || err != nil {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	users, err := database.GetUsers(pageSize, offset, name, surname, patronymic, passportNumber, address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting users: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = usersTemplate.Execute(w, users)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetWorkLog(w http.ResponseWriter, r *http.Request) {
	userIdString := chi.URLParam(r, "id")
	startPeriodString := r.URL.Query().Get("startPeriod")
	endPeriodString := r.URL.Query().Get("endPeriod")

	userId, err := strconv.Atoi(userIdString)
	if err != nil || userId < 1 {
		http.Error(w, fmt.Sprintf("Invalid user id: %v", err), http.StatusBadRequest)
		return
	}
	var tasks []models.Task

	if startPeriodString == "" || endPeriodString == "" {
		tasks, err = database.GetTasks(userId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting tasks: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		startPeriod, err := time.Parse(time.RFC3339, startPeriodString)
		if err != nil {
			http.Error(w, "Invalid start period format", http.StatusBadRequest)
			return
		}

		endPeriod, err := time.Parse(time.RFC3339, endPeriodString)
		if err != nil {
			http.Error(w, "Invalid end period format", http.StatusBadRequest)
			return
		}

		tasks, err = database.GetTasksByPeriod(userId, startPeriod, endPeriod)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting tasks: %v", err), http.StatusInternalServerError)
			return
		}
	}
	userEfforts := models.CalculateUserEffort(tasks)
	models.SortUserEfforts(userEfforts)

	w.Header().Set("Content-Type", "text/html")
	err = usersEffortTemplate.Execute(w, userEfforts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

func StartTask(w http.ResponseWriter, r *http.Request) {

	userIdString := chi.URLParam(r, "id")

	userID, err := strconv.Atoi(userIdString)
	if err != nil || userID < 1 {
		http.Error(w, fmt.Sprintf("Invalid user ID: %v", userIdString), http.StatusBadRequest)
		return
	}

	err = database.StartTaskTimer(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting task: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Task-Timer started")))
}

func StopTask(w http.ResponseWriter, r *http.Request) {
	taskIdString := chi.URLParam(r, "id")

	taskID, err := strconv.Atoi(taskIdString)
	if err != nil || taskID < 0 {
		http.Error(w, fmt.Sprintf("Invalid task ID: %v", taskIdString), http.StatusBadRequest)
		return
	}

	exist, err := database.CheckTaskExist(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("The task is completed or does not exist"), http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(w, fmt.Sprintf("Task with id %d not exist", taskID), http.StatusBadRequest)
		return
	}
	err = database.StopTaskTimer(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error stoptask: %v", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Task-Timer stopped")))
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIdString := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(userIdString)
	if err != nil || userId < 1 {
		http.Error(w, fmt.Sprintf("Invalid user ID: %v", userIdString), http.StatusBadRequest)
		return
	}

	err = database.DeleteUser(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting user: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted"))

}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIdString := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(userIdString)
	if err != nil || userId < 1 {
		http.Error(w, fmt.Sprintf("Invalid user ID: %v", userIdString), http.StatusBadRequest)
		return
	}
	exist, err := database.CheckUserExist(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting the user from the database: %v", err), http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(w, fmt.Sprintf("Invalid user ID: %v", userId), http.StatusBadRequest)
	}

	query := r.URL.Query()

	surname := query.Get("surname")
	name := query.Get("name")
	patronymic := query.Get("patronymic")
	address := query.Get("address")
	passportNumber := query.Get("passportNumber")

	err = database.UpdateUser(userId, surname, name, patronymic, passportNumber, address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User updated"))
}
