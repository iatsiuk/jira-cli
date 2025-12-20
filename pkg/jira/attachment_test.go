package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetIssueWithAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/TEST-1", r.URL.Path)

		resp, err := os.ReadFile("./testdata/issue-with-attachments.json")
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(resp)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetIssueV2("TEST-1")
	assert.NoError(t, err)
	assert.NotNil(t, actual)

	assert.Len(t, actual.Fields.Attachments, 2)

	att1 := actual.Fields.Attachments[0]
	assert.Equal(t, "10001", att1.ID)
	assert.Equal(t, "screenshot.png", att1.Filename)
	assert.Equal(t, "John Doe", att1.Author.DisplayName)
	assert.Equal(t, "2020-12-03T14:10:00.000+0100", att1.Created)
	assert.Equal(t, int64(12345), att1.Size)
	assert.Equal(t, "image/png", att1.MimeType)
	assert.Equal(t, "https://example.atlassian.net/rest/api/3/attachment/content/10001", att1.Content)

	att2 := actual.Fields.Attachments[1]
	assert.Equal(t, "10002", att2.ID)
	assert.Equal(t, "document.pdf", att2.Filename)
}

func TestGetAttachmentContent(t *testing.T) {
	expectedContent := []byte("binary file content here")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/attachment/content/10001", r.URL.Path)
		assert.Equal(t, "redirect=false", r.URL.RawQuery)
		assert.Equal(t, "*/*", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	reader, err := client.GetAttachmentContent("10001")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	defer func() { _ = reader.Close() }()

	content, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestGetAttachmentContentNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Attachment with id 99999 does not exist"]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	reader, err := client.GetAttachmentContent("99999")
	assert.Error(t, err)
	assert.Nil(t, reader)

	var unexpectedErr *ErrUnexpectedResponse
	assert.ErrorAs(t, err, &unexpectedErr)
	assert.Equal(t, http.StatusNotFound, unexpectedErr.StatusCode)
}

func TestGetAttachmentContentUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errorMessages":["Internal server error"]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	reader, err := client.GetAttachmentContent("10001")
	assert.Error(t, err)
	assert.Nil(t, reader)

	var unexpectedErr *ErrUnexpectedResponse
	assert.ErrorAs(t, err, &unexpectedErr)
	assert.Equal(t, http.StatusInternalServerError, unexpectedErr.StatusCode)
}

func TestGetAttachmentContentV2(t *testing.T) {
	expectedContent := []byte("binary file content v2")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/attachment/content/10001", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)
		assert.Equal(t, "*/*", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	reader, err := client.GetAttachmentContentV2("10001")
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	defer func() { _ = reader.Close() }()

	content, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestGetAttachmentContentV2NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Attachment not found"]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	reader, err := client.GetAttachmentContentV2("99999")
	assert.Error(t, err)
	assert.Nil(t, reader)
}
