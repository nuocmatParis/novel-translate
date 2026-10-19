package verify

import (
	"errors"
	"strings"

	"github.com/voocel/ainovel-cli/internal/translation/domain"
	"github.com/voocel/ainovel-cli/internal/translation/epub"
)

type Image struct {
	ID       string         `json:"id"`
	Resource string         `json:"resource"`
	Hash     string         `json:"hash"`
	Sequence int            `json:"sequence"`
	Decision *ImageDecision `json:"decision,omitempty"`
}
type ImageDecision struct {
	ID              string `json:"id"`
	SourceHash      string `json:"source_hash"`
	ResourceHash    string `json:"resource_hash"`
	Disposition     string `json:"disposition"`
	Reason          string `json:"reason"`
	Replacement     string `json:"replacement,omitempty"`
	ReplacementHash string `json:"replacement_hash,omitempty"`
	Transcript      string `json:"transcript,omitempty"`
}

func Images(book domain.Book) []Image {
	resources := map[string]string{}
	for _, r := range book.Resources {
		resources[r.Path] = r.Hash
	}
	var out []Image
	for _, b := range book.Blocks {
		if b.Kind == "image" {
			out = append(out, Image{ID: b.ID, Resource: b.Resource, Hash: resources[b.Resource], Sequence: b.Sequence})
		}
		for _, t := range b.Tokens {
			if t.Kind == "void" && (t.Name == "img" || t.Name == "image") {
				href := t.Attributes["src"]
				if href == "" {
					href = t.Attributes["href"]
				}
				if href == "" {
					href = t.Attributes["xlink:href"]
				}
				p, _, e := epub.Resolve(b.Document, href)
				if e == nil {
					out = append(out, Image{ID: b.ID + ":" + t.ID, Resource: p, Hash: resources[p], Sequence: b.Sequence})
				}
			}
		}
	}
	return out
}
func ImageGate(book domain.Book, decisions map[string]ImageDecision) []domain.Diagnostic {
	var issues []domain.Diagnostic
	for _, image := range Images(book) {
		d, ok := decisions[image.ID]
		if !ok || d.SourceHash != book.SourceHash || d.ResourceHash != image.Hash || strings.TrimSpace(d.Reason) == "" {
			issues = append(issues, domain.Diagnostic{Code: "image-text-unresolved", Path: image.Resource, Detail: image.ID + ": cần đánh giá chữ trong hình", Blocking: true})
			continue
		}
		switch d.Disposition {
		case "no-text", "keep-original":
			// Hợp lệ: đã kiểm tra và xác nhận
		case "replace":
			if strings.TrimSpace(d.Replacement) == "" && strings.TrimSpace(d.ReplacementHash) == "" {
				issues = append(issues, domain.Diagnostic{Code: "image-replace-missing", Path: image.Resource, Detail: image.ID + ": thiếu file ảnh thay thế", Blocking: true})
			}
		case "transcript":
			if strings.TrimSpace(d.Transcript) == "" {
				issues = append(issues, domain.Diagnostic{Code: "image-transcript-empty", Path: image.Resource, Detail: image.ID + ": thiếu nội dung transcript bản dịch", Blocking: true})
			}
		default:
			issues = append(issues, domain.Diagnostic{Code: "image-disposition-invalid", Path: image.Resource, Detail: image.ID + ": disposition không hợp lệ (" + d.Disposition + ")", Blocking: true})
		}
	}
	return issues
}
func AssessImage(book domain.Book, id, disposition, reason, replacement, transcript string, replacementBytes []byte) (ImageDecision, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ImageDecision{}, errors.New("cần ghi rõ lý do/bằng chứng đánh giá (--reason)")
	}
	disposition = strings.TrimSpace(disposition)
	switch disposition {
	case "no-text", "keep-original":
		// Hợp lệ
	case "replace":
		if strings.TrimSpace(replacement) == "" && len(replacementBytes) == 0 {
			return ImageDecision{}, errors.New("disposition replace cần đường dẫn file ảnh thay thế (--file)")
		}
	case "transcript":
		if strings.TrimSpace(transcript) == "" {
			return ImageDecision{}, errors.New("disposition transcript cần nội dung dịch chữ trong ảnh (--transcript hoặc --transcript-file)")
		}
	default:
		return ImageDecision{}, errors.New("disposition phải là no-text, keep-original, replace hoặc transcript")
	}
	for _, image := range Images(book) {
		if image.ID == id {
			repHash := ""
			if len(replacementBytes) > 0 {
				repHash = domain.Hash(replacementBytes)
			}
			return ImageDecision{
				ID:              id,
				SourceHash:      book.SourceHash,
				ResourceHash:    image.Hash,
				Disposition:     disposition,
				Reason:          reason,
				Replacement:     replacement,
				ReplacementHash: repHash,
				Transcript:      transcript,
			}, nil
		}
	}
	return ImageDecision{}, errors.New("image occurrence ID không tồn tại")
}
