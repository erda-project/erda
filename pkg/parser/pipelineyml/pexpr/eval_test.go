package pexpr

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	params := map[string]string{
		"outputs.git-checkout.commit": "69511e5bfd4d0465efa9101190de6f5f8cf48f97",
	}
	result, err := Eval("'${{ outputs.git-checkout.commit }}' != ''", params)
	assert.NoError(t, err)
	assert.True(t, result.(bool))
	spew.Dump(result)

	result, err = Eval("pipeline_status == 'Failed'", map[string]string{"pipeline_status": "Failed"})
	assert.NoError(t, err)
	spew.Dump(result)

	result, err = Eval("${{ configs.${{ configs.key1 }} }} == 123", map[string]string{
		"configs.key1": "key2",
		"configs.key2": "123",
	})
	assert.NoError(t, err)
	spew.Dump(result)

	result, err = Eval("${{ configs.${{ configs.key1 }} }} ${{ configs.key3 }}== 123", map[string]string{
		"configs.key2": "123",
	})
	fmt.Println(err)
	assert.Error(t, err)

	result, err = Eval("${{  configs.key }}", nil)
	fmt.Println(err)
	assert.Error(t, err)
}
