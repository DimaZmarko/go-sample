package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "go-sample/internal/models"
    "go-sample/internal/repository"

    "github.com/gorilla/mux"
)

type UserHandler struct {
    userRepo repository.UserRepository
}

type UpdateUserRequest struct {
    Email   string `json:"email"`
    Name    string `json:"name"`
    TeamIDs []uint `json:"team_ids"`
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
    return &UserHandler{
        userRepo: userRepo,
    }
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if err := h.userRepo.Create(&user); err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusCreated, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    user, err := h.userRepo.GetByID(uint(id))
    if err != nil {
        ErrorResponse(w, http.StatusNotFound, "User not found")
        return
    }

    user.Email = req.Email
    user.Name = req.Name

    if err := h.userRepo.UpdateWithTeams(user, req.TeamIDs); err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    if err := h.userRepo.Delete(uint(id)); err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    user, err := h.userRepo.GetWithTeams(uint(id))
    if err != nil {
        ErrorResponse(w, http.StatusNotFound, "User not found")
        return
    }

    SuccessResponse(w, http.StatusOK, user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    users, err := h.userRepo.List()
    if err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusOK, users)
} 