package v1

import (
	"effective_mobile_tz/internal/entity"
	"effective_mobile_tz/internal/service"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

type songRoutes struct {
	songService service.Song
}

func newSongRoutes(g *echo.Group, songService service.Song) {
	r := &songRoutes{
		songService: songService,
	}

	g.POST("", r.create)
	g.GET("", r.getSongsByFilter)
	g.GET("/:song_id", r.getByID)
	g.DELETE("/:song_id", r.delete)
	g.PUT("", r.updateSong)

	g.GET("/lyrics/:song_id", r.getPaginatedLyrics)
}

type songCreateInput struct {
	Group string `json:"group" validate:"required"`
	Title string `json:"title" validate:"required"`
}

// @Summary Creates a new song
// @Description This endpoint creates a new song by specifying the group and title.
// @Tags songs
// @Accept json
// @Produce json
// @Param input body songCreateInput true "Song creation input"
// @Success 200 {object} SuccessResponse "Song created successfully"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /songs [post]
func (r *songRoutes) create(c echo.Context) error {
	var input songCreateInput

	if err := c.Bind(&input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
	}

	if err := c.Validate(input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, err)
	}

	id, err := r.songService.CreateSong(c.Request().Context(), input.Group, input.Title)
	if err != nil {
		if errors.Is(err, service.ErrSongAlreadyExists) {
			return newErrorResponse(c, http.StatusBadRequest, err)
		}
		return newErrorResponse(c, http.StatusInternalServerError, err)
	}

	responseContent := struct {
		ID string
	}{
		ID: id,
	}

	return newSuccessResponse(c, "song created", responseContent)
}

// @Summary Get songs by filter
// @Description This endpoint retrieves songs from the library based on various filter criteria such as title, group, link, text, release date range, and pagination.
// @Tags songs
// @Accept json
// @Produce json
// @Param title query string false "Filter by title"
// @Param group query string false "Filter by group name"
// @Param link query string false "Filter by link"
// @Param text query string false "Filter by text (contains)"
// @Param startDate query string false "Filter by start date (YYYY-MM-DD)"
// @Param endDate query string false "Filter by end date (YYYY-MM-DD)"
// @Param page query int false "Page number for pagination (must be provided with limit)"
// @Param limit query int false "Limit of items per page (must be provided with page)"
// @Success 200 {object} SuccessResponse "List of songs retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid filter parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /songs [get]
func (r *songRoutes) getSongsByFilter(c echo.Context) error {
	params := c.QueryParams()
	title := params.Get("title")
	group := params.Get("group")
	link := params.Get("link")
	text := params.Get("text")
	startDateStr := params.Get("startDate")
	endDateStr := params.Get("endDate")
	page := params.Get("page")
	limit := params.Get("limit")

	if (startDateStr == "" && endDateStr != "") || (startDateStr != "" && endDateStr == "") {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("either both startDateStr and endDateStr should be provided, or neither of them"))
	}
	if startDateStr != "" && endDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid startDate"))
		}
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid endDate"))
		}
		if startDate.After(endDate) {
			return newErrorResponse(c, http.StatusBadRequest, errors.New("start_date cannot be after end_date"))
		}
	}

	limitInt, pageInt := 0, 0
	if (page == "" && limit != "") || (page != "" && limit == "") {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("either both page and limit should be provided, or neither of them"))
	} else if page != "" && limit != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid page number"))
		}
		limitInt, err := strconv.Atoi(limit)
		if err != nil || limitInt < 1 {
			return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid limit number"))
		}
	}

	filter := entity.SongFilter{
		Title:     title,
		Link:      link,
		Group:     group,
		Text:      text,
		StartDate: startDateStr,
		EndDate:   endDateStr,
		Limit:     limitInt,
		Offset:    pageInt,
	}

	songs, err := r.songService.GetSongsByFilter(c.Request().Context(), &filter)
	if err != nil {
		return newErrorResponse(c, http.StatusInternalServerError, err)

	}

	return newSuccessResponse(c, "songs retrieved", songs)
}

// @Summary Delete a song
// @Description This endpoint deletes a song by its ID.
// @Tags songs
// @Accept json
// @Produce json
// @Param song_id path string true "Song ID to delete"
// @Success 200 {object} SuccessResponse "Song deleted successfully"
// @Failure 400 {object} ErrorResponse "Bad request - song not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /songs/{song_id} [delete]
func (r *songRoutes) delete(c echo.Context) error {
	songID := c.Param("song_id")

	err := r.songService.DeleteSong(c.Request().Context(), songID)
	if err != nil {
		if errors.Is(err, service.ErrSongNotFound) {
			return newErrorResponse(c, http.StatusBadRequest, err)

		}
		return newErrorResponse(c, http.StatusInternalServerError, err)
	}

	return newSuccessResponse(c, "song deleted", nil)
}

// @Summary Get a song by ID
// @Description This endpoint retrieves a song's details by its ID.
// @Tags songs
// @Accept json
// @Produce json
// @Param song_id path string true "Song ID to retrieve"
// @Success 200 {object} SuccessResponse "Song retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - song not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /songs/{song_id} [get]
func (r *songRoutes) getByID(c echo.Context) error {
	songID := c.Param("song_id")

	song, err := r.songService.GetSongByID(c.Request().Context(), songID)
	if err != nil {
		if errors.Is(err, service.ErrSongNotFound) {
			return newErrorResponse(c, http.StatusBadRequest, err)
		}

		return newErrorResponse(c, http.StatusInternalServerError, err)
	}

	return newSuccessResponse(c, "song retrieved", song)
}

// @Summary Update a song
// @Description This endpoint updates a song's details. The song ID must be provided in the request body.
// @Tags songs
// @Accept json
// @Produce json
// @Param input body entity.SongUpdate true "Song update input"
// @Success 200 {object} SuccessResponse "Song updated successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid input or song not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /songs [put]
func (r *songRoutes) updateSong(c echo.Context) error {
	var input entity.SongUpdate
	if err := c.Bind(&input); err != nil {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid request body"))
	}

	if input.ID == "" {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("id not provided"))
	}

	err := r.songService.UpdateSong(c.Request().Context(), &input)
	if err != nil {
		if errors.Is(err, service.ErrSongNotFound) || errors.Is(err, service.ErrSongAlreadyExists) {
			return newErrorResponse(c, http.StatusBadRequest, err)
		}

		return newErrorResponse(c, http.StatusInternalServerError, err)
	}

	return newSuccessResponse(c, "song updated", nil)
}

// @Summary Get paginated lyrics
// @Description This endpoint retrieves paginated lyrics for a specific song by its ID.
// @Tags lyrics
// @Accept json
// @Produce json
// @Param song_id path string true "Song ID"
// @Param page query int true "Page number (must be provided with limit)"
// @Param limit query int true "Limit of items per page (must be provided with page)"
// @Success 200 {object} SuccessResponse "Lyrics retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid input parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /songs/lyrics/{song_id} [get]
func (r *songRoutes) getPaginatedLyrics(c echo.Context) error {
	songID := c.Param("song_id")
	page := c.QueryParams().Get("page")
	limit := c.QueryParams().Get("limit")

	if (page == "" && limit != "") || (page != "" && limit == "") {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("either both 'page' and 'limit' should be provided, or neither of them"))
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid page number"))

	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		return newErrorResponse(c, http.StatusBadRequest, errors.New("invalid limit number"))

	}

	lyrics, err := r.songService.GetPaginatedLyrics(c.Request().Context(), songID, pageInt, limitInt)
	if err != nil {
		return newErrorResponse(c, http.StatusInternalServerError, err)

	}

	type lyricsOutput struct {
		VerseNumber int
		Verse       string
	}
	var response []lyricsOutput
	for _, l := range lyrics {
		response = append(response, lyricsOutput{
			VerseNumber: l.VerseNumber,
			Verse:       l.Verse,
		})
	}

	return newSuccessResponse(c, "lyrics retrieved", response)
}
