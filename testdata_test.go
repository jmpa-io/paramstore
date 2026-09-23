package paramstore

// validTestdata represents a bunch of valid Parameters, used for testing.
var validTestdata = testParameters{
	{
		Parameter{
			Name:  "/hello",
			Value: "this is (possibly) hidden",
			Type:  ParameterTypeSecureString,
		},
	},
	{
		Parameter{
			Name:  "/world",
			Value: "this is plain text",
			Type:  ParameterTypeString,
		},
	},
	{
		Parameter{
			Name:  "/test",
			Value: "this,is,a,comma,list",
			Type:  ParameterTypeStringList,
		},
	},
}
