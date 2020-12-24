package services

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"strconv"

	"github.com/komfy/cloudinary"
)

// SendMail used on register or resend validation to reciever's email address
func SendMail(reciever string, id uint) error {
	url := os.Getenv("LINK")
	// Sender data.
	from := "servicegolang087@gmail.com"
	password := "1234566777oo"

	// Receiver email address.
	to := []string{
		// "sender@example.com",
		reciever,
	}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.
	body := "plaese validate your email http://" + url + "/api/auth/" + strconv.FormatUint(uint64(id), 10)
	msg := "From: " + from + "\n" +
		"To: " + to[0] + "\n" +
		"Subject: Hello there\n\n" +
		body
	message := []byte(msg)
	// fmt.Println("plaese validate your email" + url + "/api/auth/" + string(id))

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		// fmt.Println(err)
		return err
	}
	fmt.Println("Email Sent Successfully!")
	return nil
}

// cloudinary service
type downloadHandler struct {
	cs *cloudinary.Service
}

//Creating a handler to donwload from form.
//I prefer to use handlers, because it's easier to add some external services into it's logic.
func (h downloadHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Method must be POST", http.StatusMethodNotAllowed)
		return
	}
	// Parsing multipart form is not necessary, FormFile invokes it if form isn't parsed.
	file, fh, err := req.FormFile("file")
	upResp, err := h.cs.Upload(fh.Filename, file, false)
	if err != nil {
		http.Error(res, err.Error(), 505)
		return
	}
	url := upResp.SecureURL
	res.Write([]byte(url))
}

//CloudinaryService global cloudinary service
var CloudinaryService *cloudinary.Service

//InitCloudinary func
func InitCloudinary() {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		log.Fatalln("there is no env variable with given name")
	}
	s, err := cloudinary.NewService(cloudinaryURL)
	if err != nil {
		panic(err)
	}
	CloudinaryService = s
}

//UploadFile func
func UploadFile(file *multipart.FileHeader) string {
	url := ""
	fmt.Println(file.Filename)
	data, err := file.Open()
	if err == nil {
		upResp, err := CloudinaryService.Upload(file.Filename, data, false)
		if err != nil {
			panic(err)
		} else {
			url = upResp.URL
		}

	}
	return url
}
