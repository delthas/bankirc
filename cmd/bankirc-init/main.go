package main

import (
	"flag"
	"fmt"
	"github.com/delthas/bankirc"
	"github.com/frieser/nordigen-go-lib/v2"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var cfg *bankirc.Config

func main() {
	config := flag.String("config", "bankirc.yaml", "bankirc configuration file path")
	institution := flag.String("bank", "", "bank ID")
	name := flag.String("name", "", "bank name")
	flag.Parse()
	if *institution == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *name == "" {
		i := strings.LastIndexByte(*institution, '_')
		if i == -1 {
			i = len(*institution)
		}
		*name = (*institution)[:i]
	}

	var err error
	cfg, err = bankirc.ReadConfig(*config)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Fatal("client ID and client secret are required")
	}

	c, err := nordigen.NewClient(cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan struct{}, 1)
	ln, err := net.Listen("tcp4", ":")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		select {
		case ch <- struct{}{}:
		default:
		}
		w.Write([]byte("This window can now be closed."))
	})
	go func() {
		var srv http.Server
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	req := nordigen.Requisition{
		Redirect:      fmt.Sprintf("http://%v/redirect", ln.Addr()),
		InstitutionId: *institution,
	}
	r, err := c.CreateRequisition(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Open the URL below in your browser:")
	fmt.Println(r.Link)
	<-ch
	for r.Status == "CR" {
		r, err = c.GetRequisition(r.Id)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)
	}
	if len(r.Accounts) == 0 {
		log.Fatal("no accounts found")
	}

	for i, account := range cfg.Accounts {
		if account.Name == *name {
			cfg.Accounts = append(cfg.Accounts[:i], cfg.Accounts[i+1:]...)
			break
		}
	}
	cfg.Accounts = append(cfg.Accounts, bankirc.Account{
		Bank: *institution,
		Name: *name,
		ID:   r.Accounts[0],
	})

	if err := bankirc.WriteConfig(*config, cfg); err != nil {
		log.Fatal(err)
	}
}
