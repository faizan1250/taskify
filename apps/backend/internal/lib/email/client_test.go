package email_test

import (
	"testing"
	"testing/fstest"

	"github.com/faizan1250/taskify/internal/lib/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_RenderTemplate_DirectoryIndependence(t *testing.T) {
	// Change working directory to an isolated empty temporary directory
	// to ensure templates are loaded strictly from embedded FS, not relative paths.
	t.Chdir(t.TempDir())

	client := &email.Client{}

	tests := []struct {
		name         string
		templateName email.Template
		data         map[string]any
		contains     []string
	}{
		{
			name:         "welcome email",
			templateName: email.TemplateWelcome,
			data: map[string]any{
				"UserFirstName": "TestUser",
			},
			contains: []string{"Welcome to Boilerplate", "TestUser"},
		},
		{
			name:         "due date reminder",
			templateName: email.TemplateDueDateReminder,
			data: map[string]any{
				"TodoTitle":    "Important Task",
				"TodoID":       "abc-123",
				"DueDate":      "Friday, Nov 10",
				"DaysUntilDue": 3,
			},
			contains: []string{"Important Task", "3"},
		},
		{
			name:         "overdue notification",
			templateName: email.TemplateOverdueNotification,
			data: map[string]any{
				"TodoTitle":   "Urgent Task",
				"TodoID":      "xyz-789",
				"DueDate":     "Monday, Nov 3",
				"DaysOverdue": 2,
			},
			contains: []string{"Urgent Task", "2"},
		},
		{
			name:         "weekly report",
			templateName: email.TemplateWeeklyReport,
			data: map[string]any{
				"WeekStart":      "Nov 1",
				"WeekEnd":        "Nov 7",
				"CompletedCount": 10,
				"ActiveCount":    5,
				"OverdueCount":   1,
				"CompletedTodos": nil,
				"OverdueTodos":   nil,
				"HasCompleted":   true,
				"HasOverdue":     true,
			},
			contains: []string{"Weekly Productivity Report", "Nov 1", "Nov 7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := client.RenderTemplate(tt.templateName, tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, rendered)
			for _, exp := range tt.contains {
				assert.Contains(t, rendered, exp)
			}
		})
	}
}

func TestClient_RenderTemplate_NotFound(t *testing.T) {
	t.Chdir(t.TempDir())

	client := &email.Client{}
	rendered, err := client.RenderTemplate(email.Template("nonexistent"), nil)
	require.Error(t, err)
	assert.Empty(t, rendered)
	assert.Contains(t, err.Error(), "failed to parse email template nonexistent")
}

func TestClient_SetFS(t *testing.T) {
	t.Chdir(t.TempDir())

	mockFS := fstest.MapFS{
		"emails/custom.html": &fstest.MapFile{
			Data: []byte("Hello {{.Name}}!"),
		},
	}

	client := &email.Client{}
	client.SetFS(mockFS)

	rendered, err := client.RenderTemplate(email.Template("custom"), map[string]any{
		"Name": "Gopher",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello Gopher!", rendered)
}
