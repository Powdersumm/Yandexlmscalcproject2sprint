package main

import "errors"

var (
	ErrInvalidExpression  = errors.New("invalid expression")
	ErrInvalidParentheses = errors.New("invalid parentheses")
	ErrInvalidZero        = errors.New("division by zero")
	ErrInvalidOperand     = errors.New("unknown operand")
	ErrInvalidValuesCount = errors.New("invalid number of values")
	ErrInvalidCalculation = errors.New("invalid calculation")
)
