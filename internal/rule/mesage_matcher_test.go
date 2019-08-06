package rule_test

import (
	"testing"
	"time"

	"github.com/mylxsw/adanos-alert/internal/repository"
	"github.com/mylxsw/adanos-alert/internal/rule"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type messageMatcherTestCase struct {
	Rule    string
	Matched bool
}

func TestMessageMatcher_Match(t *testing.T) {

	var msg = repository.Message{
		ID:      primitive.NewObjectID(),
		Content: `{"log_level": "debug", "message": "request", "context": {"user_id": 123}}`,
		Meta: repository.MessageMeta{
			"environment": "dev",
			"server":      "192.168.1.1",
		},
		Tags:      []string{"php", "nodejs"},
		Origin:    "Filebeat",
		CreatedAt: time.Now(),
	}

	var testcases = []messageMatcherTestCase{
		{Rule: `"php" in Tags`, Matched: true},
		{Rule: `"java" in Tags`, Matched: false},
		{Rule: `"nodejs" in Tags or "java" in Tags`, Matched: true},
		{Rule: `"java" not in Tags`, Matched: true},
		{Rule: `Meta["server"] == "192.168.1.1"`, Matched: true},
		{Rule: `Meta["server"] == "192.168.1.2"`, Matched: false},
		{Rule: `Meta["environment"] != "production"`, Matched: true},
		{Rule: `Meta["environment"] in ["dev", "test"]`, Matched: true},
		{Rule: `Meta["environment"] not in ["production", "test"]`, Matched: true},
		{Rule: `Content matches "\"request\""`, Matched: true},
		{Rule: `JsonGet("context.user_id", "0") == "123"`, Matched: true},
		{Rule: `JsonGet("context.enterprise_id", "0") == "0"`, Matched: true},
		{Rule: `Content startsWith "{"`, Matched: true},
		{Rule: `Content endsWith "XX"`, Matched: false},
		{Rule: `Upper(Meta["environment"]) == "DEV"`, Matched: true},
		{Rule: `Lower(Origin) == "filebeat"`, Matched: true},
	}

	for _, tc := range testcases {
		matcher, err := rule.NewMessageMatcher(repository.Rule{Rule: tc.Rule,})
		assert.NoError(t, err)
		matched, err := matcher.Match(msg)
		assert.NoError(t, err)
		assert.Equal(t, tc.Matched, matched)

		assert.Equal(t, tc.Rule, matcher.Rule().Rule)
	}

	_, err := rule.NewMessageMatcher(repository.Rule{Rule: `xxxxxxx`})
	assert.Error(t, err)
}
