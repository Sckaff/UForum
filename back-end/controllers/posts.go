package controllers

import (
	"net/http"

	"cen/backend/models"
	"cen/backend/utils/token"

	"strings"

	"github.com/gin-gonic/gin"
)

type CreatePostInput struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Category string `json:"category" binding:"required"`
}

type UpdatePostInput struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Category string `json:"category" binding:"required"`
}

func GetPosts(c *gin.Context) {
	var posts []models.Post
	models.DB.Find(&posts)

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// REQUIRES LOGIN
func GetUserPosts(c *gin.Context) {
	user_id, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := models.GetUserByID(user_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var username = u.Username

	var posts []models.Post
	models.DB.Where("User = ?", username).Find(&posts)

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// REQUIRES LOGIN
func CreatePost(c *gin.Context) {
	user_id, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := models.GetUserByID(user_id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var username = u.Username

	var post_input CreatePostInput
	if err := c.ShouldBindJSON(&post_input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.DB.Find(&models.Category{}, "title = ?", post_input.Category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found!"})
		return
	}

	post := models.Post{Title: post_input.Title, Body: post_input.Body, Category: post_input.Category, User: username, Likes: "", Dislikes: "", NetRating: 0}
	models.DB.Create(&post)

	c.JSON(http.StatusOK, gin.H{"data": post})
}

func GetPostByID(c *gin.Context) {
	var post models.Post

	if err := models.DB.Where("id = ?", c.Param("id")).First(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}

// REQUIRES LOGIN
func PatchPost(c *gin.Context) {
	user_id, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := models.GetUserByID(user_id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var username = u.Username

	var old_post models.Post
	if err := models.DB.Where("id = ?", c.Param("id")).First(&old_post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found!"})
		return
	}

	var post_input UpdatePostInput
	if err := c.ShouldBindJSON(&post_input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.DB.Find(&models.Category{}, "title = ?", post_input.Category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found!"})
		return
	}

	if old_post.User != username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not authorized to edit this post!"})
	} else {
		new_post := models.Post{Title: post_input.Title, Body: post_input.Body, Category: post_input.Category, User: username, Likes: "", Dislikes: ""}
		models.DB.Model(&old_post).Updates(new_post)

		c.JSON(http.StatusOK, gin.H{"data": new_post})
	}

}

// REQUIRES LOGIN
func DeletePost(c *gin.Context) {
	user_id, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := models.GetUserByID(user_id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var username = u.Username

	var post models.Post
	if err := models.DB.Where("id = ?", c.Param("id")).First(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found!"})
		return
	}

	// Check if the post belongs to the user, delete if it does
	if post.User != username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not authorized to delete this post!"})
		return
	} else {
		models.DB.Delete(&post)
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

func GetPostsByCategory(c *gin.Context) {
	categoryTitle := c.Param("category")

	var posts []models.Post

	if err := models.DB.Where("category = ?", categoryTitle).Find(&posts).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// REQUIRES LOGIN
func LikePost(c *gin.Context) {
	userID, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetUserByID(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username

	var post models.Post
	if err := models.DB.Where("id = ?", c.Param("id")).First(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found!"})
		return
	}

	var originalposter models.User
	models.DB.Where("username = ?", post.User).First(&originalposter)

	// checks if the user has liked the post
	if strings.Contains(post.Likes, username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked this post!"})
		return
	}

	// checks if the user has disliked the post
	if strings.Contains(post.Dislikes, username) {
		// removes user from dislikes string
		post.Dislikes = strings.ReplaceAll(post.Dislikes, username+",", "")
		post.Likes += username + ","
		post.NetRating += 2
		originalposter.UpvoteCount += 2
		//post.User.UpvoteCount += 2
	} else {
		// otherwise user has not rated the post yet
		post.Likes += username + ","
		post.NetRating += 1
		originalposter.UpvoteCount += 1
	}

	if err := models.DB.Save(&originalposter).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update upvote count of orginal poster!"})
		return
	}

	if err := models.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update like array!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}

// REQUIRES LOGIN
func DislikePost(c *gin.Context) {
	userID, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetUserByID(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username

	var post models.Post
	if err := models.DB.Where("id = ?", c.Param("id")).First(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found!"})
		return
	}

	var originalposter models.User
	models.DB.Where("username = ?", post.User).First(&originalposter)

	// checks if the user has disliked the post
	if strings.Contains(post.Dislikes, username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already disliked this post!"})
		return
	}

	// checks if the user has liked the post
	if strings.Contains(post.Likes, username) {
		// removes user from dislikes string
		post.Likes = strings.ReplaceAll(post.Likes, username+",", "")
		post.Dislikes += username + ","
		post.NetRating -= 2
		originalposter.UpvoteCount -= 2
	} else {
		// otherwise user has not rated the post yet
		post.Dislikes += username + ","
		post.NetRating -= 1
		originalposter.UpvoteCount -= 1
	}

	if err := models.DB.Save(&originalposter).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update upvote count of orginal poster!"})
		return
	}

	if err := models.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update like array!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}

// REQUIRES LOGIN
func ClearPostLikes(c *gin.Context) {
	userID, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetUserByID(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username

	var post models.Post
	if err := models.DB.Where("id = ?", c.Param("id")).First(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found!"})
		return
	}

	var originalposter models.User
	models.DB.Where("username = ?", post.User).First(&originalposter)

	// checks if the user has liked the post
	if strings.Contains(post.Likes, username) {
		// removes user from likes string
		post.Likes = strings.ReplaceAll(post.Likes, username+",", "")
		post.NetRating -= 1
		originalposter.UpvoteCount -= 1
	} else if strings.Contains(post.Dislikes, username) { // checks if the user has disliked the post
		// removes user from dislikes string
		post.Dislikes = strings.ReplaceAll(post.Dislikes, username+",", "")
		post.NetRating += 1
		originalposter.UpvoteCount += 1
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have not rated this post!"})
		return
	}

	if err := models.DB.Save(&originalposter).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update upvote count of orginal poster!"})
		return
	}

	if err := models.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update like/dislike array!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": post})
}
