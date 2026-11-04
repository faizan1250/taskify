package email

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"

	"github.com/faizan1250/taskify/internal/config"
	"github.com/faizan1250/taskify/templates"
	"github.com/pkg/errors"
	"github.com/resend/resend-go/v2"
	"github.com/rs/zerolog"
)

type Client struct {
	client *resend.Client
	logger *zerolog.Logger
	fs     fs.FS
}

func NewClient(cfg *config.Config, logger *zerolog.Logger) *Client {
	return &Client{
		client: resend.NewClient(cfg.Integration.ResendAPIKey),
		logger: logger,
		fs:     templates.Files,
	}
}

func (c *Client) SetFS(fsys fs.FS) {
	c.fs = fsys
}

func (c *Client) RenderTemplate(templateName Template, data map[string]any) (string, error) {
	fsys := c.fs
	if fsys == nil {
		fsys = templates.Files
	}

	tmplPath := fmt.Sprintf("emails/%s.html", templateName)
	tmpl, err := template.ParseFS(fsys, tmplPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse email template %s", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", errors.Wrapf(err, "failed to execute email template %s", templateName)
	}

	return body.String(), nil
}

func (c *Client) SendEmail(to, subject string, templateName Template, data map[string]any) error {
	htmlBody, err := c.RenderTemplate(templateName, data)
	if err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", "Boilerplate", "onboarding@resend.dev"),
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err = c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
