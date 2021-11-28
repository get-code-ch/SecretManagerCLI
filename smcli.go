package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/get-code-ch/SecretManager"
	"log"
	"os"
	"regexp"
	"strings"
)

type ReadCommand struct {
	fs          *flag.FlagSet
	application string
}

type DeleteCommand struct {
	fs          *flag.FlagSet
	application string
}

type UpsertCommand struct {
	fs          *flag.FlagSet
	application string
	username    string
	password    string
	parameters  Params
}

type Param struct {
	key   string
	value string
}

type Params []Param

type Runner interface {
	Init([]string) error
	Name() string
	Run(*SecretManager.Vault) error
}

// Runner interface implementation for Read sub-command
func (r *ReadCommand) Init(args []string) error {
	return r.fs.Parse(args)
}

func (r *ReadCommand) Run(vault *SecretManager.Vault) error {
	if sRead, err := vault.Read(r.application); err == nil {
		fmt.Printf("Username -> %s\n", sRead.Username)
		fmt.Printf("Password -> %s\n", sRead.Password)
		fmt.Printf("Params -> %v\n", sRead.Parameters)
	} else {
		return err
	}
	return nil
}

func (r *ReadCommand) Name() string {
	return r.fs.Name()
}

// Runner interface implementation for Upsert sub-command
func (u *UpsertCommand) Init(args []string) error {
	return u.fs.Parse(args)
}

func (u *UpsertCommand) Run(vault *SecretManager.Vault) error {
	var secret SecretManager.Secret
	var err error

	if secret, err = vault.Read(u.application); err == nil {
		if u.username != "" {
			secret.Username = u.username
		}
		if u.password != "" {
			secret.Password = u.password
		}
	} else {
		secret = SecretManager.Secret{Application: u.application, Username: u.username, Password: u.password}
	}
	if secret.Parameters == nil && len(u.parameters) > 0 {
		secret.Parameters = make(map[string]string)
	}
	for _, item := range u.parameters {
		secret.Parameters[item.key] = item.value
	}
	vault.Upsert(secret)
	return nil
}

func (u *UpsertCommand) Name() string {
	return u.fs.Name()
}

// Runner interface implementation for Delete sub-command
func (d *DeleteCommand) Init(args []string) error {
	return d.fs.Parse(args)
}

func (d *DeleteCommand) Run(vault *SecretManager.Vault) error {
	return vault.Delete(d.application)
}

func (d *DeleteCommand) Name() string {
	return d.fs.Name()
}


func (p *Params) String() string {
	var params string
	if len(*p) > 0 {
		for _, value := range *p {
			params += "/" + value.key
		}
	}
	return ""
}

func (p *Params) Set(s string) error {
	sep := regexp.MustCompile(`:|,|;|,|=`)
	if split := sep.Split(s, -1); len(split) != 2 {
		return errors.New("invalid key:value argument for parameter")
	} else {
		*p = append(*p, Param{key: split[0], value: split[1]})
	}
	return nil
}

func main() {
	readCmd := &ReadCommand{
		fs: flag.NewFlagSet("read", flag.ContinueOnError),
	}
	readCmd.fs.StringVar(&readCmd.application, "application", "", "Application name")

	deleteCmd := &DeleteCommand{
		fs: flag.NewFlagSet("delete", flag.ContinueOnError),
	}
	deleteCmd.fs.StringVar(&deleteCmd.application, "application", "", "Application name")

	upsertCmd := &UpsertCommand{
		fs: flag.NewFlagSet("upsert", flag.ContinueOnError),
	}
	upsertCmd.fs.StringVar(&upsertCmd.application, "application", "", "Application name")
	upsertCmd.fs.StringVar(&upsertCmd.username, "username", "", "Application username")
	upsertCmd.fs.StringVar(&upsertCmd.password, "password", "", "Application password")
	upsertCmd.fs.Var(&upsertCmd.parameters, "parameter", "Application parameters format key:value or key=value")

	cmds := []Runner{
		readCmd,
		upsertCmd,
		deleteCmd,
	}

	subcommand := os.Args[1]

	vault := new(SecretManager.Vault)
	if err := vault.Open(); err != nil {
		log.Fatalf("Error opening vault --> %v", err)
	}
	defer vault.Close()

	for _, cmd := range cmds {
		if strings.ToUpper(cmd.Name()) == strings.ToUpper(subcommand) {
			cmd.Init(os.Args[2:])
			if err := cmd.Run(vault); err != nil {
				log.Fatalf("Error executing command %s --> %v", cmd.Name(), err)
			}
			break
		}
	}
}
