package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"task-management-system/models"
	"time"

	"github.com/gorilla/mux"
)

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new task and assign it to a user
// @Tags tasks
// @Accept  json
// @Produce  json
// @Param task body models.Task true "Task info"
// @Success 201 {object} models.Task
// @Failure 400 {object} string
// @Failure 401 {object} string
// @Failure 500 {object} string
// @Router /tasks [post]
func (db *AppHandler) CreateTask() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//user id'yi contexten alalım
		userID := r.Context().Value("userID").(int)
		task.UserID = userID

		//role'ü de contexten alalım
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Only admin can create tasks", http.StatusUnauthorized)
			return
		}

		//tarih formatı kontrolü
		var err error
		task.StartDate, err = time.Parse(time.RFC3339, task.StartDate.Format(time.RFC3339))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		task.DueDate, err = time.Parse(time.RFC3339, task.DueDate.Format(time.RFC3339))
		if err != nil {
			http.Error(w, "Invalid due date format", http.StatusBadRequest)
			return
		}

		_, err = db.DB.Exec("INSERT INTO tasks (title, description, status, start_date, due_date, user_id, assigned_to) VALUES (?, ?, ?, ?, ?, ?, ?)",
			task.Title, task.Description, task.Status, task.StartDate, task.DueDate, task.UserID, task.AssignedTo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	})
}

// UpdateTask godoc
// @Summary Update an existing task
// @Description Update an existing task with new details
// @Tags tasks
// @Accept  json
// @Produce  json
// @Param task_id path int true "Task ID"
// @Param task body models.Task true "Task info"
// @Success 200 {object} models.Task
// @Failure 400 {object} string
// @Failure 401 {object} string
// @Failure 500 {object} string
// @Router /tasks/{task_id} [put]
func (db *AppHandler) UpdateTask() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		taskID := vars["task_id"]

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userRole := r.Context().Value("role").(string)
		if userRole != "admin" {
			http.Error(w, "Only admin can update tasks", http.StatusInternalServerError)
			return
		}

		var existingTask models.Task
		row := db.DB.QueryRow("SELECT id, title, description, status, start_date, due_date, assigned_to FROM tasks WHERE id = ?", taskID)
		if err := row.Scan(&existingTask.ID, &existingTask.Title, &existingTask.Description, &existingTask.Status, &existingTask.StartDate, &existingTask.DueDate, &existingTask.AssignedTo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if task.Title != "" {
			existingTask.Title = task.Title
		}
		if task.Description != "" {
			existingTask.Description = task.Description
		}
		if task.Status != "" {
			existingTask.Status = task.Status
		}
		if !task.StartDate.IsZero() {
			existingTask.StartDate = task.StartDate
		}
		if !task.DueDate.IsZero() {
			existingTask.DueDate = task.DueDate
		}
		if task.AssignedTo != 0 {
			existingTask.AssignedTo = task.AssignedTo
		}

		_, err := db.DB.Exec("UPDATE tasks SET title = ?, description = ?, status = ?, start_date = ?, due_date = ?, assigned_to = ? WHERE id = ?",
			existingTask.Title, existingTask.Description, existingTask.Status, existingTask.StartDate, existingTask.DueDate, existingTask.AssignedTo, taskID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)

	})
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Delete a task by task ID
// @Tags tasks
// @Param task_id path int true "Task ID"
// @Success 200 {object} string
// @Failure 401 {object} string
// @Failure 500 {object} string
// @Router /tasks/{task_id} [delete]
func (db *AppHandler) DeleteTask() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		taskID := vars["task_id"]

		userRole := r.Context().Value("role").(string)
		if userRole != "admin" {
			http.Error(w, "Only admin can delete tasks", http.StatusInternalServerError)
			return
		}

		_, err := db.DB.Exec("DELETE FROM tasks WHERE id = ?", taskID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

// GetTasks godoc
// @Summary Get tasks for the user
// @Description Get all tasks assigned to the user or created by the admin
// @Tags tasks
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Task
// @Failure 401 {object} string
// @Failure 500 {object} string
// @Router /tasks [get]
func (db *AppHandler) GetTasks() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tasks []models.Task

		userIDValue := r.Context().Value("userID")
		userRoleValue := r.Context().Value("role")
		// userID veya role nil mi kontrol et
		if userIDValue == nil || userRoleValue == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Println("Unauthorized access: missing userID or role in context")
			return
		}

		userID, ok := userIDValue.(int)
		if !ok {
			http.Error(w, "Invalid userID", http.StatusInternalServerError)
			log.Println("Invalid userID in context")
			return
		}

		userRole, ok := userRoleValue.(string)
		if !ok {
			http.Error(w, "Invalid role", http.StatusInternalServerError)
			log.Println("Invalid role in context")
			return
		}

		log.Printf("User ID: %d, Role: %s", userID, userRole)

		if userRole == "admin" {
			rows, err := db.DB.Query("SELECT * FROM tasks WHERE user_id = ?", userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var task models.Task
				if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.StartDate, &task.DueDate, &task.UserID, &task.AssignedTo); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tasks = append(tasks, task)
			}

		} else if userRole == "user" {
			rows, err := db.DB.Query("SELECT * FROM tasks WHERE assigned_to = ?", userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var task models.Task
				if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.StartDate, &task.DueDate, &task.UserID, &task.AssignedTo); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tasks = append(tasks, task)
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
	})
}
