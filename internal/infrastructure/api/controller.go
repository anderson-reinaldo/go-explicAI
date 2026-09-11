package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

type ExplicaServer struct {
}

func NewExplicaServer() *ExplicaServer {
	return &ExplicaServer{}
}

func (api *ExplicaServer) Register(server *echo.Echo) {
	server.POST("/upload", api.Upload)
	server.GET("/summaries", api.ListSummaries)
	server.GET("/summaries/:externalId", api.GetSummaryByExternalID)
	server.DELETE("/summaries/:externalId", api.DeleteSummaryByExternalID)
}

func (api *ExplicaServer) Upload(c echo.Context) error {
	ctx := c.Request().Context()
	_, err := api.getFileFromRequest(ctx, c)
	if err != nil {
		return echo.ErrBadRequest
	}

	// TODO: init flow
	return c.JSON(http.StatusCreated, nil)
}

func (api *ExplicaServer) ListSummaries(c echo.Context) error {
	//ctx := c.Request().Context()

	//init get flow

	return c.JSON(http.StatusOK, nil)
}

func (api *ExplicaServer) GetSummaryByExternalID(c echo.Context) error {
	//ctx := c.Request().Context()
	//externalId := c.Param("externalId")

	//parsedExtenalID, err := uuid.Parse(externalId)
	//if err != nil {
	//	return echo.ErrBadRequest
	//}

	return c.JSON(http.StatusOK, nil)
}

func (api *ExplicaServer) DeleteSummaryByExternalID(c echo.Context) error {
	//ctx := c.Request().Context()
	//externalId := c.Param("externalId")

	//parsedExtenalID, err := uuid.Parse(externalId)
	//if err != nil {
	//	return echo.ErrBadRequest
	//}

	//TODO: delete flow

	return c.JSON(http.StatusOK, map[string]string{
		"message": "O resumo foi removido",
	})
}

func (api *ExplicaServer) getFileFromRequest(ctx context.Context, c echo.Context) ([]byte, error) {
	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println("missing file")
		return nil, errors.New("missing file")
	}

	allowedExtesions := map[string]bool{
		".mp3":  true,
		".mp4":  true,
		".mpeg": true,
		".mpga": true,
		".m4a":  true,
		".wav":  true,
		".webm": true,
	}

	fileExtension := strings.ToLower(filepath.Ext(file.Filename))

	if !allowedExtesions[fileExtension] {
		return nil, errors.New("invalid file")
	}

	src, err := file.Open()
	if err != nil {
		return nil, errors.New("fail to open file")
	}
	defer src.Close()

	var buf bytes.Buffer

	_, err = io.Copy(&buf, src)
	if err != nil {
		fmt.Println("fail to read file")
	}

	return buf.Bytes(), nil
}
