package verify

import (
	"testing"

	"github.com/voocel/ainovel-cli/internal/translation/domain"
)

func TestAssessImageAndGate(t *testing.T) {
	book := domain.Book{
		SourceHash: "hash123",
		Resources: []domain.Resource{
			{Path: "OPS/img1.jpg", Hash: "imghash1"},
			{Path: "OPS/img2.jpg", Hash: "imghash2"},
			{Path: "OPS/img3.jpg", Hash: "imghash3"},
			{Path: "OPS/img4.jpg", Hash: "imghash4"},
		},
		Blocks: []domain.Block{
			{ID: "b1", Kind: "image", Resource: "OPS/img1.jpg"},
			{ID: "b2", Kind: "image", Resource: "OPS/img2.jpg"},
			{ID: "b3", Kind: "image", Resource: "OPS/img3.jpg"},
			{ID: "b4", Kind: "image", Resource: "OPS/img4.jpg"},
		},
	}

	images := Images(book)
	if len(images) != 4 {
		t.Fatalf("expected 4 images, got %d", len(images))
	}

	// 1. no-text
	d1, err := AssessImage(book, "b1", "no-text", "ảnh minh họa đen trắng không có chữ", "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error for no-text: %v", err)
	}
	if d1.Disposition != "no-text" {
		t.Fatalf("expected disposition no-text, got %s", d1.Disposition)
	}

	// 2. keep-original
	d2, err := AssessImage(book, "b2", "keep-original", "chữ ký của tác giả giữ nguyên bản", "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error for keep-original: %v", err)
	}
	if d2.Disposition != "keep-original" {
		t.Fatalf("expected disposition keep-original, got %s", d2.Disposition)
	}

	// 3. replace
	fakeData := []byte("fake-replacement-image-data")
	d3, err := AssessImage(book, "b3", "replace", "ảnh bìa đã edit chữ tiếng Việt", "covers/vol01_vi.jpg", "", fakeData)
	if err != nil {
		t.Fatalf("unexpected error for replace: %v", err)
	}
	if d3.ReplacementHash != domain.Hash(fakeData) {
		t.Fatalf("replacement hash mismatch: got %s", d3.ReplacementHash)
	}

	// 4. transcript
	transcriptText := "Subaru: Đợi đã!\nEmilia: Có chuyện gì thế?"
	d4, err := AssessImage(book, "b4", "transcript", "lời thoại trên tranh giới thiệu nhân vật", "", transcriptText, nil)
	if err != nil {
		t.Fatalf("unexpected error for transcript: %v", err)
	}
	if d4.Transcript != transcriptText {
		t.Fatalf("expected transcript text preserved")
	}

	// Validation errors
	if _, err := AssessImage(book, "b1", "invalid-disp", "reason", "", "", nil); err == nil {
		t.Fatal("expected error on invalid disposition")
	}
	if _, err := AssessImage(book, "b1", "no-text", "   ", "", "", nil); err == nil {
		t.Fatal("expected error on empty reason")
	}
	if _, err := AssessImage(book, "b3", "replace", "reason", "", "", nil); err == nil {
		t.Fatal("expected error on replace without file")
	}
	if _, err := AssessImage(book, "b4", "transcript", "reason", "", "  ", nil); err == nil {
		t.Fatal("expected error on transcript without text")
	}
	if _, err := AssessImage(book, "non-existent", "no-text", "reason", "", "", nil); err == nil {
		t.Fatal("expected error on non-existent image ID")
	}

	// Gate: all assessed
	decisions := map[string]ImageDecision{
		"b1": d1,
		"b2": d2,
		"b3": d3,
		"b4": d4,
	}
	issues := ImageGate(book, decisions)
	if len(issues) != 0 {
		t.Fatalf("expected 0 issues when all images are properly assessed, got: %+v", issues)
	}

	// Gate: missing assessment for b4
	delete(decisions, "b4")
	issues = ImageGate(book, decisions)
	if len(issues) != 1 || issues[0].Code != "image-text-unresolved" {
		t.Fatalf("expected image-text-unresolved for missing b4, got: %+v", issues)
	}

	// Gate: invalid replace decision (empty replacement path and hash)
	decisions["b3"] = ImageDecision{ID: "b3", SourceHash: book.SourceHash, ResourceHash: "imghash3", Disposition: "replace", Reason: "reason"}
	issues = ImageGate(book, decisions)
	foundReplaceMissing := false
	for _, iss := range issues {
		if iss.Code == "image-replace-missing" {
			foundReplaceMissing = true
		}
	}
	if !foundReplaceMissing {
		t.Fatalf("expected image-replace-missing issue for invalid replace decision, got: %+v", issues)
	}
}
