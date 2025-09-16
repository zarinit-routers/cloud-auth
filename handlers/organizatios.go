package handlers

import (
	"auth-service/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const (
	URIGetOrganizations = "/api/organizations/get-user-organization"
)

func genIPCToken() (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Minute * 2).Unix(),
	}

	secretKey := []byte(viper.GetString("jwt-security-key"))

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(secretKey)

	if err != nil {
		log.Error("Failed to generate IPC token", "error", err)
		return "", err
	}

	return token, nil
}

type OrganizationRequest struct {
	UserID uuid.UUID `json:"id"`
}

func getOrganizationsURL() string {
	url := &url.URL{
		Host:   fmt.Sprintf("%s:%d", viper.GetString("organizations.host"), viper.GetInt("organizations.port")),
		Path:   URIGetOrganizations,
		Scheme: viper.GetString("organizations.scheme"),
	}
	return url.String()
}
func getOrganizationId(user models.User) (uuid.UUID, error) {
	token, err := genIPCToken()
	if err != nil {
		log.Error("Failed to generate IPC token", "error", err)
		return uuid.Nil, err
	}
	requestBody := OrganizationRequest{
		UserID: user.ID,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Error("Failed to encode request body", "error", err)
		return uuid.Nil, fmt.Errorf("failed marshall request body: %s", err)
	}
	url := getOrganizationsURL()
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		log.Error("Failed to create request", "error", err)
		return uuid.Nil, fmt.Errorf("failed create request: %s", err)
	}
	request.Header.Add("Authorization", token)

	request.Header.Add("Content-Type", "application/json")
	if _, err := request.Body.Read(bodyJSON); err != nil {
		return uuid.Nil, fmt.Errorf("failed read request body")
	}

	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		log.Error("Failed to get organization ID", "error", err)
		return uuid.Nil, err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Error("Failed to close response body", "error", err)
		}
	}()

	var resp struct {
		OrganizationID uuid.UUID `json:"organizationId"`
	}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		log.Error("Failed to decode organization ID", "error", err)
		return uuid.Nil, err
	}

	return resp.OrganizationID, nil

}
