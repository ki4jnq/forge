package db

import (
	"fmt"
	"os"
	"os/exec"
)

// executer manages running SQL files.
type Executer struct{}

func (r Executer) ExecSQLFile(filename string, conn DBConn) error {
	fmt.Println("going to execute filename with psql")
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	args := []string{}
	argMap := map[string]string{
		"-f": filename,
		"-d": conn.DBName,
		"-U": conn.DBUser,
		"-W": conn.DBPass,
		"-h": conn.DBHost,
		"-p": conn.DBPort,
	}

	for arg, val := range argMap {
		if val != "" {
			args = append(args, arg, val)
		}
	}

	fmt.Println("Executing psql")
	cmd := exec.Command("psql", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = file

	if err = cmd.Run(); err != nil {
		fmt.Println("Encountered an error executing psql")
		return err
	}
	return nil
}
