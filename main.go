package evaluate

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(Evaluate("true && true", func(v string) (bool, error) {
		if v == "true" {
			return true, nil
		}
		return false, nil
	}, nil))
}

// Evaluate :
func Evaluate(v string, check func(string) (bool, error), validateOperand func(string) error) (bool, error) {

	fail := func(err error) (bool, error) {
		return false, err
	}

	if check == nil {
		return fail(fmt.Errorf("check func cannot be nil"))
	}

	v = strings.ReplaceAll(v, " ", "")
	length := len(v)
	index := 0
	parenthesesStack := make([]int, 0)

	findIndex := func(index int, chars string) int {
		endIndex := strings.IndexAny(v[index:], chars)
		if endIndex == -1 {
			endIndex = length
		} else {
			endIndex += index
		}
		return endIndex
	}

	var evaluate func() (bool, error)
	evaluate = func() (bool, error) {
		result := false
		operation := byte('n')
		for index < length {
			not := false
			if v[index] == '!' {
				not = true
				index++
			}
			var tmpResult bool
			var err error
			if v[index] == '(' {
				parenthesesStack = append(parenthesesStack, index)
				index++
				tmpResult, err = evaluate()
				if err != nil {
					return fail(err)
				}
			} else {
				endIndex := findIndex(index, "()&|")
				min := func(a, b int) int {
					if a < b {
						return a
					}
					return b
				}
				if v[min(endIndex, length-1)] == '(' {
					return fail(fmt.Errorf("syntax error: invalid expression"))
				}
				tmpResult, err = check(v[index:endIndex])
				if err != nil {
					return fail(err)
				}
				index = endIndex
			}
			if not {
				tmpResult = !tmpResult
			}

			switch operation {
			case '&':
				result = result && tmpResult
			case '|':
				result = result || tmpResult
			default:
				result = tmpResult
			}
			operation = byte('n')

		NextOperator:
			if index < length && v[index] == ')' {
				if len(parenthesesStack) == 0 {
					return fail(fmt.Errorf("syntax error: invalid expression"))
				}
				parenthesesStack = parenthesesStack[:len(parenthesesStack)-1]
				index++
				break
			} else if index+1 < length {
				operation = byte('n')
				validateOperator := func(operator byte) error {
					if v[index+1] == operator {
						operation = operator
						index += 2
					} else {
						return fmt.Errorf("syntax error: invalid operator at %v%c", operator, v[index+1])
					}
					if (result && operator == '|') || (!result && operator == '&') {
						if index < length {
							if v[index] != '(' {
								endIndex := findIndex(index, ")&|")
								if validateOperand != nil {
									err := validateOperand(v[index:endIndex])
									if err != nil {
										return err
									}
								}
								index = endIndex
							} else {
								parenthesesStack := make([]int, 0)
								parenthesesStack = append(parenthesesStack, index)
								for {
									index++
									nextOperandIndex := strings.IndexAny(v[index:], "()")
									if nextOperandIndex == -1 {
										return fmt.Errorf("syntax error: invalid expression")
									}
									index += nextOperandIndex
									if v[index] == '(' {
										parenthesesStack = append(parenthesesStack, index)
									} else {
										parenthesesStack = parenthesesStack[:len(parenthesesStack)-1]
										if len(parenthesesStack) == 0 {
											break
										}
									}
								}
								index++
							}
							operation = byte('n')
						}
					}
					return nil
				}
				if v[index] == '&' {
					err := validateOperator('&')
					if err != nil {
						return fail(err)
					}
				} else if v[index] == '|' {
					err := validateOperator('|')
					if err != nil {
						return fail(err)
					}
				} else {
					return fail(fmt.Errorf("syntax error: invalid operator '%c'", v[index]))
				}
				if operation == 'n' {
					goto NextOperator
				}
			} else if index < length {
				return fail(fmt.Errorf("syntax error: invalid expression"))
			}
		}

		if operation != byte('n') {
			return fail(fmt.Errorf("invalid end of expression"))
		}
		return result, nil
	}
	result, err := evaluate()
	if err != nil {
		return fail(err)
	}

	if len(parenthesesStack) != 0 {
		return fail(fmt.Errorf("invalid end of expression"))
	}
	return result, nil
}
