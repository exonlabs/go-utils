// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package comm

import (
	"github.com/exonlabs/go-utils/pkg/logging"
)

// LogMsg logs general communication messages.
func LogMsg(log *logging.Logger, msg string, args ...any) {
	if log != nil && msg != "" {
		log.Info(msg, args...)
	}
}

// LogTx logs communication transmitted data in formatted hexadecimal.
//
// Example
//
//	2006-01-02 15:04:05.000000 TX >> 0102030405060708090A0B0C0D0E0F
func LogTx(log *logging.Logger, data []byte, addr any) {
	if log != nil && len(data) > 0 {
		if addr != nil {
			log.Info("(%s) TX >> %X", addr, data)
		} else {
			log.Info("TX >> %X", data)
		}
	}
}

// LogRx logs communication received data in formatted hexadecimal.
//
// Example
//
//	2006-01-02 15:04:05.000000 RX << 0102030405060708090A0B0C0D0E0F
func LogRx(log *logging.Logger, data []byte, addr any) {
	if log != nil && len(data) > 0 {
		if addr != nil {
			log.Info("(%s) RX << %X", addr, data)
		} else {
			log.Info("RX << %X", data)
		}
	}
}
