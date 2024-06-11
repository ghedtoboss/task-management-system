package handlers

import (
	"encoding/json"
	"net/http"
	"task-management-system/models"
)

// CreateFriendship godoc
// @Summary Create a new friendship request
// @Description Create a new friendship request and set status to pending
// @Tags friendship
// @Accept  json
// @Produce  json
// @Param friendship body models.Friendship true "Friendship info"
// @Success 201 {object} models.Friendship
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /friends [post]
func (db *AppHandler) CreateFriendship() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var friendship models.Friendship
		if err := json.NewDecoder(r.Body).Decode(&friendship); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("userID").(int)
		friendship.UserID = userID
		friendship.Status = "pending"

		_, err := db.DB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES (?, ?, ?)", friendship.UserID, friendship.FriendID, friendship.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(friendship)
	})
}

// AcceptFriendRequest godoc
// @Summary Accept a friendship request
// @Description Accept a friendship request by updating status to accepted
// @Tags friendship
// @Accept  json
// @Produce  json
// @Param friendship body models.Friendship true "Friendship info"
// @Success 200 {object} models.Friendship
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /friends/accept [post]
func (db *AppHandler) AcceptFriendRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var friendship models.Friendship
		if err := json.NewDecoder(r.Body).Decode(&friendship); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("userID").(int)
		_, err := db.DB.Exec("UPDATE friendships SET status = 'accepted' WHERE friend_id = ? AND user_id = ?", userID, friendship.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		friendship.Status = "accepted"
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(friendship)
	})
}

// RejectFriendRequest godoc
// @Summary Reject a friendship request
// @Description Reject a friendship request by updating status to rejected
// @Tags friendship
// @Accept  json
// @Produce  json
// @Param friendship body models.Friendship true "Friendship info"
// @Success 200 {object} models.Friendship
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /friends/reject [post]
func (db *AppHandler) RejectFriendRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var friendship models.Friendship
		if err := json.NewDecoder(r.Body).Decode(&friendship); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("userID").(int)
		_, err := db.DB.Exec("UPDATE friendships SET status = 'rejected' WHERE friend_id = ? AND user_id = ?", userID, friendship.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		friendship.Status = "rejected"
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(friendship)
	})
}
