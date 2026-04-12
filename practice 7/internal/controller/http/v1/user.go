package v1

import (
	"net/http"

	"practice7/internal/controller/http/middleware"
	"practice7/internal/entity"
	"practice7/internal/usecase"
	"practice7/internal/utils"

	"github.com/gin-gonic/gin"
)

type userRoutes struct {
	t usecase.UserInterface
}

func NewUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, jwtManager *utils.JWTManager) {
	r := &userRoutes{t: t}

	h := handler.Group("/users")
	{
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)

		protected := h.Group("")
		protected.Use(middleware.JWTAuthMiddleware(jwtManager))
		{
			protected.GET("/me", r.GetMe)
			protected.GET("/protected/hello", r.ProtectedFunc)
			protected.PATCH("/promote/:id", middleware.RoleMiddleware(entity.RoleAdmin), r.PromoteUser)
		}
	}
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
	var input entity.CreateUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, sessionID, err := r.t.RegisterUser(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "user registered successfully",
		"session_id": sessionID,
		"user":       user,
	})
}

func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (r *userRoutes) GetMe(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	user, err := r.t.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (r *userRoutes) PromoteUser(c *gin.Context) {
	user, err := r.t.PromoteUser(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user promoted to admin",
		"user":    user,
	})
}

func (r *userRoutes) ProtectedFunc(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
