package main

import (
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
			expression:     "1+ 2/7 * w -4/2",
			expectedResult: 1,
			expectedError:  nil,
			name:           "long expression including all operators",
		},
		test{
			originalWeight: 7.0,
			expression:     "w * w",
			expectedResult: 49,
			expectedError:  nil,
			name:           "including the operator multiple times",
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
