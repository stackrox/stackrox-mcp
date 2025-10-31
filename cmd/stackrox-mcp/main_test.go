package main

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogging(t *testing.T) {
	setupLogging()
	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
}
