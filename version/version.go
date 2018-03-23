package version

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ki4jnq/forge"
)

var (
	cmd   *forge.Cmd
	flags = flag.NewFlagSet("version", flag.ExitOnError)
	conf  = &Config{
		Format:   "%d.%d.%d",
		FileName: "VERSION",
	}
)

type Config struct {
	Format   string
	FileName string `yaml:"file"`

	MajInc bool
	MinInc bool
	PatInc bool
	Save   bool
}

type ioStream struct {
	io.Reader
	io.Writer
}

func init() {
	flags.StringVar(&conf.Format, "format", conf.Format, "The format of the version.")
	flags.StringVar(&conf.FileName, "f", conf.FileName, "Read the version from the specified file, '-' reads stdin.")

	flags.BoolVar(&conf.MajInc, "Mi", false, "Increment the major version.")
	flags.BoolVar(&conf.MinInc, "mi", false, "Increment the minor version.")
	flags.BoolVar(&conf.PatInc, "pi", false, "Increment the patch version.")
	flags.BoolVar(&conf.Save, "s", false, "Write the new version to the version file.")

	cmd = &forge.Cmd{
		Name:      "version",
		Flags:     flags,
		SubConf:   conf,
		SubRunner: run,
	}
	forge.Register(cmd)
}

func run() error {
	ver := make([]int, 3)

	ioStream, finalizer, err := conf.getDataStream()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer finalizer()

	if _, err := fmt.Fscanf(ioStream, conf.Format, &ver[0], &ver[1], &ver[2]); err != nil {
		fmt.Printf("Error scanning version number, bad format: %v\n", err)
		return err
	}

	newVer := conf.incrementAll(ver[0], ver[1], ver[2])
	verStr := fmt.Sprintf(conf.Format, newVer[0], newVer[1], newVer[2])

	fmt.Println(verStr)
	if conf.Save {
		if _, err := fmt.Fprintf(ioStream, verStr); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) getDataStream() (io.ReadWriter, func(), error) {
	switch c.FileName {
	case "-":
		return ioStream{os.Stdin, os.Stdout}, func() {}, nil
	default:
		rd, _ := os.Open(c.FileName)
		wr, err := os.OpenFile(c.FileName, os.O_WRONLY, 0)
		return ioStream{rd, wr}, func() { wr.Sync(); rd.Close(); wr.Close() }, err
	}
}

func (c *Config) incrementAll(maj, min, pat int) []int {
	if c.MajInc {
		maj += 1
		min = 0
		pat = 0
	} else if c.MinInc {
		min += 1
		pat = 0
	} else if c.PatInc {
		pat += 1
	}
	return []int{maj, min, pat}
}
