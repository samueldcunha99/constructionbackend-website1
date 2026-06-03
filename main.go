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

func sendEmail(to string) error {
	const gmailUser = "Samueldcunha99@gmail.com"
	const gmailAppPassword = "rlxk lmnn bicg ccou"

	m := gomail.NewMessage()

	m.SetHeader("From", gmailUser)
	m.SetHeader("To", to)
	m.SetHeader("Cc", gmailUser)

	m.SetHeader(
		"Subject",
		"Samuel Construction - Inquiry Received",
	)

	m.SetBody(
		"text/html",
		`
		<h2>Thank You For Contacting Samuel Construction</h2>
		<p>We have received your inquiry successfully.</p>
		<p>Our team will contact you shortly.</p>
		<br>
		<p>Regards,</p>
		<p><b>Samuel Construction</b></p>
		`,
	)

	d := gomail.NewDialer(
		"smtp.gmail.com",
		587,
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

		err = sendEmail(contact.Email)

		if err != nil {
			fmt.Println("EMAIL ERROR:", err.Error())
		}

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
