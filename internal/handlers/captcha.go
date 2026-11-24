package handlers

import (
	"context"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
)

// CAP proof-of-work gate shared by the public endpoints that require one when
// the backend runs with CAPTCHA_MODE=on (login, media scan).
//
// verifyCaptcha returns "" when the request may proceed — the verifier is
// disabled (CAPTCHA_MODE=off) or the supplied token was accepted — and the
// ErrorResponse message for a 400 response otherwise. The caller is
// responsible for mapping the message onto its operation's 400 response, so
// CAPTCHA failures are always 400, never 401/403.
func (s *shared) verifyCaptcha(ctx context.Context, token *gen.CapToken) string {
	if s.verifier == nil || !s.verifier.Enabled() {
		return ""
	}
	if token == nil || *token == "" {
		return "captcha required"
	}
	ok, err := s.verifier.Verify(ctx, *token)
	if err != nil {
		return "captcha unavailable"
	}
	if !ok {
		return "captcha failed"
	}
	return ""
}
