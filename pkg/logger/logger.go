package logger

import (
    "zbxtable/pkg/utils"
)

// Log 鍏ㄥ眬鏃ュ織瀹炰緥 (鍏煎鏃т唬鐮?
var Log = utils.Log

// InitLogger 鍒濆鍖栨棩蹇?
func InitLogger(logPath string, level, maxday, maxlines, maxsize int, daily bool) error {
    return utils.InitLogger(logPath, level, maxday, maxlines, maxsize, daily)
}
