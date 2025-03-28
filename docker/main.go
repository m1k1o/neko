/*
This program processes a Dockerfile. When it encounters a FROM command with a relative path,
it pastes the content of the referenced Dockerfile into the current Dockerfile with some modifications:
  - It ensures that all ADD and COPY commands point to the correct context path by adding the relative path
    to the first part of the command (the file or directory being copied).
  - It takes the ARG variables defined before the FROM command and prepends them with the alias of the
    FROM command. It also replaces any occurrences of the ARG variables in the Dockerfile with the new prefixed
    variables. Then it writes them to the beginning of the new Dockerfile.

It allows to split large multi-stage Dockerfiles into own directories where they can be built independently. It also
allows to dynamically join these Dockerfiles into a single Dockerfile based on various conditions.
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	inputPath := flag.String("i", "", "Path to the input Dockerfile")
	outputPath := flag.String("o", "", "Path to the output Dockerfile")
	flag.Parse()

	if *inputPath == "" {
		log.Println("Usage: go run main.go -i <input Dockerfile> [-o <output Dockerfile>]")
		os.Exit(1)
	}

	buildcontext, err := ButidContextFromPath(*inputPath)
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = processDockerfile(buildcontext, *outputPath)
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// relativeDockerFile reads the Dockerfile, modifies it to point to the new context path, and returns the global ARGs
func relativeDockerFile(buf *bytes.Buffer, ctx BuildContext, newContextPath, alias string) (ArgCommand, error) {
	// read the Dockerfile
	file, err := os.Open(ctx.DockerfilePath())
	if err != nil {
		return nil, fmt.Errorf("failed to open Dockerfile: %w", err)
	}
	defer file.Close()

	// new context path relative to the current context path
	newContextPath, err = filepath.Rel(newContextPath, ctx.ContextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	// use argPrefix to prepend the alias to the ARG variables
	argPrefix := strings.ToUpper(alias) + "_"
	// replace - with _ in the alias
	argPrefix = strings.ReplaceAll(argPrefix, "-", "_")

	beforeFrom := true
	globalArgs := ArgCommand{}

	// read the Dockerfile line by line and modify it
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// handle ARG lines defined before FROM
		if !beforeFrom {
			line = globalArgs.ReplaceArgPrefix(argPrefix, line)
		}

		// we need to move the ARG lines before the FROM line
		if strings.HasPrefix(line, "ARG") && beforeFrom {
			args, err := ParseArgCommand(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ARG command: %w", err)
			}
			globalArgs = append(globalArgs, args...)
			continue
		}

		// modify FROM lines
		if strings.HasPrefix(line, "FROM") {
			// parse the FROM command
			cmd, err := ParseFromCommand(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse FROM command: %w", err)
			}

			// handle the case where ARGs are defined before FROM
			cmd.Image = globalArgs.ReplaceArgPrefix(argPrefix, cmd.Image)

			// add new alias if it is not already present
			if alias != "" {
				cmd.Alias = alias
			}

			beforeFrom = false
			buf.WriteString(cmd.String() + "\n")
			continue
		}

		// modify COPY and ADD lines
		if strings.HasPrefix(line, "COPY") || strings.HasPrefix(line, "ADD") {
			parts := strings.Fields(line)

			containsFrom := false
			localPathIndex := 0
			for i, part := range parts {
				if strings.HasPrefix(part, "--from=") {
					containsFrom = true
					continue
				}
				if strings.HasPrefix(part, "--") {
					continue
				}
				if localPathIndex == 0 && i > 0 {
					localPathIndex = i
				}
			}

			if !containsFrom {
				// replace the local part with the new context path
				parts[localPathIndex] = filepath.Join(newContextPath, parts[localPathIndex])
				newLine := strings.Join(parts, " ")
				buf.WriteString(newLine + "\n")
				continue
			}
		}

		// write the line as is
		buf.WriteString(line + "\n")
	}

	// add prefix to global ARGs
	for i := range globalArgs {
		if globalArgs[i].Key != "" {
			globalArgs[i].Key = argPrefix + globalArgs[i].Key
		}
	}

	return globalArgs, scanner.Err()
}

// processDockerfile processes the Dockerfile and resolves sub-Dockerfiles in it
func processDockerfile(ctx BuildContext, outputPath string) error {
	// read the Dockerfile
	file, err := os.Open(ctx.DockerfilePath())
	if err != nil {
		return fmt.Errorf("failed to open Dockerfile: %w", err)
	}
	defer file.Close()

	globalArgs := ArgCommand{}

	// read the Dockerfile line by line and modify it
	newDockerfile := bytes.NewBuffer(nil)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// modify FROM lines
		if strings.HasPrefix(line, "FROM ./") {
			// parse the FROM command
			cmd, err := ParseFromCommand(line)
			if err != nil {
				return fmt.Errorf("failed to parse FROM command: %w", err)
			}

			// resolve environment variables in the image name
			cmd.Image = os.ExpandEnv(cmd.Image)

			// create a new build context
			newBuildcontext, err := ButidContextFromPath(filepath.Join(ctx.ContextPath, cmd.Image))
			if err != nil {
				return fmt.Errorf("failed to get build context: %w", err)
			}

			// resolve the dockerfile content
			args, err := relativeDockerFile(newDockerfile, newBuildcontext, ctx.ContextPath, cmd.Alias)
			if err != nil {
				return fmt.Errorf("failed to get relative Dockerfile: %w", err)
			}
			globalArgs = append(globalArgs, args...)

			continue
		}

		// copy all other lines as is
		newDockerfile.WriteString(line + "\n")
	}

	// check for errors while reading the Dockerfile
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input Dockerfile: %w", err)
	}

	// add the global ARGs to the beginning of the new Dockerfile
	outBytes := append([]byte(globalArgs.MultiLineString()), newDockerfile.Bytes()...)

	if outputPath != "" {
		// write the new Dockerfile to the output path
		return os.WriteFile(outputPath, outBytes, 0644)
	}

	// write to stdout
	fmt.Print(string(outBytes))
	return nil
}

// BuildContext represents the build context for a Dockerfile
type BuildContext struct {
	ContextPath string
	Dockerfile  string // if empty, use the default Dockerfile name
}

func ButidContextFromPath(path string) (BuildContext, error) {
	// check if the path exists
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return BuildContext{}, fmt.Errorf("path does not exist: %s", path)
	}

	// check if the path is a directory
	if err == nil && fi.IsDir() {
		return BuildContext{
			ContextPath: path,
			Dockerfile:  "Dockerfile",
		}, nil
	}

	return BuildContext{
		ContextPath: filepath.Dir(path),
		Dockerfile:  filepath.Base(path),
	}, nil
}

func (bc *BuildContext) DockerfilePath() string {
	if bc.Dockerfile != "" {
		return filepath.Join(bc.ContextPath, bc.Dockerfile)
	}
	return filepath.Join(bc.ContextPath, "Dockerfile")
}

// FromCommand represents the FROM command in a Dockerfile
type FromCommand struct {
	Image    string
	Alias    string
	Platform string
}

func ParseFromCommand(line string) (fc FromCommand, err error) {
	parts := strings.Fields(line)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "from" {
		err = fmt.Errorf("invalid FROM line: %s", line)
		return
	}
	for i := 1; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "--platform=") {
			fc.Platform = strings.TrimPrefix(parts[i], "--platform=")
		}
		if strings.ToLower(parts[i]) == "as" && i+1 < len(parts) {
			fc.Alias = parts[i+1]
			break
		}
		fc.Image = parts[i]
	}
	return
}

func (fc *FromCommand) String() string {
	var sb strings.Builder
	sb.WriteString("FROM ")
	if fc.Platform != "" {
		sb.WriteString(fmt.Sprintf("--platform=%s ", fc.Platform))
	}
	sb.WriteString(fc.Image)
	if fc.Alias != "" {
		sb.WriteString(fmt.Sprintf(" AS %s", fc.Alias))
	}
	return sb.String()
}

// ArgCommand represents the ARG command in a Dockerfile
type Arg struct {
	Key   string
	Value string
}

type ArgCommand []Arg

func ParseArgCommand(line string) (ac ArgCommand, err error) {
	parts := strings.Fields(line)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "arg" {
		err = fmt.Errorf("invalid ARG line: %s", line)
		return
	}

	for i := 1; i < len(parts); i++ {
		if strings.Contains(parts[i], "=") {
			kv := strings.SplitN(parts[i], "=", 2)
			if len(kv) == 2 {
				ac = append(ac, Arg{Key: kv[0], Value: kv[1]})
			} else {
				ac = append(ac, Arg{Key: kv[0], Value: ""})
			}
		} else {
			ac = append(ac, Arg{Key: parts[i], Value: ""})
		}
	}

	return
}

func (ac ArgCommand) String() string {
	var sb strings.Builder
	sb.WriteString("ARG ")
	for _, arg := range ac {
		sb.WriteString(arg.Key)
		if v := arg.Value; v != "" {
			sb.WriteString("=" + v)
		}
		sb.WriteString(" ")
	}
	return sb.String()
}

func (ac ArgCommand) MultiLineString() string {
	var sb strings.Builder
	for _, arg := range ac {
		sb.WriteString("ARG ")
		sb.WriteString(arg.Key)
		if v := arg.Value; v != "" {
			sb.WriteString("=" + v)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (ac ArgCommand) ReplaceArgPrefix(prefix string, val string) string {
	for _, arg := range ac {
		val = strings.ReplaceAll(val, "$"+arg.Key, "$"+prefix+arg.Key)
	}
	return val
}
