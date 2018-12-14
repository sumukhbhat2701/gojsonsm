// Copyright 2018-2019 Couchbase, Inc. All rights reserved.

package gojsonsm

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterExpressionParser(t *testing.T) {
	assert := assert.New(t)
	parser, fe, err := NewFilterExpressionParser("`field` = TRUE")
	assert.Nil(err)
	assert.NotNil(parser)
	expr, err := fe.OutputExpression()
	assert.Nil(err)
	var trans Transformer
	matchDef := trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)

	err = parser.ParseString("TRUE OR FALSE AND NOT FALSE", fe)
	assert.Nil(err)
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(1, len(fe.AndConditions[0].OrConditions))
	assert.Equal(2, len(fe.AndConditions[1].OrConditions))
	assert.NotNil(fe.AndConditions[1].OrConditions[1].Not)
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("((TRUE OR FALSE))", fe)
	assert.Nil(err)
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(2, len(fe.AndConditions[0].OpenParens))
	assert.Equal(2, len(fe.AndConditions[1].CloseParens))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE AND FALSE)", fe)
	assert.Nil(err)
	assert.Equal(1, len(fe.AndConditions))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE OR FALSE) AND (FALSE OR TRUE)", fe)
	assert.Nil(err)
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(1, len(fe.SubFilterExpr))
	assert.Equal(2, len(fe.SubFilterExpr[0].AndConditions))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE OR FALSE) AND (FALSE OR TRUE) AND TRUE", fe)
	assert.Nil(err)
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(1, len(fe.SubFilterExpr))
	assert.Equal(2, len(fe.SubFilterExpr[0].AndConditions))
	assert.Equal(1, len(fe.SubFilterExpr[0].SubFilterExpr))
	assert.Equal(1, len(fe.SubFilterExpr[0].SubFilterExpr[0].AndConditions))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE AND FALSE) OR (FALSE AND TRUE)", fe)
	assert.Nil(err)
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(2, len(fe.AndConditions[0].OrConditions))
	assert.Equal(2, len(fe.AndConditions[1].OrConditions))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("((TRUE OR FALSE)) OR (TRUE)", fe)
	assert.Nil(err)
	assert.Equal(3, len(fe.AndConditions))
	assert.Equal(0, len(fe.SubFilterExpr))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE AND FALSE) OR (FALSE)", fe)
	assert.Nil(err)
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(2, len(fe.AndConditions[0].OrConditions))
	assert.Equal(1, len(fe.AndConditions[1].OrConditions))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("TRUE AND (TRUE OR FALSE) AND FALSE", fe)
	assert.Nil(err)
	assert.Equal(1, len(fe.AndConditions))
	assert.Equal(2, len(fe.AndConditions[0].OrConditions))
	assert.Nil(fe.AndConditions[0].OrConditions[0].SubExpr) // TRUE (AND...)
	assert.NotNil(fe.AndConditions[0].OrConditions[1].SubExpr)
	assert.Equal(2, len(fe.AndConditions[0].OrConditions[1].SubExpr.AndConditions))                 // (TRUE OR FALSE) AND FALSE
	assert.Equal(1, len(fe.AndConditions[0].OrConditions[1].SubExpr.AndConditions[0].OrConditions)) // (TRUE OR FALSE)
	assert.Equal(1, len(fe.AndConditions[0].OrConditions[1].SubExpr.AndConditions[1].OrConditions)) //  FALSE
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE OR FALSE) AND (FALSE OR TRUE)", fe)
	assert.Nil(err)
	assert.Equal(1, len(fe.AndConditions[0].OrConditions))
	assert.Equal(1, len(fe.AndConditions[1].OrConditions))
	assert.Equal(1, len(fe.SubFilterExpr))
	assert.Equal(2, len(fe.AndConditions))
	assert.Equal(2, len(fe.SubFilterExpr[0].AndConditions))
	assert.Equal(1, len(fe.SubFilterExpr[0].AndConditions[0].OrConditions))
	assert.Equal(1, len(fe.SubFilterExpr[0].AndConditions[1].OrConditions))
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("NOT NOT NOT TRUE", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Nil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path >= field2", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsGreaterThanOrEqualTo())
	assert.False(fe.AndConditions[0].OrConditions[0].Operand.Op.IsGreaterThan())
	assert.Equal("field2", fe.AndConditions[0].OrConditions[0].Operand.RHS.Field.String())

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path IS NOT NULL", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.CheckOp.IsNotNull())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m := NewFastMatcher(matchDef)
	userData := map[string]interface{}{
		"fieldpath": map[string]interface{}{
			"path": 0,
		},
	}
	udMarsh, _ := json.Marshal(userData)
	match, err := m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path = \"value\"", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsEqual())
	assert.Equal("value", fe.AndConditions[0].OrConditions[0].Operand.RHS.Value.String())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"fieldpath": map[string]interface{}{
			"path": "value",
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("`onePath.Only` < field2", fe)
	assert.Nil(err)
	assert.Equal("onePath.Only", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("field2", fe.AndConditions[0].OrConditions[0].Operand.RHS.Field.String())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"onePath.Only": -2,
		"field2":       2,
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("onePath.field1 < onePath.field2", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"onePath": map[string]interface{}{
			"field1": -2,
			"field2": 2,
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("`onePath.Only` <> \"value\" OR `onePath.Only` <> \"value2\"", fe)
	assert.Nil(err)
	assert.Equal("onePath.Only", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsNotEqual())
	assert.Equal("value", fe.AndConditions[0].OrConditions[0].Operand.RHS.Value.String())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"onePath.Only": -2,
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)
	userData = map[string]interface{}{}
	udMarsh, _ = json.Marshal(userData)
	m.Reset()
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("onePath.field1 <> onePath.field2", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("META().`onePath.Only` = \"value\"", fe)
	assert.Nil(err)
	assert.Equal("META()", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("onePath.Only", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsEqual())
	assert.Equal("value", fe.AndConditions[0].OrConditions[0].Operand.RHS.Value.String())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"META()": map[string]interface{}{
			"onePath.Only": "value",
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("`[$%XDCRInternalMeta*%$]`.metaKey = \"value\"", fe)
	assert.Equal("metaKey", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsEqual())
	assert.Equal("value", fe.AndConditions[0].OrConditions[0].Operand.RHS.Value.String())
	err = parser.ParseString("`[$%XDCRInternalMeta*%$]`.metaKey EXISTS AND `[$%XDCRInternalMeta*%$]`.metaKey = \"value\"", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"[$%XDCRInternalMeta*%$]": map[string]interface{}{
			"metaKey": "value",
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	// path name with leading number must be escaped - TODO this should be documented
	// We're not supporting neg index as of now
	fe = &FilterExpression{}
	err = parser.ParseString("`2DarrayPath`[1][-2] = fieldpath2.path2", fe)
	assert.NotNil(err)
	//	assert.Equal("2DarrayPath [1] [-2]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	//	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsEqual())

	fe = &FilterExpression{}
	err = parser.ParseString("`1DarrayPath`[1] = \"arrayVal1\"", fe)
	assert.Nil(err)
	assert.Equal("1DarrayPath [1]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsEqual())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{"1DarrayPath": [2]string{"arrayVal0", "arrayVal1"}}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	// No negative indexes for now
	fe = &FilterExpression{}
	err = parser.ParseString("arrayPath[1].path2.arrayPath3[-10].`multiword array`[20] = fieldpath2.path2", fe)
	assert.NotNil(err)
	//	assert.Equal("arrayPath [1]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	//	assert.Equal("arrayPath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].StrValue)
	//	assert.Equal("path2", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].StrValue)
	//	assert.Equal(0, len(fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].ArrayIndexes))
	//	assert.Equal("arrayPath3 [-10]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[2].String())
	//	assert.Equal("multiword array [20]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[3].String())

	fe = &FilterExpression{}
	err = parser.ParseString("arrayPath[1].path2.arrayPath3[10].`multiword array`[20] = fieldpath2.path2", fe)
	assert.Nil(err)
	assert.Equal("arrayPath [1]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("arrayPath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].StrValue)
	assert.Equal("path2", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].StrValue)
	assert.Equal(0, len(fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].ArrayIndexes))
	assert.Equal("arrayPath3 [10]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[2].String())
	assert.Equal("multiword array [20]", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[3].String())

	fe = &FilterExpression{}
	err = parser.ParseString("key < PI()", fe)
	assert.Nil(err)
	assert.Equal("key", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.True(*fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncNoArg.ConstFuncNoArgName.Pi)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{"key": 3.14}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path <= ABS(5)", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsLessThanOrEqualTo())
	assert.Equal("ABS", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.ConstFuncOneArgName.String())
	assert.Equal("5", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.String())
	assert.Nil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.SubFunc)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"fieldpath": map[string]interface{}{
			"path": -2,
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("DATE(fieldpath.path) = DATE(\"2019-01-01\")", fe)
	assert.Nil(err)
	assert.Equal("DATE", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.ConstFuncOneArgName.String())
	assert.Equal("2019-01-01", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.String())
	assert.Nil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.SubFunc)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"fieldpath": map[string]interface{}{
			"path": "2019-01-01",
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path = DATE(`field with spaces`)", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.Equal("DATE", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.ConstFuncOneArgName.String())
	assert.Equal("field with spaces", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.Field.String())
	assert.Nil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.SubFunc)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path >= ABS(CEIL(PI()))", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.Equal("ABS", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.ConstFuncOneArgName.String())
	assert.Nil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.Argument)
	assert.NotNil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.SubFunc)
	assert.NotNil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.SubFunc.ConstFuncOneArg)
	assert.NotNil(fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncOneArg.Argument.SubFunc.ConstFuncOneArg.Argument.SubFunc.ConstFuncNoArg)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"fieldpath": map[string]interface{}{
			"path": 10,
		},
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path <> POW(ABS(CEIL(PI())),2)", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.True(fe.AndConditions[0].OrConditions[0].Operand.Op.IsNotEqual())
	assert.Equal("POW", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncTwoArgs.ConstFuncTwoArgsName.String())
	assert.Equal("ABS", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncTwoArgs.Argument0.SubFunc.ConstFuncOneArg.ConstFuncOneArgName.String())

	fe = &FilterExpression{}
	err = parser.ParseString("REGEXP_CONTAINS(`[$%XDCRInternalKey*%$]`, \"^xyz*\")", fe)
	assert.Nil(err)
	assert.Equal("REGEXP_CONTAINS", fe.AndConditions[0].OrConditions[0].Operand.BooleanExpr.BooleanFunc.BooleanFuncTwoArgs.BooleanFuncTwoArgsName.String())
	assert.Equal("[$%XDCRInternalKey*%$]", fe.AndConditions[0].OrConditions[0].Operand.BooleanExpr.BooleanFunc.BooleanFuncTwoArgs.Argument0.Field.String())
	assert.Equal("^xyz*", fe.AndConditions[0].OrConditions[0].Operand.BooleanExpr.BooleanFunc.BooleanFuncTwoArgs.Argument1.Argument.String())
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)
	userData = map[string]interface{}{
		"[$%XDCRInternalKey*%$]": "xyzzzzz",
	}
	udMarsh, _ = json.Marshal(userData)
	match, err = m.Match(udMarsh)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("Testdoc = true AND REGEXP_CONTAINS(`[$%XDCRInternalKey*%$]`, \"^abc\")", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Nil(err)
	matchDef = trans.Transform([]Expression{expr})
	assert.NotNil(matchDef)
	m = NewFastMatcher(matchDef)

	testRaw := json.RawMessage(`{"Testdoc": true}`)
	testData, err := testRaw.MarshalJSON()
	tempMap := make(map[string]interface{})
	err = json.Unmarshal(testData, &tempMap)
	tempMap["[$%XDCRInternalKey*%$]"] = "abcdef"
	testData2, err := json.Marshal(tempMap)
	match, err = m.Match(testData2)
	assert.True(match)

	fe = &FilterExpression{}
	err = parser.ParseString("fieldpath.path = POW(ABS(CEIL(PI())),2) AND REGEXP_CONTAINS(fieldPath2, \"^abc*$\")", fe)
	assert.Nil(err)
	assert.Equal("fieldpath", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[0].String())
	assert.Equal("path", fe.AndConditions[0].OrConditions[0].Operand.LHS.Field.Path[1].String())
	assert.Equal("POW", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncTwoArgs.ConstFuncTwoArgsName.String())
	assert.Equal("ABS", fe.AndConditions[0].OrConditions[0].Operand.RHS.Func.ConstFuncTwoArgs.Argument0.SubFunc.ConstFuncOneArg.ConstFuncOneArgName.String())
	assert.Equal(1, len(fe.AndConditions))
	assert.Equal(2, len(fe.AndConditions[0].OrConditions))
	assert.NotNil(fe.AndConditions[0].OrConditions[1].Operand.BooleanExpr)
	assert.Equal("fieldPath2", fe.AndConditions[0].OrConditions[1].Operand.BooleanExpr.BooleanFunc.BooleanFuncTwoArgs.Argument0.String())
	assert.Equal("^abc*$", fe.AndConditions[0].OrConditions[1].Operand.BooleanExpr.BooleanFunc.BooleanFuncTwoArgs.Argument1.Argument.String())

	var testStr string = "`field.Path` = \"value\""
	_, err = GetFilterExpressionMatcher(testStr)
	assert.Nil(err)

	// Negative
	_, _, err = NewFilterExpressionParser("fieldpath.`path = fieldPath2")
	assert.NotNil(err)

	fe = &FilterExpression{}
	err = parser.ParseString("(TRUE) OR FALSE)", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Equal(ErrorMalformedParenthesis, err)

	fe = &FilterExpression{}
	err = parser.ParseString("(((TRUE) OR FALSE) OR FALSE))", fe)
	assert.Nil(err)
	expr, err = fe.OutputExpression()
	assert.Equal(ErrorMalformedParenthesis, err)

}