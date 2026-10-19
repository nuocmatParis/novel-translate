package model

import (
	"github.com/voocel/ainovel-cli/internal/llmcontract"
	"sort"
)

func str() map[string]any { return map[string]any{"type": "string"} }
func object(props map[string]any) map[string]any {
	keys := []string{}
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return map[string]any{"type": "object", "properties": props, "required": keys, "additionalProperties": false}
}
func array(items map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": items}
}
func TranslationContract() llmcontract.Contract {
	return llmcontract.Contract{Name: "jp_vi_translation", Description: "Vietnamese translation of owned source units", Schema: object(map[string]any{"units": array(object(map[string]any{"id": str(), "target": str()})), "uncertainties": array(object(map[string]any{"block_id": str(), "quote": str(), "explanation": str()})), "continuity_candidates": array(object(map[string]any{"block_id": str(), "quote": str(), "speaker": str(), "listener": str(), "value": str()})), "rule_candidates": array(object(map[string]any{"block_id": str(), "quote": str(), "key": str(), "value": str(), "kind": str()}))})}
}
func ReviewContract() llmcontract.Contract {
	return llmcontract.Contract{Name: "jp_vi_review", Description: "Clause-level source and target comparison", Schema: object(map[string]any{"reviewed": array(object(map[string]any{"id": str(), "target_hash": str()})), "issues": array(object(map[string]any{"id": str(), "type": str(), "severity": str(), "block_id": str(), "source_quote": str(), "target_quote": str(), "explanation": str()})), "ruby_decisions": array(object(map[string]any{"index": map[string]any{"type": "integer"}, "base": str(), "reading": str(), "decision": str(), "target_quote": str(), "reason": str()}))})}
}
