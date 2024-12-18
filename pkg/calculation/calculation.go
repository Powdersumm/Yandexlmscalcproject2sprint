package calculation

import (
	"strconv"
	"strings"
)

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	result, err := evaluateexpression(expression)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func searchnumbers(expression string, index int) (float64, int) {
	start := index
	for index < len(expression) && (isDigit(expression[index]) || expression[index] == '.') {
		index++
	}
	val, _ := strconv.ParseFloat(expression[start:index], 64)
	return val, index
}

func precedence(op rune) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	}
	return 0
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isOperator(char byte) bool {
	return char == '+' || char == '-' || char == '*' || char == '/'
}

func evaluateexpression(expression string) (float64, error) {
	var ops []rune
	var values []float64
	for i := 0; i < len(expression); i++ {
		char := expression[i]
		if isDigit(char) {
			val, nextindex := searchnumbers(expression, i)
			values = append(values, val)
			i = nextindex - 1
		} else if char == '(' {
			ops = append(ops, '(')
		} else if char == ')' {
			for len(ops) > 0 && ops[len(ops)-1] != '(' {
				var err error
				values, err = attachOperator(ops[len(ops)-1], values)
				if err != nil {
					return 0, ErrInvalidExpression
				}
				ops = ops[:len(ops)-1]
			}
			if len(ops) == 0 {
				return 0, ErrInvalidParentheses
			}
			ops = ops[:len(ops)-1]
		} else if isOperator(char) {
			for len(ops) > 0 && precedence(rune(char)) <= precedence(ops[len(ops)-1]) {
				var err error
				values, err = attachOperator(ops[len(ops)-1], values)
				if err != nil {
					return 0, err
				}
				ops = ops[:len(ops)-1]
			}
			ops = append(ops, rune(char))
		} else {
			return 0, ErrInvalidCalculation
		}
	}
	for len(ops) > 0 {
		var err error
		values, err = attachOperator(ops[len(ops)-1], values)
		if err != nil {
			return 0, err
		}
		ops = ops[:len(ops)-1]
	}
	if len(values) != 1 {
		return 0, ErrInvalidValuesCount
	}
	return values[0], nil
}

func attachOperator(op rune, values []float64) ([]float64, error) {
	if len(values) < 2 {
		return values, ErrInvalidValuesCount
	}
	a := values[len(values)-1]
	b := values[len(values)-2]
	values = values[:len(values)-2]
	var result float64
	switch op {
	case '+':
		result = b + a
	case '-':
		result = b - a
	case '*':
		result = b * a
	case '/':
		if a == 0 {
			return values, ErrInvalidZero
		}
		result = b / a
	default:
		return values, ErrInvalidoperand
	}
	values = append(values, result)
	return values, nil
}
