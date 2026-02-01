package main

import (
	"log"
	"os"

	authkit "github.com/Prescott-Data/dromos-authkit"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Configure authentication
	cfg := authkit.Config{
		IssuerURL: getEnv("ZITADEL_ISSUER_URL", "http://localhost:8080"),
		Audience:  getEnv("ZITADEL_AUDIENCE", "your-project-id"),
		SkipPaths: []string{"/health", "/public"},
	}

	// Apply authentication middleware globally
	r.Use(authkit.AuthN(cfg))

	// Public routes (skipped by middleware)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "This is a public endpoint"})
	})

	// Protected routes (require authentication)
	api := r.Group("/api")
	{
		// Get current user info
		api.GET("/me", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"user_id": authkit.UserID(c),
				"email":   authkit.Email(c),
				"org_id":  authkit.OrgID(c),
			})
		})

		// Get full claims
		api.GET("/claims", func(c *gin.Context) {
			claims := authkit.GetClaims(c)
			c.JSON(200, claims)
		})

		// Admin-only routes
		admin := api.Group("/admin")
		admin.Use(authkit.RequireRole("admin"))
		{
			admin.GET("/users", listUsers)
			admin.POST("/users", createUser)
			admin.DELETE("/users/:id", deleteUser)
		}

		// Editor routes (admin or editor)
		content := api.Group("/content")
		content.Use(authkit.RequireRole("editor", "admin"))
		{
			content.GET("/articles", listArticles)
			content.POST("/articles", createArticle)
			content.PUT("/articles/:id", updateArticle)
		}

		// Custom role check
		api.POST("/publish", func(c *gin.Context) {
			if !authkit.HasAnyRole(c, "publisher", "admin") {
				c.JSON(403, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(200, gin.H{"status": "published"})
		})
	}

	port := getEnv("PORT", "3000")
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// Handler implementations
func listUsers(c *gin.Context) {
	c.JSON(200, gin.H{"users": []string{"user1", "user2"}})
}

func createUser(c *gin.Context) {
	c.JSON(201, gin.H{"message": "user created"})
}

func deleteUser(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(200, gin.H{"message": "user deleted", "id": userID})
}

func listArticles(c *gin.Context) {
	c.JSON(200, gin.H{"articles": []string{"article1", "article2"}})
}

func createArticle(c *gin.Context) {
	c.JSON(201, gin.H{"message": "article created"})
}

func updateArticle(c *gin.Context) {
	articleID := c.Param("id")
	c.JSON(200, gin.H{"message": "article updated", "id": articleID})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
