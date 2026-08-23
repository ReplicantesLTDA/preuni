package mail

import "testing"

// TestRender_UnsupportedTypeReturnsAnError covers render's default branch
// (neither *html/template.Template nor *text/template.Template), which the
// existing Build/buildWelcome/etc. tests never reach since they always pass
// a real template value.
func TestRender_UnsupportedTypeReturnsAnError(t *testing.T) {
	_, err := render(123, nil)
	if err == nil {
		t.Fatal("expected an error for an unsupported template type")
	}
}
