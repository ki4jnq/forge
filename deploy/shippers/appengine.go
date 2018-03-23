package shippers

import (
	"context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
)

type AppEngine struct {
	tmpAppEngineConfig string

	version string
	image   string
	appYaml map[interface{}]interface{}
}

func NewAppEngineShipper(opts map[string]interface{}, args CmdArgs) *AppEngine {
	image, _ := opts["image"].(string)

	appYaml, ok := opts["gcloud"].(map[interface{}]interface{})
	if !ok {
		appYaml = make(map[interface{}]interface{})
	}

	ae := &AppEngine{
		version: args.AppEngine.ImageTag,
		image:   image,
		appYaml: appYaml,
	}

	return ae
}

func (ae *AppEngine) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	go func() {
		run := func(f func(context.Context) error) {
			if e := f(ctx); e != nil {
				panic(e)
			}
		}

		defer close(ch)
		defer failSafe("AppEngine", ch)
		defer run(ae.cleanup)

		run(ae.generateAppYaml)
		run(ae.deploy)
	}()
	return ch
}

// No need to explicitly do anything here.
func (ae *AppEngine) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (ae *AppEngine) generateAppYaml(_ context.Context) error {
	yamlBody, err := yaml.Marshal(&ae.appYaml)
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile("", "app-engine-config")
	if err != nil {
		return err
	}
	defer file.Close()
	ae.tmpAppEngineConfig = file.Name()

	if _, err := file.Write(yamlBody); err != nil {
		return err
	}

	if err := os.Symlink(ae.tmpAppEngineConfig, "app.yaml"); err != nil {
		return err
	}

	return nil
}

func (ae *AppEngine) deploy(ctx context.Context) error {
	// NOTE: Eventually we could use exec.CommandContext so that if the
	// context is canceled the build will be automatically halted.
	cmdArgs := []string{"app", "deploy", "--quiet"}
	if ae.image != "" && ae.version != "" {
		cmdArgs = append(cmdArgs, "--image-url", ae.image+":"+ae.version)
	} else if ae.image != "" {
		cmdArgs = append(cmdArgs, "--image-url", ae.image)
	}

	cmd := exec.CommandContext(ctx, "gcloud", cmdArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (ae *AppEngine) cleanup(_ context.Context) error {
	if err := os.Remove(ae.tmpAppEngineConfig); err != nil {
		return err
	}

	if err := os.Remove("app.yaml"); err != nil {
		return err
	}
	return nil
}
