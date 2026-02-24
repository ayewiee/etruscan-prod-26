package dsl

import (
	"fmt"
	"strconv"
	"strings"
)

// precedence levels for normalization logic
const (
	PrecedenceNone = iota
	PrecedenceOr
	PrecedenceAnd
	PrecedenceNot
	PrecedenceCompare
	PrecedenceAtom
)

type Node interface {
	Eval(ctx map[string]interface{}) (bool, error)
	String() string
	Precedence() int
}

// BinaryExpression - AND / OR
type BinaryExpression struct {
	Left     Node
	Operator string // "AND" | "OR"
	Right    Node
}

func (b *BinaryExpression) Eval(ctx map[string]interface{}) (bool, error) {
	lv, err := b.Left.Eval(ctx)
	if err != nil {
		return false, err
	}
	// short-circuiting logic
	if b.Operator == "OR" && lv {
		return true, nil
	}
	if b.Operator == "AND" && !lv {
		return false, nil
	}

	rv, err := b.Right.Eval(ctx)
	if err != nil {
		return false, err
	}

	if b.Operator == "AND" {
		return lv && rv, nil
	}
	return lv || rv, nil
}

// String implements the normalization rules (spacing, parenthesis removal)
func (b *BinaryExpression) String() string {
	leftStr := b.Left.String()
	if b.Left.Precedence() < b.Precedence() {
		leftStr = "(" + leftStr + ")"
	}

	rightStr := b.Right.String()
	if b.Right.Precedence() < b.Precedence() {
		rightStr = "(" + rightStr + ")"
	}

	return fmt.Sprintf("%s %s %s", leftStr, b.Operator, rightStr)
}

func (b *BinaryExpression) Precedence() int {
	if b.Operator == "AND" {
		return PrecedenceAnd
	}
	return PrecedenceOr
}

// UnaryExpression - NOT
type UnaryExpression struct {
	Right Node
}

func (u *UnaryExpression) Eval(ctx map[string]interface{}) (bool, error) {
	val, err := u.Right.Eval(ctx)
	return !val, err
}

func (u *UnaryExpression) String() string {
	rightStr := u.Right.String()
	// if child has lower precedence, wrap it (e.g. NOT (A OR B))
	if u.Right.Precedence() < u.Precedence() {
		rightStr = "(" + rightStr + ")"
	}
	return "NOT " + rightStr
}

func (u *UnaryExpression) Precedence() int {
	return PrecedenceNot
}

// ComparisonExpression - field op value (or value list for IN / NOT IN)
type ComparisonExpression struct {
	Field    string
	Operator string
	Value    interface{} // string, float64 or []interface{} for IN/NOT IN
}

func (c *ComparisonExpression) Eval(ctx map[string]interface{}) (bool, error) {
	fieldVal, err := getFieldValue(ctx, c.Field)
	if err != nil {
		return false, err
	}
	if fieldVal == nil {
		return false, nil
	}

	// normalize comparison to float/string or list membership
	// note: we assume the parser validated types reasonably well, yet still do runtime checks.
	switch v := c.Value.(type) {
	case []interface{}:
		// IN / NOT IN
		in := false
		for _, item := range v {
			switch tv := item.(type) {
			case float64:
				// numeric list: convert field to float
				fVal, ok := fieldVal.(float64)
				if !ok {
					if iVal, okInt := fieldVal.(int); okInt {
						fVal = float64(iVal)
					} else {
						continue
					}
				}
				if compareFloats(fVal, "=", tv) {
					in = true
					break
				}
			case string:
				sVal, ok := fieldVal.(string)
				if !ok {
					continue
				}
				if compareStrings(sVal, "=", tv) {
					in = true
					break
				}
			}
		}
		if c.Operator == "IN" {
			return in, nil
		}
		if c.Operator == "NOT IN" {
			return !in, nil
		}
		return false, fmt.Errorf("unsupported operator %s for list value", c.Operator)
	case float64:
		// numeric comparison
		fVal, ok := fieldVal.(float64)
		if !ok {
			// try converting int to float if needed
			if iVal, okInt := fieldVal.(int); okInt {
				fVal = float64(iVal)
			} else {
				return false, fmt.Errorf("runtime type mismatch for field %s", c.Field)
			}
		}
		return compareFloats(fVal, c.Operator, v), nil
	case string:
		// string or semantic version comparison
		sVal, ok := fieldVal.(string)
		if !ok {
			return false, fmt.Errorf("runtime type mismatch for field %s", c.Field)
		}

		// If both look like versions (e.g. 1.5.7), compare as semantic versions.
		if cmp, okVer := compareVersions(sVal, v); okVer {
			switch c.Operator {
			case ">":
				return cmp > 0, nil
			case ">=":
				return cmp >= 0, nil
			case "<":
				return cmp < 0, nil
			case "<=":
				return cmp <= 0, nil
			case "=":
				return cmp == 0, nil
			case "!=":
				return cmp != 0, nil
			default:
				// fall back to plain string comparison for unsupported ops
			}
		}

		return compareStrings(sVal, c.Operator, v), nil
	}
	return false, nil
}

func (c *ComparisonExpression) String() string {
	valStr := ""
	switch v := c.Value.(type) {
	case string:
		valStr = fmt.Sprintf("'%s'", v)
	case float64:
		// Format to remove trailing zeros if integer-like
		valStr = strconv.FormatFloat(v, 'f', -1, 64)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			switch tv := item.(type) {
			case string:
				parts = append(parts, fmt.Sprintf("'%s'", tv))
			case float64:
				parts = append(parts, strconv.FormatFloat(tv, 'f', -1, 64))
			default:
				parts = append(parts, "?")
			}
		}
		valStr = "(" + strings.Join(parts, ", ") + ")"
	default:
		valStr = "?"
	}
	return fmt.Sprintf("%s %s %s", c.Field, c.Operator, valStr)
}

func (c *ComparisonExpression) Precedence() int {
	return PrecedenceCompare
}
