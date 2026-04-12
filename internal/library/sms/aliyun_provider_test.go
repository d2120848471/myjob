package sms

import (
	"testing"

	modelruntime "myjob/internal/model/runtime"

	"github.com/stretchr/testify/require"
)

func TestNewAliyunSMSClient_ValidateCredentials(t *testing.T) {
	t.Parallel()

	_, err := newAliyunSMSClient(modelruntime.SMSConfig{})
	require.ErrorContains(t, err, "AccessKey")

	_, err = newAliyunSMSClient(modelruntime.SMSConfig{AccessKey: "ak"})
	require.ErrorContains(t, err, "AccessKey Secret")
}
