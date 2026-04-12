package id

import (
	"errors"
	"testing"
)

func TestUnifyMachineID(t *testing.T) {
	tests := []struct {
		name    string
		mockID  string
		mockErr error
		wantLen int
		wantErr bool
	}{
		{
			name:    "标准GUID格式",
			mockID:  "{550e8400-e29b-41d4-a716-446655440000}",
			wantLen: 32,
			wantErr: false,
		},
		{
			name:    "短ID自动哈希降级",
			mockID:  "short-id",
			wantLen: 32, // 降级后仍为32位
			wantErr: false,
		},
		{
			name:    "获取失败",
			mockErr: errors.New("system error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFetcher := func() (string, error) {
				return tt.mockID, tt.mockErr
			}
			got, err := unifyMachineIDInternal(mockFetcher)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("len(got) = %d, want %d", len(got), tt.wantLen)
			}
			// 验证结果始终是32位小写hex
			if !tt.wantErr && !isValidHex32(got) {
				t.Errorf("result not valid 32-char hex: %s", got)
			}
		})
	}
}

func isValidHex32(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func TestFormatMachineID(t *testing.T) {
	tests := []struct {
		name       string
		mockID     string
		mockErr    error
		opt        FormatOption
		wantLen    int
		wantHyphen bool
		wantUpper  bool
		wantErr    bool
	}{
		{
			name:       "小写无横线",
			mockID:     "550e8400e29b41d4a716446655440000",
			opt:        FormatOption{UpperCase: false, WithHyphen: false},
			wantLen:    32,
			wantHyphen: false,
			wantUpper:  false,
			wantErr:    false,
		},
		{
			name:       "大写无横线",
			mockID:     "550e8400e29b41d4a716446655440000",
			opt:        FormatOption{UpperCase: true, WithHyphen: false},
			wantLen:    32,
			wantHyphen: false,
			wantUpper:  true,
			wantErr:    false,
		},
		{
			name:       "小写带横线",
			mockID:     "550e8400e29b41d4a716446655440000",
			opt:        FormatOption{UpperCase: false, WithHyphen: true},
			wantLen:    36,
			wantHyphen: true,
			wantUpper:  false,
			wantErr:    false,
		},
		{
			name:       "大写带横线",
			mockID:     "550e8400e29b41d4a716446655440000",
			opt:        FormatOption{UpperCase: true, WithHyphen: true},
			wantLen:    36,
			wantHyphen: true,
			wantUpper:  true,
			wantErr:    false,
		},
		{
			name:       "降级哈希小写无横线",
			mockID:     "short-id",
			opt:        FormatOption{UpperCase: false, WithHyphen: false},
			wantLen:    32,
			wantHyphen: false,
			wantUpper:  false,
			wantErr:    false,
		},
		{
			name:       "降级哈希大写带横线",
			mockID:     "short-id",
			opt:        FormatOption{UpperCase: true, WithHyphen: true},
			wantLen:    36,
			wantHyphen: true,
			wantUpper:  true,
			wantErr:    false,
		},
		{
			name:    "获取失败",
			mockErr: errors.New("fail"),
			opt:     FormatOption{UpperCase: false, WithHyphen: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFetcher := func() (string, error) {
				return tt.mockID, tt.mockErr
			}
			got, err := formatMachineIDWithFetcher(tt.opt, mockFetcher)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(got) != tt.wantLen {
					t.Errorf("len(got) = %d, want %d", len(got), tt.wantLen)
				}
				if tt.wantHyphen && (len(got) == 36 && (got[8] != '-' || got[13] != '-' || got[18] != '-' || got[23] != '-')) {
					t.Errorf("hyphen not in expected positions: %s", got)
				}
				if tt.wantUpper && hasLowerHex(got) {
					t.Errorf("should be upper case: %s", got)
				}
				if !tt.wantUpper && hasUpperHex(got) {
					t.Errorf("should be lower case: %s", got)
				}
			}
		})
	}
}

func hasUpperHex(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'F' {
			return true
		}
	}
	return false
}

func hasLowerHex(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'f' {
			return true
		}
	}
	return false
}
