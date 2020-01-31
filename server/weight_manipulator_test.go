package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeightManipulator(t *testing.T) {

	type test struct {
		originalWeight float32
		expression     string
		expectedResult float32
		expectedError  error
		name           string
	}

	tests := []test{

		test{
			originalWeight: 2.0,
			expression:     "7",
			expectedResult: 7.0,
			expectedError:  nil,
			name:           "single operand, no operators",
		},
		test{
			originalWeight: 2.0,
			expression:     "17",
			expectedResult: 17.0,
			expectedError:  nil,
			name:           "single operand, more than one digit",
		},
		test{
			originalWeight: 2.0,
			expression:     "15.662",
			expectedResult: 15.662,
			expectedError:  nil,
			name:           "single operand, floating point using . as decimal",
		},
		test{
			originalWeight: 2.0,
			expression:     "w * 2",
			expectedResult: 4.0,
			expectedError:  nil,
			name:           "simple multiplication",
		},
		test{
			originalWeight: 2.0,
			expression:     "w * 2 * 3 * 4",
			expectedResult: 48.0,
			expectedError:  nil,
			name:           "multiplication with several operands",
		},
		test{
			originalWeight: 2.0,
			expression:     "w + 3",
			expectedResult: 5.0,
			expectedError:  nil,
			name:           "simple addition",
		},
		test{
			originalWeight: 2.0,
			expression:     "w + 3 + 7",
			expectedResult: 12.0,
			expectedError:  nil,
			name:           "additional with several operands",
		},
		test{
			originalWeight: 2.0,
			expression:     "1+2*3+4",
			expectedResult: 11.0,
			expectedError:  nil,
			name:           "mixing operators with different precedence",
		},
		test{
			originalWeight: 2.0,
			expression:     "1+2*3-4",
			expectedResult: 3.0,
			expectedError:  nil,
			name:           "mixing operators with different precedence, including -",
		},
		test{
			originalWeight: 2.0,
			expression:     "1+2/4-4",
			expectedResult: -2.5,
			expectedError:  nil,
			name:           "mixing operators with different precedence, including /",
		},
		test{
			originalWeight: 7.0,
			expression:     "1+ 2.5/7 * w -4/2",
			expectedResult: 1.5,
			expectedError:  nil,
			name:           "long expression including all operators",
		},
		test{
			originalWeight: 7.0,
			expression:     "w * w",
			expectedResult: 49,
			expectedError:  nil,
			name:           "including the weight variable multiple times",
		},
		test{
			originalWeight: 7.0,
			expression:     "2 * (1+3)",
			expectedError:  fmt.Errorf("using parantheses in the expression is not supported"),
			name:           "using parantheses",
		},
		test{
			originalWeight: 7.0,
			expression:     "a + b * c",
			expectedError:  fmt.Errorf("unrecognized variable 'a', use 'w' to represent original weight"),
			name:           "using a variable other than w",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := NewEvaluator(test.expression, test.originalWeight).Do()
			require.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedResult, res)
		})

	}
}
