package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"user-service/internal/models"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	UserService *services.UserService
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.UserService.CreateUserService(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.GetHeader("X-User-ID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user, err := h.UserService.GetUserService(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.GetHeader("X-User-ID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.ID = id

	if err := h.UserService.UpdateUserService(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.GetHeader("X-User-ID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.UserService.DeleteUserService(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) AddAddress(c *gin.Context) {
	idStr := c.GetHeader("X-User-ID")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var address models.UserAddress
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	address.UserID = userID

	if err := h.UserService.AddAddressService(&address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, address)
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	var address models.UserAddress
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.UserService.UpdateAddressService(&address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, address)
}

func (h *UserHandler) GetUserAddresses(c *gin.Context) {
	idStr := c.GetHeader("X-User-ID")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit, offset := 10, 0 
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	addresses, err := h.UserService.GetUserAddressesService(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	addressIDStr := c.Param("address_id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}

	if err := h.UserService.DeleteAddressService(addressID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "Address deleted successfully")
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	userType := c.GetHeader("X-User-Type")
    if userType == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User type is nil"})
        return
    }

    page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
    if err != nil || page < 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page"})
        return
    }

    limit, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
    if err != nil || limit < 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
        return
    }

	status := c.Query("status")

    offset := (page - 1) * limit  // Tính offset từ page

    users, err := h.UserService.ListUsersService(int(limit), int(offset), userType, status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userType := c.GetHeader("X-User-Type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User type is nil"})
		return
	}
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.UserService.GetUserByIDService(userID, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}	

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userType := c.GetHeader("X-User-Type")
	if userType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User type is nil"})
		return
	}
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.UserService.UpdateUserStatusService(userID, userType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "User status updated successfully")
}

func (h *UserHandler) AdminDeleteUser(c *gin.Context) {
	adminType := c.GetHeader("X-User-Type")
	if adminType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin type is nil"})
		return
	}
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.UserService.AdminDeleteUserService(userID, adminType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "User deleted successfully by admin")
}

func (h *UserHandler) GetUserStatistics(c *gin.Context) {
	month, _ := strconv.Atoi(c.Query("month"))
    year, _ := strconv.Atoi(c.Query("year"))

    log.Printf("Received month: %d, year: %d", month, year)

    userType := c.GetHeader("X-User-Type")
    if userType == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User type is nil"})
        return
    }
	if userType != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

    if month == 0 || year == 0 {
        now := time.Now()
        if month == 0 {
            month = int(now.Month())
        }
        if year == 0 {
            year = now.Year()
        }
    }

    curUsers, growthPercent, err := h.UserService.GetUserStatisticsService(month, year)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "current_month_users":  curUsers,
        "growth_percentage":    growthPercent,
    })
}