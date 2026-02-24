package dsl

import "etruscan/internal/domain/models"

// Evaluate parses and evaluates a DSL expression against a flat context map.
// returns:
// - matches: whether the context matches the rule (false if rule is empty)
// - validation: parse result with normalized expression and errors
// - err: runtime evaluation error (e.g. type mismatch)
func Evaluate(rule string, ctx map[string]interface{}) (matches bool, validation *models.DSLValidationResult, err error) {
	if rule == "" {
		// no rule means "everyone can participate"
		return true, &models.DSLValidationResult{
			IsValid:              true,
			NormalizedExpression: nil,
			Errors:               nil,
		}, nil
	}

	p := NewParser(rule)
	node, parseErrs := p.Parse()
	if len(parseErrs) > 0 || node == nil {
		return false, &models.DSLValidationResult{
			IsValid:              false,
			NormalizedExpression: nil,
			Errors:               parseErrs,
		}, nil
	}

	res, evalErr := node.Eval(ctx)
	norm := node.String()

	return res, &models.DSLValidationResult{
		IsValid:              true,
		NormalizedExpression: &norm,
		Errors:               nil,
	}, evalErr
}
