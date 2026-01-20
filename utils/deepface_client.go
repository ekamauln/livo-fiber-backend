package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type RegisterResult struct {
	Status string `json:"status"`
	UserID string `json:"userId"`
}

type VerifyResult struct {
	Matched    bool    `json:"matched"`
	UserID     string  `json:"userId"`
	Confidence float64 `json:"confidence"`
}

func SendToDeepFaceRegister(userID uint, imagePath string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// user_id field
	_ = writer.WriteField("user_id", strconv.Itoa(int(userID)))

	// image file
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="image"; filename="%s"`, filepath.Base(imagePath)))
	h.Set("Content-Type", "image/jpeg")

	part, err := writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	writer.Close()

	req, err := http.NewRequest("POST", os.Getenv("DEEPFACE_URL")+"/register", body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-SERVICE-KEY", os.Getenv("DEEPFACE_SERVICE_KEY"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(fmt.Errorf("DeepFace register failed: %s", string(data)).Error())
	}

	return nil
}

func SendToDeepFaceVerify(userID uint, imagePath string) (*VerifyResult, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// user_id
	_ = writer.WriteField("user_id", strconv.Itoa(int(userID)))

	// image
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	h.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="image"; filename="%s"`, filepath.Base(imagePath)),
	)
	h.Set("Content-Type", "image/jpeg")

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	writer.Close()

	req, err := http.NewRequest(
		"POST",
		os.Getenv("DEEPFACE_URL")+"/verify",
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-SERVICE-KEY", os.Getenv("DEEPFACE_SERVICE_KEY"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepface verify failed: %s", string(data))
	}

	// Log the response for debugging
	fmt.Printf("DeepFace Response: %s\n", string(data))

	var result VerifyResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse deepface response: %v, raw response: %s", err, string(data))
	}

	return &result, nil
}
