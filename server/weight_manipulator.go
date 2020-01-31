package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
)

type Evaluator struct {
	expression     string
	originalWeight float32
	parsedStack    []string
}

func NewEvaluator(expr string, weight float32) *Evaluator {
	return &Evaluator{expression: expr, originalWeight: weight}
}

func (e *Evaluator) Do() (float32, error) {
	if err := e.parseExpression(); err != nil {
		return 0, err
	}

	return e.evaluate()
}

func (e *Evaluator) parseExpression() error {
	var operatorStack []string

	// a number might be made of of multiple digits, this variable acts as a
	// temporary storage for signle digits
	var currOperandDigits []string

	for _, r := range e.expression {
		if unicode.IsSpace(r) {
			continue
		}

		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			// dont' direclty append to stack, append to OperandDigitsStack first, as
			// this might be just a single digit of a multi-digit number
			currOperandDigits = append(currOperandDigits, string(r))
			continue
		}

		if isOperator(string(r)) {
			// check if the operandStack contains elements, if so we need to append
			// and clear that first

			if len(currOperandDigits) > 0 {
				e.parsedStack = append(e.parsedStack, strings.Join(currOperandDigits, ""))
				currOperandDigits = nil
			}

			operatorStack = append(operatorStack, string(r))
			continue
		}

		return fmt.Errorf("uncrecognized operator: %s", string(r))
	}

	// in case the number ends with an operand, we need to check again if the
	// temp digit stack still contains elements
	if len(currOperandDigits) > 0 {
		e.parsedStack = append(e.parsedStack, strings.Join(currOperandDigits, ""))
		currOperandDigits = nil
	}

	e.parsedStack = append(e.parsedStack, operatorStack...)
	spew.Dump(e.parsedStack)
	return nil
}

func (e Evaluator) evaluate() (float32, error) {
	var operandStack []float32
	for _, item := range e.parsedStack {
		if !isOperator(item) {
			// not an operator, so it must be an operand
			num, err := e.parseNumberOrVariable(item)
			if err != nil {
				return 0, err
			}

			operandStack = append(operandStack, num)
			continue
		}

		// is an operator
		if len(operandStack) < 2 {
			return 0, fmt.Errorf("invalid or unsupported math expression")
		}

		op1 := operandStack[len(operandStack)-1]
		op2 := operandStack[len(operandStack)-2]
		operandStack = operandStack[:len(operandStack)-2]

		switch item {
		case "+":
			operandStack = append(operandStack, op1+op2)
		case "*":
			operandStack = append(operandStack, op1*op2)
		default:
			return 0, fmt.Errorf("this should be unreachable")
		}

	}

	if len(operandStack) != 1 {
		return 0, fmt.Errorf("could not evaluate mathematical expression")
	}

	return operandStack[0], nil
}

func isOperator(in string) bool {
	switch in {
	case "*", "+":
		return true
	default:
		return false
	}
}

func (e *Evaluator) parseNumberOrVariable(in string) (float32, error) {
	r := rune(in[0])
	if unicode.IsNumber(r) {
		res, err := strconv.ParseFloat(in, 32)
		if err != nil {
			return 0, err
		}

		return float32(res), nil
	} else {
		if in == "w" {
			return e.originalWeight, nil
		}
		return 0, fmt.Errorf("unrecoginzed variable '%s', use 'w' to represend original weight", in)
	}
}
