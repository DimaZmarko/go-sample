package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "go-sample/internal/models"
    "go-sample/internal/repository"

    "github.com/gorilla/mux"
)

type TeamHandler struct {
    teamRepo repository.TeamRepository
}

func NewTeamHandler(teamRepo repository.TeamRepository) *TeamHandler {
    return &TeamHandler{
        teamRepo: teamRepo,
    }
}

func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
    var team models.Team
    if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if err := h.teamRepo.Create(&team); err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusCreated, team)
}

func (h *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid team ID")
        return
    }

    var team models.Team
    if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    team.ID = uint(id)

    if err := h.teamRepo.Update(&team); err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusOK, team)
}

func (h *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid team ID")
        return
    }

    if err := h.teamRepo.Delete(uint(id)); err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusOK, map[string]string{"message": "Team deleted successfully"})
}

func (h *TeamHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        ErrorResponse(w, http.StatusBadRequest, "Invalid team ID")
        return
    }

    team, err := h.teamRepo.GetByID(uint(id))
    if err != nil {
        ErrorResponse(w, http.StatusNotFound, "Team not found")
        return
    }

    SuccessResponse(w, http.StatusOK, team)
}

func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
    teams, err := h.teamRepo.List()
    if err != nil {
        ErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    SuccessResponse(w, http.StatusOK, teams)
}

func (h *TeamHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Parse request body
	var request struct {
		UserID uint `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get team
	team, err := h.teamRepo.GetByID(uint(teamID))
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "Team not found")
		return
	}

	// Add user to team
	if err := h.teamRepo.AddUser(team.ID, request.UserID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, http.StatusOK, map[string]string{"message": "User added to team successfully"})
} 