/*
This program processes a Dockerfile. When it encounters a FROM command with a relative path,
it pastes the content of the referenced Dockerfile into the current Dockerfile with some modifications:
  - It ensures that all ADD and COPY commands point to the correct context path by adding the relative path
    to the first part of the command (the file or directory being copied).
  - It takes the ARG variables defined before the FROM command and prepends them with the alias of the
    FROM command. It also replaces any occurrences of the ARG variables in the Dockerfile with the new prefixed
    variables. Then it writes them to the beginning of the new Dockerfile.
  - It allows user to specify -client flag to just include already built client directory in the Dockerfile.
    If no client path is specified, it will build the client from the Dockerfile.

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
	clientPath := flag.String("client", "", "Path to the client directory, if not set, the client will be built")
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

	err = ProcessDockerfile(buildcontext, *outputPath, *clientPath)
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

type Dockerfile struct {
	ctx  BuildContext // build context for the current Dockerfile
	args ArgCommand   // global args defined in the Dockerfile

	w *bytes.Buffer
}

// Include reads the requested Dockerfile, modifies it to point to the new context path, and includes it in the
// current Dockerfile. It also replaces the ARG variables with the new prefixed variables.
func (d *Dockerfile) Include(ctx BuildContext, alias string) error {
	// read the Dockerfile
	raw, err := os.ReadFile(ctx.String())
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile: %w", err)
	}

	// count how many FROM lines are in the Dockerfile, we need to know which one is the last one
	// to replace it with our alias
	fromCount := 0
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FROM") {
			fromCount++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read Dockerfile: %w", err)
	}

	// new context path relative to the current context path
	newContextPath, err := filepath.Rel(d.ctx.ContextPath, ctx.ContextPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// use argPrefix to prepend the alias to the ARG variables
	argPrefix := strings.ToUpper(alias) + "_"
	// replace - with _ in the alias
	argPrefix = strings.ReplaceAll(argPrefix, "-", "_")
	// use aliasPrefix to prepend the alias to the ARG variables
	aliasPrefix := alias + "-"

	beforeFrom := true
	globalArgs := ArgCommand{}

	// read the Dockerfile line by line and modify it
	scanner = bufio.NewScanner(bytes.NewReader(raw))
	nthFrom := 0
	for scanner.Scan() {
		line := scanner.Text()

		// handle ARG lines defined before FROM
		if !beforeFrom {
			line = globalArgs.ReplaceArgPrefix(argPrefix, line)
		}

		// we need to move the ARG lines before the FROM line
		if strings.HasPrefix(line, "ARG") {
			args, err := ParseArgCommand(line)
			if err != nil {
				return fmt.Errorf("failed to parse ARG command: %w", err)
			}
			if beforeFrom {
				globalArgs = append(globalArgs, args...)
				log.Printf("[%s] Found global %q before FROM, moving it to the beginning.\n", ctx, args)
			} else {
				// if we are not before FROM and it matches one of the global args, we need to add prefix to it
				// because they may be redefined in the Dockerfile
				argKeys := make(map[string]struct{})
				for _, arg := range globalArgs {
					argKeys[arg.Key] = struct{}{}
				}
				for i := range args {
					if _, ok := argKeys[args[i].Key]; ok {
						log.Printf("[%s] Found global ARG %q after FROM, adding %q prefix.\n", ctx, args[i].Key, argPrefix)
						args[i].Key = argPrefix + args[i].Key
					}
				}
				d.w.WriteString(args.String() + "\n")
			}
			continue
		}

		// modify FROM lines
		if strings.HasPrefix(line, "FROM") {
			nthFrom++

			// parse the FROM command
			cmd, err := ParseFromCommand(line)
			if err != nil {
				return fmt.Errorf("failed to parse FROM command: %w", err)
			}

			// handle the case where ARGs are defined before FROM
			cmd.Image = globalArgs.ReplaceArgPrefix(argPrefix, cmd.Image)

			if nthFrom == fromCount && cmd.Alias != alias {
				log.Printf("[%s] Replacing alias in %q with %q.\n", ctx, cmd, cmd.Alias)
				// if this is the last FROM line, we need to replace with our alias
				cmd.Alias = alias
			}
			if nthFrom != fromCount && alias != "" {
				log.Printf("[%s] Adding alias prefix %q to %q.\n", ctx, aliasPrefix, cmd)
				// this is not the last FROM line, add prefix to the alias
				cmd.Alias = aliasPrefix + cmd.Alias
			}

			beforeFrom = false
			d.w.WriteString(cmd.String() + "\n")
			continue
		}

		// modify COPY and ADD lines
		if strings.HasPrefix(line, "COPY") || strings.HasPrefix(line, "ADD") {
			// parse the COPY/ADD command
			cmd, err := ParseCopyAddCommand(line)
			if err != nil {
				return fmt.Errorf("failed to parse COPY/ADD command: %w", err)
			}

			if _, ok := cmd.Args["from"]; !ok {
				// replace the from part with the new context path
				newFrom := filepath.Join(newContextPath, cmd.From)
				log.Printf("[%s] Path replace: %s -> %s\n", ctx, cmd.From, newFrom)
				cmd.From = newFrom
			} else {
				// add alias prefix to the --from argument
				log.Printf("[%s] Found COPY/ADD with --from=%s, adding %q alias prefix.\n", ctx, cmd.Args["from"], aliasPrefix)
				cmd.Args["from"] = aliasPrefix + cmd.Args["from"]
			}

			d.w.WriteString(cmd.String() + "\n")
			continue
		}

		// write the line as is
		d.w.WriteString(line + "\n")
	}

	// add prefix to global ARGs
	globalArgs.WithPrefix(argPrefix)

	// add the global ARGs to the beginning of the new Dockerfile
	d.args = append(d.args, globalArgs...)

	return scanner.Err()
}

// Process processes the Dockerfile and resolves sub-Dockerfiles in it
func ProcessDockerfile(ctx BuildContext, outputPath, clientPath string) error {
	d := &Dockerfile{
		ctx:  ctx,
		args: make(ArgCommand, 0),
		w:    bytes.NewBuffer(nil),
	}

	// read the Dockerfile
	raw, err := os.ReadFile(ctx.String())
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile: %w", err)
	}

	// read the Dockerfile line by line and modify it
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// modify FROM lines
		if strings.HasPrefix(line, "FROM ./") {
			// parse the FROM command
			cmd, err := ParseFromCommand(line)
			if err != nil {
				return fmt.Errorf("failed to parse FROM command: %w", err)
			}

			// if we are not building the client, skip this line
			if clientPath != "" && cmd.Alias == "client" {
				log.Printf("[%s] Skipping FROM client line.\n", ctx)
				continue
			}

			// resolve environment variables in the image name
			cmd.Image = os.ExpandEnv(cmd.Image)

			// create a new build context
			newBuildcontext, err := ButidContextFromPath(filepath.Join(ctx.ContextPath, cmd.Image))
			if err != nil {
				return fmt.Errorf("failed to get build context: %w", err)
			}

			// resolve the dockerfile content
			err = d.Include(newBuildcontext, cmd.Alias)
			if err != nil {
				return fmt.Errorf("failed to get relative Dockerfile: %w", err)
			}

			continue
		}

		// modify COPY and ADD lines
		if strings.HasPrefix(line, "COPY") || strings.HasPrefix(line, "ADD") {
			// parse the COPY/ADD command
			cmd, err := ParseCopyAddCommand(line)
			if err != nil {
				return fmt.Errorf("failed to parse COPY/ADD command: %w", err)
			}

			// if we are not building the client, take if from the client path
			if clientPath != "" && cmd.Args["from"] == "client" {
				log.Printf("[%s] Replacing COPY/ADD --from=client with %q.\n", ctx, clientPath)
				delete(cmd.Args, "from")
				cmd.From = clientPath
				d.w.WriteString(cmd.String() + "\n")
				continue
			}
		}

		// copy all other lines as is
		d.w.WriteString(line + "\n")
	}

	// check for errors while reading the Dockerfile
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input Dockerfile: %w", err)
	}

	// add the global ARGs to the beginning of the new Dockerfile
	prefix := "# THIS FILE IS GENERATED, DO NOT EDIT\n"
	outBytes := append([]byte(prefix+d.args.MultiLineString()), d.w.Bytes()...)

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

func (bc BuildContext) String() string {
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

func (fc FromCommand) String() string {
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

func (ac ArgCommand) WithPrefix(prefix string) {
	for i := range ac {
		if ac[i].Key != "" {
			ac[i].Key = prefix + ac[i].Key
		}
	}
}

func (ac ArgCommand) ReplaceArgPrefix(prefix string, val string) string {
	for _, arg := range ac {
		val = strings.ReplaceAll(val, "$"+arg.Key, "$"+prefix+arg.Key)
		val = strings.ReplaceAll(val, "${"+arg.Key+"}", "${"+prefix+arg.Key+"}")
	}
	return val
}

// CopyAddCommand represents the COPY and ADD commands in a Dockerfile
type CopyAddCommand struct {
	Command string
	Args    map[string]string
	From    string
	To      string
}

func ParseCopyAddCommand(line string) (ca CopyAddCommand, err error) {
	parts := strings.Fields(line)
	if len(parts) < 2 || (strings.ToLower(parts[0]) != "copy" && strings.ToLower(parts[0]) != "add") {
		err = fmt.Errorf("invalid COPY/ADD line: %s", line)
		return
	}

	ca.Command = parts[0]

	ca.Args = make(map[string]string)
	for i := 1; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "--") {
			kv := strings.SplitN(parts[i][2:], "=", 2)
			if len(kv) == 2 {
				ca.Args[kv[0]] = kv[1]
			} else {
				ca.Args[kv[0]] = ""
			}
			continue
		}
		if ca.From == "" {
			ca.From = parts[i]
			continue
		}
		if ca.To == "" {
			ca.To = parts[i]
			continue
		}
	}

	return
}

func (ca CopyAddCommand) String() string {
	var sb strings.Builder
	sb.WriteString(ca.Command + " ")
	for k, v := range ca.Args {
		sb.WriteString("--" + k)
		if v != "" {
			sb.WriteString("=" + v)
		}
		sb.WriteString(" ")
	}
	if ca.From != "" {
		sb.WriteString(ca.From + " ")
	}
	if ca.To != "" {
		sb.WriteString(ca.To)
	}
	return sb.String()
}
