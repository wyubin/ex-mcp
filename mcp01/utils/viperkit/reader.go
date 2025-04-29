package viperkit

import (
	"bufio"
	"io"
	"strings"

	"github.com/spf13/viper"
)

func ReaderEnv(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		infoPair := strings.Split(scanner.Text(), "=")
		if len(infoPair) != 2 {
			continue
		}
		viper.SetDefault(infoPair[0], infoPair[1])
	}
}
