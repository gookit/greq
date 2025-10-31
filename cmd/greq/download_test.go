package main

import (
	"testing"
	"time"

	"github.com/gookit/goutil/testutil/assert"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		assert.Eq(t, tt.expected, result)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{30 * time.Second, "00:30"},
		{90 * time.Second, "01:30"},
		{3661 * time.Second, "01:01:01"},
		{7200 * time.Second, "02:00:00"},
		{-1 * time.Second, "00:00"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.input)
		assert.Eq(t, tt.expected, result)
	}
}

func TestShowDownloadProgress(t *testing.T) {
	// 测试有文件大小的情况
	startTime := time.Now().Add(-30 * time.Second)
	showDownloadProgress(52428800, 104857600, startTime)
	
	// 测试未知文件大小的情况
	startTime2 := time.Now().Add(-10 * time.Second)
	showDownloadProgress(10485760, 0, startTime2)
}