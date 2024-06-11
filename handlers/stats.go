package handlers

import (
	"encoding/json"
	"net/http"
	"task-management-system/models"
)

// GetStats godoc
// @Summary Get user stats
// @Description Get statistics of tasks assigned to the user
// @Tags stats
// @Accept  json
// @Produce  json
// @Success 200 {object} models.UserStats
// @Failure 401 {object} string
// @Failure 500 {object} string
// @Router /user/stats [get]
func (db *AppHandler) GetStats() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDValue := r.Context().Value("userID")
		if userIDValue == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := userIDValue.(int)

		var stats models.UserStats

		//toplam görev sayısı
		err := db.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE assigned_to = ?", userID).Scan(&stats.TotalTasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//tamamlanan görev sayısı
		err = db.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE assigned_to = ? AND status = 'completed'", userID).Scan(&stats.CompletedTasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Bekleyen görev sayısı
		err = db.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE assigned_to = ? AND status = 'pending'", userID).Scan(&stats.PendingTasks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	})
}
