package infrastructure

import (
	"Go-Service/src/main/infrastructure/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateViewerIDFromIP_SameIPSameResult(t *testing.T) {
	id1 := util.GenerateViewerIDFromIP("1.2.3.4", "secret")
	id2 := util.GenerateViewerIDFromIP("1.2.3.4", "secret")
	assert.Equal(t, id1, id2)
}

func TestGenerateViewerIDFromIP_DifferentIPDifferentResult(t *testing.T) {
	id1 := util.GenerateViewerIDFromIP("1.2.3.4", "secret")
	id2 := util.GenerateViewerIDFromIP("5.6.7.8", "secret")
	assert.NotEqual(t, id1, id2)
}

func TestGenerateViewerIDFromIP_Format(t *testing.T) {
	id := util.GenerateViewerIDFromIP("192.168.1.1", "testsecret")
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, id)
}

func TestGenerateViewerIDFromIP_DifferentSecret(t *testing.T) {
	id1 := util.GenerateViewerIDFromIP("1.2.3.4", "secret1")
	id2 := util.GenerateViewerIDFromIP("1.2.3.4", "secret2")
	assert.NotEqual(t, id1, id2)
}
