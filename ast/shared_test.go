package ast

func astLines() []string {
	return []string{
		"SelectWithUnionQuery (children 1)",
		" ExpressionList (children 1)",
		"  SelectQuery (children 2)",
		"   ExpressionList (children 1)",
		"    Asterisk",
		"    Identifier z (alias n)",
		"    Function z (alias r) (children 1)",
		"     ExpressionList",
		"   TablesInSelectQuery (children 1)",
		"    TablesInSelectQueryElement (children 1)",
		"     TableExpression (children 1)",
		"      TableIdentifier my_table_or_view",
	}
}

func createFunctionAstLines() []string {
	return []string{
		"CreateFunctionQuery z (children 2)",
		" Identifier z",
		" Function lambda (children 1)",
		"  ExpressionList (children 2)",
		"   Function tuple (children 1)",
		"    ExpressionList (children 1)",
		"     Identifier z",
		"   Literal Bool_1",
	}
}

func createQueryAstLines() []string {
	return []string{
		"CreateQuery  my_table_or_view (children 2)",
		" Identifier my_table_or_view",
		" SelectWithUnionQuery (children 1)",
		"  ExpressionList (children 1)",
		"   SelectQuery (children 2)",
		"    ExpressionList (children 2)",
		"     Asterisk",
		"     Literal ' '",
		"     Literal 'a literal with spaces'",
		"     Literal Array_['an', 'array', 'literal']",
		"    TablesInSelectQuery (children 1)",
		"     TablesInSelectQueryElement (children 1)",
		"      TableExpression (children 1)",
		"       TableIdentifier z",
	}
}

func queryWithTupleLines() []string {
	return []string{
		"SelectWithUnionQuery (children 1)",
		" ExpressionList (children 1)",
		"  SelectQuery (children 1)",
		"   ExpressionList (children 1)",
		"    Function describe (children 1)",
		"     ExpressionList (children 2)",
		"      Literal 'thing'",
		"      Function arrayMap (children 1)",
		"       ExpressionList (children 2)",
		"        Function lambda (children 1)",
		"         ExpressionList (children 2)",
		"          Function tuple (children 1)",
		"           ExpressionList (children 3)",
		"            Identifier z",
		"            Identifier a",
		"            Identifier b",
		"          Function expect (children 1)",
		"           ExpressionList (children 3)",
		"            Function format (children 1)",
		"             ExpressionList (children 4)",
		"              Literal '{} and {} have {}'",
		"              Identifier a",
		"              Identifier b",
		"              Function toString (children 1)",
		"               ExpressionList (children 1)",
		"                Identifier tokens",
		"            Identifier tokens",
		"            Function funky (children 1)",
		"             ExpressionList (children 2)",
		"              Identifier a",
		"              Identifier b",
		"        Function array (children 1)",
		"         ExpressionList (children 3)",
		"          Literal Tuple_(UInt64_3, 'abc', 'abc')",
		"          Literal Tuple_(UInt64_5, 'maximus', 'jebediah')",
		"          Literal Tuple_(UInt64_3, 'maximus', 'asdf')",
	}
}
