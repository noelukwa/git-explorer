package handlers_test

// func TestAddIntent(t *testing.T) {
// 	repo := inmem.NewIntentRepository()
// 	intentService := service.NewIntentService(repo)
// 	handler := handlers.NewIntentHandler(intentService)

// 	e := echo.New()

// 	tests := []struct {
// 		name           string
// 		requestBody    string
// 		expectedStatus int
// 		expectedError  string
// 	}{
// 		{
// 			name:           "Valid request",
// 			requestBody:    `{"repo": "example/repo", "since": "2023-01-01T00:00:00Z"}`,
// 			expectedStatus: http.StatusCreated,
// 		},
// 		{
// 			name:           "Invalid request - missing repo",
// 			requestBody:    `{"since": "2023-01-01T00:00:00Z"}`,
// 			expectedStatus: http.StatusBadRequest,
// 			expectedError:  "error",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := httptest.NewRequest(http.MethodPost, "/intents", strings.NewReader(tt.requestBody))
// 			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 			rec := httptest.NewRecorder()
// 			c := e.NewContext(req, rec)

// 			err := handler.AddIntent(c)

// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedStatus, rec.Code)

// 			if tt.expectedStatus == http.StatusCreated {
// 				var response models.Intent
// 				err := json.Unmarshal(rec.Body.Bytes(), &response)
// 				assert.NoError(t, err)
// 				assert.NotEmpty(t, response.ID)
// 			} else {
// 				var errorResponse map[string]string
// 				err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
// 				assert.NoError(t, err)
// 				assert.Contains(t, errorResponse, tt.expectedError)
// 			}
// 		})
// 	}
// }
