package nebo

import "testing"

func TestViewBuilderBasic(t *testing.T) {
	view := NewView("dashboard", "My Dashboard").
		Heading("title", "Status", "h2").
		Text("info", "All systems operational").
		Button("refresh", "Refresh", "primary").
		Divider("sep").
		Build()

	if view.ViewID != "dashboard" {
		t.Errorf("ViewID = %q, want dashboard", view.ViewID)
	}
	if view.Title != "My Dashboard" {
		t.Errorf("Title = %q, want My Dashboard", view.Title)
	}
	if len(view.Blocks) != 4 {
		t.Fatalf("len(Blocks) = %d, want 4", len(view.Blocks))
	}

	// Heading
	if view.Blocks[0].Type != "heading" || view.Blocks[0].Text != "Status" || view.Blocks[0].Variant != "h2" {
		t.Errorf("block 0 = %+v, want heading/Status/h2", view.Blocks[0])
	}

	// Text
	if view.Blocks[1].Type != "text" || view.Blocks[1].Text != "All systems operational" {
		t.Errorf("block 1 = %+v, want text/All systems operational", view.Blocks[1])
	}

	// Button
	if view.Blocks[2].Type != "button" || view.Blocks[2].Text != "Refresh" || view.Blocks[2].Variant != "primary" {
		t.Errorf("block 2 = %+v, want button/Refresh/primary", view.Blocks[2])
	}

	// Divider
	if view.Blocks[3].Type != "divider" {
		t.Errorf("block 3 = %+v, want divider", view.Blocks[3])
	}
}

func TestViewBuilderAllBlockTypes(t *testing.T) {
	view := NewView("test", "Test View").
		Heading("h", "Title", "h1").
		Text("t", "Hello").
		Input("i", "default", "Type here...").
		Button("b", "Click", "secondary").
		Select("s", "opt1", []SelectOption{{Label: "Option 1", Value: "opt1"}, {Label: "Option 2", Value: "opt2"}}).
		Toggle("tog", "Dark Mode", true).
		Divider("d").
		Image("img", "https://example.com/pic.png", "A picture").
		Build()

	if len(view.Blocks) != 8 {
		t.Fatalf("len(Blocks) = %d, want 8", len(view.Blocks))
	}

	expectedTypes := []string{"heading", "input", "button", "select", "toggle", "divider", "image"}
	// Check specific block types
	if view.Blocks[2].Type != "input" || view.Blocks[2].Placeholder != "Type here..." {
		t.Errorf("input block = %+v", view.Blocks[2])
	}
	if view.Blocks[4].Type != "select" || len(view.Blocks[4].Options) != 2 {
		t.Errorf("select block = %+v", view.Blocks[4])
	}
	if view.Blocks[5].Type != "toggle" || view.Blocks[5].Value != "true" {
		t.Errorf("toggle block = %+v", view.Blocks[5])
	}
	if view.Blocks[7].Type != "image" || view.Blocks[7].Src != "https://example.com/pic.png" {
		t.Errorf("image block = %+v", view.Blocks[7])
	}

	_ = expectedTypes // used for documentation
}

func TestViewBuilderEmpty(t *testing.T) {
	view := NewView("empty", "Empty").Build()

	if len(view.Blocks) != 0 {
		t.Errorf("len(Blocks) = %d, want 0", len(view.Blocks))
	}
}

func TestToggleFalse(t *testing.T) {
	view := NewView("test", "Test").
		Toggle("t", "Off", false).
		Build()

	if view.Blocks[0].Value != "false" {
		t.Errorf("toggle value = %q, want false", view.Blocks[0].Value)
	}
}
