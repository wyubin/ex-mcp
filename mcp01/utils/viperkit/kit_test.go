package viperkit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	_, filename, _, _ = runtime.Caller(0)
	dirTest, _        = filepath.Abs(filepath.Join(filepath.Dir(filename), "test_files"))
)

func TestReaderEnv(t *testing.T) {
	rawCfg, err := os.Open(filepath.Join(dirTest, "env_default"))
	assert.NoError(t, err)
	ReaderEnv(rawCfg)
	assert.Equal(t, "1.0.0", viper.GetString("version"))
}
