package abi

import (
	"errors"
	"fmt"
	"strings"
)

type SelectorMarshaling struct {
	Name   string               `json:"name"`
	Type   string               `json:"type"`
	Inputs []ArgumentMarshaling `json:"inputs"`
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isSpace(c byte) bool {
	return c == ' '
}

func isIdentifierSymbol(c byte) bool {
	return c == '$' || c == '_'
}

func parseToken(unescapedSelector string, isIdent bool) (string, string, error) {
	if len(unescapedSelector) == 0 {
		return "", "", errors.New("empty token")
	}
	firstChar := unescapedSelector[0]
	position := 1
	if !(isAlpha(firstChar) || (isIdent && isIdentifierSymbol(firstChar))) {
		return "", "", fmt.Errorf("invalid token start: %c", firstChar)
	}
	for position < len(unescapedSelector) {
		char := unescapedSelector[position]
		if !(isAlpha(char) || isDigit(char) || (isIdent && isIdentifierSymbol(char)) || (!isIdent && isSpace(char))) {
			break
		}
		position++
	}
	return unescapedSelector[:position], unescapedSelector[position:], nil
}

func parseIdentifier(unescapedSelector string) (string, string, error) {
	return parseToken(unescapedSelector, true)
}

func parseElementaryType(unescapedSelector string) (parsedType string, rest string, err error) {
	parsedType, rest, err = parseToken(unescapedSelector, false)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse elementary type: %v", err)
	}
	parts := strings.Split(parsedType, " ")
	if len(parts) > 1 {
		parsedType = parsedType[len(parts[0])+1:]
	}
	// handle arrays
	for len(rest) > 0 && rest[0] == '[' {
		parsedType = parsedType + string(rest[0])
		rest = rest[1:]
		for len(rest) > 0 && isDigit(rest[0]) {
			parsedType = parsedType + string(rest[0])
			rest = rest[1:]
		}
		if len(rest) == 0 || rest[0] != ']' {
			return "", "", fmt.Errorf("failed to parse array: expected ']', got %c", unescapedSelector[0])
		}
		parsedType = parsedType + string(rest[0])
		rest = rest[1:]
	}
	return parsedType, rest, nil
}

func parseElementaryTypeWithName(unescapedSelector string) (parsedType string, rest string, err error) {
	parsedType, rest, err = parseToken(unescapedSelector, false)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse elementary type: %v", err)
	}
	// handle arrays
	for len(rest) > 0 && rest[0] == '[' {
		parsedType = parsedType + string(rest[0])
		rest = rest[1:]
		for len(rest) > 0 && isDigit(rest[0]) {
			parsedType = parsedType + string(rest[0])
			rest = rest[1:]
		}
		if len(rest) == 0 || rest[0] != ']' {
			return "", "", fmt.Errorf("failed to parse array: expected ']', got %c", unescapedSelector[0])
		}
		parsedType = parsedType + string(rest[0])
		rest = rest[1:]
	}
	return parsedType, rest, nil
}

func parseCompositeType(unescapedSelector string) (result []interface{}, rest string, err error) {
	if len(unescapedSelector) == 0 || unescapedSelector[0] != '(' {
		return nil, "", fmt.Errorf("expected '(...', got %s", unescapedSelector)
	}
	var parsedType interface{}
	parsedType, rest, err = parseType(unescapedSelector[1:])
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse type: %v", err)
	}
	result = []interface{}{parsedType}
	for len(rest) > 0 && rest[0] != ')' {
		parsedType, rest, err = parseType(rest[1:])
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse type: %v", err)
		}
		result = append(result, parsedType)
	}
	if len(rest) == 0 || rest[0] != ')' {
		return nil, "", fmt.Errorf("expected ')', got '%s'", rest)
	}
	if len(rest) >= 3 && rest[1] == '[' && rest[2] == ']' {
		return append(result, "[]"), rest[3:], nil
	}
	return result, rest[1:], nil
}

func parseCompositeTypeWithName(unescapedSelector string) (result []interface{}, rest string, err error) {
	var name string
	parts := strings.Split(unescapedSelector, " ")
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("expected name in the beginning, got %s", unescapedSelector)
	} else {
		name = parts[0]
		unescapedSelector = unescapedSelector[len(parts[0])+1:]
	}
	if len(unescapedSelector) == 0 || unescapedSelector[0] != '(' {
		return nil, "", fmt.Errorf("expected '(...', got %s", unescapedSelector)
	}
	result = []interface{}{name}
	var parsedType interface{}
	var counter int64
	parsedType, rest, err = parseTypeWithName(unescapedSelector[1:], counter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse type: %v", err)
	}
	result = append(result, parsedType)
	for len(rest) > 0 && rest[0] != ')' {
		counter += 1
		parsedType, rest, err = parseTypeWithName(rest[1:], counter)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse type: %v", err)
		}
		result = append(result, parsedType)
	}
	if len(rest) == 0 || rest[0] != ')' {
		return nil, "", fmt.Errorf("expected ')', got '%s'", rest)
	}
	if len(rest) >= 3 && rest[1] == '[' && rest[2] == ']' {
		return append(result, "[]"), rest[3:], nil
	}
	return result, rest[1:], nil
}

func parseFunctionsArgs(unescapedSelector string) (result []interface{}, rest string, err error) {
	if len(unescapedSelector) == 0 || unescapedSelector[0] != '(' {
		return nil, "", fmt.Errorf("expected '(...', got %s", unescapedSelector)
	}
	var parsedType interface{}
	var counter int64
	parsedType, rest, err = parseTypeWithName(unescapedSelector[1:], counter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse type: %v", err)
	}
	result = []interface{}{parsedType}

	for len(rest) > 0 && rest[0] != ')' {
		counter += 1
		parsedType, rest, err = parseTypeWithName(rest[1:], counter)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse type: %v", err)
		}
		result = append(result, parsedType)
	}
	if len(rest) == 0 || rest[0] != ')' {
		return nil, "", fmt.Errorf("expected ')', got '%s'", rest)
	}
	if len(rest) >= 3 && rest[1] == '[' && rest[2] == ']' {
		return append(result, "[]"), rest[3:], nil
	}
	return result, rest[1:], nil
}

func parseType(unescapedSelector string) (interface{}, string, error) {
	parts := strings.Split(unescapedSelector, " ")
	if len(parts) > 1 {
		unescapedSelector = unescapedSelector[len(parts[0])+1:]
	}
	if len(unescapedSelector) == 0 {
		return nil, "", errors.New("empty type")
	}
	if unescapedSelector[0] == '(' {
		return parseCompositeType(unescapedSelector)
	} else {
		return parseElementaryType(unescapedSelector)
	}
}

func parseTypeWithName(unescapedSelector string, counter int64) (interface{}, string, error) {
	name, rest, _ := parseIdentifier(unescapedSelector)
	if len(rest) > 0 && rest[0] == ' ' {
		unescapedSelector = unescapedSelector[len(name)+1:]
	} else {
		name = fmt.Sprintf("name%d", counter)
	}
	if len(unescapedSelector) == 0 {
		return nil, "", errors.New("empty type")
	}
	if unescapedSelector[0] == '(' {
		return parseCompositeTypeWithName(fmt.Sprintf("%v %v", name, unescapedSelector))
	} else {
		return parseElementaryTypeWithName(fmt.Sprintf("%v %v", name, unescapedSelector))
	}
}

func assembleArgs(args []interface{}) (arguments []ArgumentMarshaling, err error) {
	arguments = make([]ArgumentMarshaling, 0)
	for _, arg := range args {
		var name string
		if s, ok := arg.(string); ok {
			if s == "[]" {
				arguments = append(arguments, ArgumentMarshaling{Name: name, Type: s, InternalType: s})
				continue
			}
			parts := strings.Split(s, " ")
			if len(parts) < 2 {
				return nil, fmt.Errorf("no name in arg %s", s)
			} else {
				name = parts[0]
				s = s[len(name)+1:]
			}
			arguments = append(arguments, ArgumentMarshaling{Name: name, Type: s, InternalType: s})
		} else if components, ok := arg.([]interface{}); ok {
			var subArgs []ArgumentMarshaling
			if len(components) < 2 {
				return nil, fmt.Errorf("no name in components %s", components)
			} else {
				name = components[0].(string)
				components = components[1:]
			}
			subArgs, err = assembleArgs(components)
			if err != nil {
				return nil, fmt.Errorf("failed to assemble components: %v", err)
			}
			tupleType := "tuple"
			if len(subArgs) != 0 && subArgs[len(subArgs)-1].Type == "[]" {
				subArgs = subArgs[:len(subArgs)-1]
				tupleType = "tuple[]"
			}
			arguments = append(arguments, ArgumentMarshaling{Name: name, Type: tupleType, InternalType: tupleType, Components: subArgs})
		} else {
			return nil, fmt.Errorf("failed to assemble args: unexpected type %T", arg)
		}
	}
	return arguments, nil
}

// ParseSelector converts a method selector into a struct that can be JSON encoded
// and consumed by other functions in this package.
// Note, although uppercase letters are not part of the ABI spec, this function
// still accepts it as the general format is valid.
func ParseSelector(unescapedSelector string) (m SelectorMarshaling, err error) {
	var name, rest string
	name, rest, err = parseIdentifier(unescapedSelector)
	if err != nil {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': %v", unescapedSelector, err)
	}
	args := make([]interface{}, 0)
	if len(rest) >= 2 && rest[0] == '(' && rest[1] == ')' {
		rest = rest[2:]
	} else {
		args, rest, err = parseFunctionsArgs(rest)
		if err != nil {
			return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': %v", unescapedSelector, err)
		}
	}
	if len(rest) > 0 {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': unexpected string '%s'", unescapedSelector, rest)
	}

	// Reassemble the fake ABI and construct the JSON
	var fakeArgs []ArgumentMarshaling
	fakeArgs, err = assembleArgs(args)
	if err != nil {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector: %v", err)
	}

	return SelectorMarshaling{name, "function", fakeArgs}, nil
}
