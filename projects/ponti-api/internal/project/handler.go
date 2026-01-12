package project

import (
	"net/http"
	"strconv"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	dto "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/handler/dto"
	"github.com/gin-gonic/gin"
)

// Handler encapsulates all dependencies for the Project HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Project handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{ucs: u, gsv: s, mws: m}
}

// Routes registers all project routes.
func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	apiV := h.gsv.GetApiVersion()
	base := "/api/" + apiV + "/projects"

	public := r.Group(base + "/public")
	{
		public.POST("", h.CreateProject)                        // Create a project
		public.GET("", h.ListProjects)                          // List all projects
		public.GET("/customer/:id", h.ListProjectsByCustomerID) // List projects by customer ID
		public.GET("/:id", h.GetProject)                        // Get a project by ID
		public.PUT("/:id", h.UpdateProject)                     // Update a project
		public.DELETE("/:id", h.DeleteProject)                  // Delete a project
	}
}

// CreateProject handles project creation.
func (h *Handler) CreateProject(c *gin.Context) {
	var req dto.CreateProject
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	pID, err := h.ucs.CreateProject(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateProjectResponse{Message: "created", ProjectID: pID})
}

// ListProjectsByCustomerID returns projects filtered by customer ID (path param).
func (h *Handler) ListProjectsByCustomerID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "missing customer id in path"})
		return
	}
	customerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid customer id"})
		return
	}

	projects, err := h.ucs.ListProjectsByCustomerID(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	out := make([]dto.Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, *dto.FromDomain(&p))
	}
	c.JSON(http.StatusOK, out)
}

// ListProjects returns all projects.
func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.ucs.ListProjects(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	out := make([]dto.Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, *dto.FromDomain(&p))
	}
	c.JSON(http.StatusOK, out)
}

// GetProject returns a single project by ID.
func (h *Handler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	proj, err := h.ucs.GetProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDomain(proj))
}

// UpdateProject handles a full project update.
func (h *Handler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	var req dto.UpdateProject
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
		return
	}
	dom := req.ToDomain()
	dom.ID = id
	if err := h.ucs.UpdateProject(c.Request.Context(), dom); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "updated"})
}

// DeleteProject removes a project by ID.
func (h *Handler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid project id"})
		return
	}
	if err := h.ucs.DeleteProject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "deleted"})
}
