package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	logger "gx/ipfs/QmcaSwFc5RBg8yCq54QURwEU4nwjfCpjbpmaAm4VbdGLKv/go-logging"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/go-homedir"
	"github.com/textileio/textile-go/cmd"
	"github.com/textileio/textile-go/core"
	"github.com/textileio/textile-go/gateway"
	"github.com/textileio/textile-go/keypair"
	"github.com/textileio/textile-go/wallet"
	"gopkg.in/abiosoft/ishell.v2"
)

type ipfsOptions struct {
	ServerMode bool   `long:"server" description:"Apply IPFS server profile."`
	SwarmPorts string `long:"swarm-ports" description:"Set the swarm ports (TCP,WS). Random ports are chosen by default."`
}

type logOptions struct {
	Level   string `short:"l" long:"log-level" description:"Set the logging level [debug, info, notice, warning, error, critical]." default:"error"`
	NoFiles bool   `short:"n" long:"no-log-files" description:"Write logs to stdout instead of rolling files."`
}

type addressOptions struct {
	ApiBindAddr     string `short:"a" long:"api-bind-addr" description:"Set the local API address." default:"127.0.0.1:40600"`
	CafeApiBindAddr string `short:"c" long:"cafe-bind-addr" description:"Set the cafe REST API address." default:"127.0.0.1:40601"`
	GatewayBindAddr string `short:"g" long:"gateway-bind-addr" description:"Set the IPFS gateway address." default:"127.0.0.1:5050"`
}

type cafeOptions struct {
	PublicIP string `long:"public-ip" description:"Required with --cafe-open on an image / server with a public IP address."`
	Open     bool   `long:"cafe-open" description:"Opens the p2p Cafe Service for other peers."`
}

type options struct{}

type walletCmd struct {
	Init     walletInitCmd     `command:"init"`
	Accounts walletAccountsCmd `command:"accounts"`
}

type walletInitCmd struct {
	WordCount int    `short:"w" long:"word-count" description:"Number of mnemonic recovery phrase words: 12,15,18,21,24." default:"12"`
	Password  string `short:"p" long:"password" description:"Mnemonic recovery phrase password (omit if none)."`
}

type walletAccountsCmd struct {
	Password string `short:"p" long:"password" description:"Mnemonic recovery phrase password (omit if none)."`
	Depth    int    `short:"d" long:"depth" description:"Number of accounts to show." default:"1"`
	Offset   int    `short:"o" long:"offset" description:"Account depth to start from." default:"0"`
}

type versionCmd struct{}

type initCmd struct {
	AccountSeed string         `required:"true" short:"s" long:"seed" description:"Account seed (run 'wallet' command to generate new seeds)."`
	PinCode     string         `short:"p" long:"pin-code" description:"Specify a pin code for datastore encryption."`
	RepoPath    string         `short:"r" long:"repo-dir" description:"Specify a custom repository path."`
	Addresses   addressOptions `group:"Address Options"`
	CafeOptions cafeOptions    `group:"Cafe Options"`
	IPFS        ipfsOptions    `group:"IPFS Options"`
	Logs        logOptions     `group:"Log Options"`
}

type migrateCmd struct {
	RepoPath string `short:"r" long:"repo-dir" description:"Specify a custom repository path."`
	PinCode  string `short:"p" long:"pin-code" description:"Specify the pin code for datastore encryption (omit of none was used during init)."`
}

type daemonCmd struct {
	PinCode  string `short:"p" long:"pin-code" description:"Specify the pin code for datastore encryption (omit of none was used during init)."`
	RepoPath string `short:"r" long:"repo-dir" description:"Specify a custom repository path."`
}

var node *core.Textile

var parser = flags.NewParser(&options{}, flags.Default)

func init() {
	// add main commands
	parser.AddCommand("version",
		"Print version and exit",
		"Print the current version and exit.",
		&versionCmd{})
	parser.AddCommand("wallet",
		"Manage or create an account wallet",
		"Initialize a new wallet, or view accounts from an existing wallet.",
		&walletCmd{})
	parser.AddCommand("init",
		"Init the node repo and exit",
		"Initialize the node repository and exit.",
		&initCmd{})
	parser.AddCommand("migrate",
		"Migrate the node repo and exit",
		"Migrate the node repository and exit.",
		&migrateCmd{})
	parser.AddCommand("daemon",
		"Start the daemon",
		"Start a node daemon session.",
		&daemonCmd{})

	// add cmd commands
	for _, c := range cmd.Cmds() {
		parser.AddCommand(c.Name(), c.Short(), c.Long(), c)
	}
}

func main() {
	parser.Parse()
}

func (x *walletInitCmd) Execute(args []string) error {
	wcount, err := wallet.NewWordCount(x.WordCount)
	if err != nil {
		return err
	}

	w, err := wallet.NewWallet(wcount.EntropySize())
	if err != nil {
		return err
	}

	fmt.Println(strings.Repeat("-", len(w.RecoveryPhrase)+4))
	fmt.Println("| " + w.RecoveryPhrase + " |")
	fmt.Println(strings.Repeat("-", len(w.RecoveryPhrase)+4))
	fmt.Println("WARNING! Store these words above in a safe place!")
	fmt.Println("WARNING! If you lose your words, you will lose access to data in all derived accounts!")
	fmt.Println("WARNING! Anyone who has access to these words can access your wallet accounts!")
	fmt.Println("")
	fmt.Println("Use: `wallet accounts` command to inspect more accounts.")
	fmt.Println("")

	// show first account
	kp, err := w.AccountAt(0, x.Password)
	if err != nil {
		return err
	}
	fmt.Println("--- ACCOUNT 0 ---")
	fmt.Println(kp.Address())
	fmt.Println(kp.Seed())

	return nil
}

var wordsRegexp = regexp.MustCompile(`^[a-z]+$`)

func (x *walletAccountsCmd) Execute(args []string) error {
	if x.Depth < 1 || x.Depth > 100 {
		return errors.New("depth must be greater than 0 and less than 100")
	}
	if x.Offset < 0 || x.Offset > x.Depth {
		return errors.New("offset must be greater than 0 and less than depth")
	}

	// create a shell for reading input
	shell := ishell.New()

	// determine word count
	count := shell.MultiChoice([]string{"12", "15", "18", "21", "24"}, "How many words are in your mnemonic recovery phrase?")
	var wcount wallet.WordCount
	switch count {
	case 0:
		wcount = wallet.TwelveWords
	case 1:
		wcount = wallet.FifteenWords
	case 2:
		wcount = wallet.EighteenWords
	case 3:
		wcount = wallet.TwentyOneWords
	case 4:
		wcount = wallet.TwentyFourWords
	default:
		return wallet.ErrInvalidWordCount
	}

	words := make([]string, int(wcount))
	for i := 0; i < int(wcount); i++ {
		shell.Print(fmt.Sprintf("Enter word #%d: ", i+1))
		words[i] = shell.ReadLine()
		if !wordsRegexp.MatchString(words[i]) {
			shell.Println("Invalid word, try again.")
			i--
		}
	}
	wall := wallet.NewWalletFromRecoveryPhrase(strings.Join(words, " "))

	for i := x.Offset; i < x.Offset+x.Depth; i++ {
		kp, err := wall.AccountAt(i, x.Password)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("--- ACCOUNT %d ---", i))
		fmt.Println(fmt.Sprintf("PUBLIC ADDR: %s", kp.Address()))
		fmt.Println(fmt.Sprintf("SECRET SEED: %s", kp.Seed()))
	}
	return nil
}

func (x *versionCmd) Execute(args []string) error {
	fmt.Println(core.Version)
	return nil
}

func (x *initCmd) Execute(args []string) error {
	// build keypair from provided seed
	kp, err := keypair.Parse(x.AccountSeed)
	if err != nil {
		return errors.New(fmt.Sprintf("parse account seed failed: %s", err))
	}
	accnt, ok := kp.(*keypair.Full)
	if !ok {
		return keypair.ErrInvalidKey
	}

	repoPath, err := getRepoPath(x.RepoPath)
	if err != nil {
		return err
	}

	level, err := logger.LogLevel(strings.ToUpper(x.Logs.Level))
	if err != nil {
		return errors.New(fmt.Sprintf("failed to determine log level: %s", err))
	}

	config := core.InitConfig{
		Account:      accnt,
		PinCode:      x.PinCode,
		RepoPath:     repoPath,
		SwarmPorts:   x.IPFS.SwarmPorts,
		ApiAddr:      x.Addresses.ApiBindAddr,
		CafeApiAddr:  x.Addresses.CafeApiBindAddr,
		GatewayAddr:  x.Addresses.GatewayBindAddr,
		IsMobile:     false,
		IsServer:     x.IPFS.ServerMode,
		LogLevel:     level,
		LogToDisk:    !x.Logs.NoFiles,
		CafeOpen:     x.CafeOptions.Open,
		CafePublicIP: x.CafeOptions.PublicIP,
	}

	if err := core.InitRepo(config); err != nil {
		return errors.New(fmt.Sprintf("initialize failed: %s", err))
	}
	fmt.Printf("Initialized account with address %s\n", accnt.Address())
	return nil
}

func (x *migrateCmd) Execute(args []string) error {
	repoPath, err := getRepoPath(x.RepoPath)
	if err != nil {
		return err
	}

	if err := core.MigrateRepo(core.MigrateConfig{
		PinCode:  x.PinCode,
		RepoPath: repoPath,
	}); err != nil {
		return errors.New(fmt.Sprintf("migrate repo: %s", err))
	}
	fmt.Println("Repo was successfully migrated")
	return nil
}

func (x *daemonCmd) Execute(args []string) error {
	if err := buildNode(x.PinCode, x.RepoPath); err != nil {
		return err
	}
	printSplash()

	// handle interrupt
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Interrupted")
	fmt.Printf("Shutting down...")
	if err := stopNode(); err != nil && err != core.ErrStopped {
		fmt.Println(err.Error())
	} else {
		fmt.Print("done\n")
	}
	os.Exit(1)
	return nil
}

func getRepoPath(repoPath string) (string, error) {
	if len(repoPath) == 0 {
		// get homedir
		home, err := homedir.Dir()
		if err != nil {
			return "", errors.New(fmt.Sprintf("get homedir failed: %s", err))
		}

		// ensure app folder is created
		appDir := filepath.Join(home, ".textile")
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return "", errors.New(fmt.Sprintf("create repo directory failed: %s", err))
		}
		repoPath = filepath.Join(appDir, "repo")
	}
	return repoPath, nil
}

func buildNode(pinCode string, repoPath string) error {
	repoPathf, err := getRepoPath(repoPath)
	if err != nil {
		return err
	}

	node, err = core.NewTextile(core.RunConfig{
		PinCode:  pinCode,
		RepoPath: repoPathf,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("create node failed: %s", err))
	}

	gateway.Host = &gateway.Gateway{
		Node: node,
	}

	if err := startNode(); err != nil {
		return errors.New(fmt.Sprintf("start node failed: %s", err))
	}
	return nil
}

func startNode() error {
	listener := node.GetThreadUpdateListener()

	if err := node.Start(); err != nil {
		return err
	}

	// subscribe to wallet updates
	go func() {
		for {
			select {
			case update, ok := <-node.UpdateCh():
				if !ok {
					return
				}
				switch update.Type {
				case core.ThreadAdded:
					break
				case core.ThreadRemoved:
					break
				case core.AccountPeerAdded:
					break
				case core.AccountPeerRemoved:
					break
				}
			}
		}
	}()

	// subscribe to thread updates
	go func() {
		for {
			select {
			case value, ok := <-listener.Ch:
				if !ok {
					return
				}
				if update, ok := value.(core.ThreadUpdate); ok {
					date := update.Block.Date.Format(time.RFC822)
					desc := update.Block.Type
					thrd := update.ThreadId[len(update.ThreadId)-8:]

					if update.Block.Username != "" {
						update.Block.Username += " "
					}

					msg := cmd.Grey(date+"  "+update.Block.Username+"added ") +
						cmd.Green(desc) + cmd.Grey(" update to thread "+thrd)
					fmt.Println(msg)
				}
			}
		}
	}()

	// subscribe to notifications
	go func() {
		for {
			select {
			case note, ok := <-node.NotificationCh():
				if !ok {
					return
				}

				date := note.Date.Format(time.RFC822)
				var subject string
				if len(note.SubjectId) >= 7 {
					subject = note.SubjectId[len(note.SubjectId)-7:]
				}

				msg := cmd.Grey(date+"  "+note.Username+" ") + cmd.Cyan(note.Body) +
					cmd.Grey(" "+subject)
				fmt.Println(msg)
			}
		}
	}()

	// start apis
	node.StartApi(node.Config().Addresses.API)
	gateway.Host.Start(node.Config().Addresses.Gateway)

	<-node.OnlineCh()

	return nil
}

func stopNode() error {
	if err := node.StopApi(); err != nil {
		return err
	}
	if err := gateway.Host.Stop(); err != nil {
		return err
	}
	if err := node.Stop(); err != nil {
		return err
	}

	node.CloseChns()
	return nil
}

func printSplash() {
	pid, err := node.PeerId()
	if err != nil {
		log.Fatalf("get peer id failed: %s", err)
	}
	fmt.Println(cmd.Grey("Textile daemon version v" + core.Version))
	fmt.Println(cmd.Grey("Repo:    ") + cmd.Grey(node.RepoPath()))
	fmt.Println(cmd.Grey("API:     ") + cmd.Grey(node.ApiAddr()))
	fmt.Println(cmd.Grey("Gateway: ") + cmd.Grey(gateway.Host.Addr()))
	if node.CafeApiAddr() != "" {
		fmt.Println(cmd.Grey("Cafe:    ") + cmd.Grey(node.CafeApiAddr()))
	}
	fmt.Println(cmd.Grey("PeerID:  ") + cmd.Green(pid.Pretty()))
	fmt.Println(cmd.Grey("Account: ") + cmd.Cyan(node.Account().Address()))
}
