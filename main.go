package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gomail.v2"
)

type Contact struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

var db *sql.DB

func connectDB() {
	var err error

	db, err = sql.Open(
		"mysql",
		"root:gFSIguKgTkEgLiFfoKchkPyLiMwVjQjW@tcp(acela.proxy.rlwy.net:41408)/railway",
	)

	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Database Connected Successfully")
}

func sendEmail(contact Contact) error {
	const gmailUser = "Samueldcunha99@gmail.com"
	const gmailAppPassword = "rlxk lmnn bicg ccou"

	m := gomail.NewMessage()

	m.SetHeader("From", gmailUser)
	m.SetHeader("To", gmailUser)
	m.SetHeader("Reply-To", contact.Email)

	m.SetHeader(
		"Subject",
		"New Inquiry - NMS ENTERPRISES",
	)

	body := fmt.Sprintf(`
		<h2>New Inquiry Received</h2>
		<p><b>Name:</b> %s</p>
		<p><b>Email:</b> %s</p>
		<p><b>Phone:</b> %s</p>
		<p><b>Message:</b> %s</p>
		<br>
		<p>This inquiry was submitted from your construction website.</p>
	`, contact.Name, contact.Email, contact.Phone, contact.Message)

	plainBody := fmt.Sprintf("New Inquiry Received\n\nName: %s\nEmail: %s\nPhone: %s\nMessage: %s\n", contact.Name, contact.Email, contact.Phone, contact.Message)

	m.SetBody("text/plain", plainBody)
	m.AddAlternative("text/html", body)

	d := gomail.NewDialer(
		"smtp.gmail.com",
		465,
		gmailUser,
		gmailAppPassword,
	)

	return d.DialAndSend(m)
}

func main() {
	connectDB()
	defer db.Close()

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Backend Running",
		})
	})

	router.POST("/contact", func(c *gin.Context) {
		var contact Contact

		if err := c.BindJSON(&contact); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Data",
			})
			return
		}

		emailRegex := regexp.MustCompile(
			`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
		)

		if !emailRegex.MatchString(contact.Email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Email Address",
			})
			return
		}

		phoneRegex := regexp.MustCompile(`^[0-9]{10}$`)

		if !phoneRegex.MatchString(contact.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Phone must contain exactly 10 digits",
			})
			return
		}

		_, err := db.Exec(
			"INSERT INTO contacts (name, email, phone, message) VALUES (?, ?, ?, ?)",
			contact.Name,
			contact.Email,
			contact.Phone,
			contact.Message,
		)

		if err != nil {
			fmt.Println("DATABASE ERROR:", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save data",
			})
			return
		}
		contactCopy := contact
		go func() {
			fmt.Println("EMAIL SEND STARTED:", contactCopy.Email)
			if err := sendEmail(contactCopy); err != nil {
				fmt.Println("EMAIL ERROR:", err.Error())
				return
			}
			fmt.Println("EMAIL SENT SUCCESSFULLY:", contactCopy.Email)
		}()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Form Submitted Successfully",
		})
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	router.Run("0.0.0.0:" + port)
}
