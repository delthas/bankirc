package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/delthas/bankirc"
	"github.com/frieser/nordigen-go-lib/v2"
	"gopkg.in/irc.v3"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

type transactionKey struct {
	Account string
	Pending bool
}

var cfg *bankirc.Config

var lines = make(chan string, 1024)

func runBank() error {
	cache := make(map[transactionKey]string)
	c, err := nordigen.NewClient(cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		return fmt.Errorf("opening nordigen: %v", err)
	}

	go func() {
		for {
			for _, account := range cfg.Accounts {
				fmt.Println(account.Name, account.Bank)
				res, err := c.GetAccountTransactions(account.ID)
				if err != nil {
					log.Printf("getting transactions for %v: %v", account.Name, err)
					continue
				}
				for _, txs := range []*[]nordigen.Transaction{&res.Transactions.Booked, &res.Transactions.Pending} {
					pending := txs == &res.Transactions.Pending
					key := transactionKey{
						Account: account.Name,
						Pending: pending,
					}
					last, ok := cache[key]
					if ok && last != "" {
						for i, tx := range *txs {
							if tx.InternalTransactionId == last {
								*txs = (*txs)[:i]
								break
							}
						}
					}
					if len(*txs) > 0 {
						cache[key] = (*txs)[0].InternalTransactionId
					} else {
						cache[key] = ""
					}
					if !ok {
						continue
					}
					for i := len(*txs) - 1; i >= 0; i-- {
						tx := (*txs)[i]
						var desc string
						if len(tx.RemittanceInformationUnstructuredArray) > 0 {
							desc = tx.RemittanceInformationUnstructuredArray[0]
						}
						if pending {
							desc += " (pending)"
						}
						text := fmt.Sprintf("%v: %v: %v %v: %v", account.Name, tx.BookingDate, tx.TransactionAmount.Amount, tx.TransactionAmount.Currency, desc)
						lines <- text
					}
				}
			}
			time.Sleep(1 * time.Hour)
		}
	}()
	return nil
}

func runIRC() error {
	var doTLS bool
	var host string
	if u, err := url.Parse(cfg.IRCServer); err == nil && u.Scheme != "" && u.Host != "" {
		switch u.Scheme {
		case "ircs":
			doTLS = true
		case "irc+insecure", "irc":
		default:
			return fmt.Errorf("invalid IRC addr scheme: %v", cfg.IRCServer)
		}
		host = u.Host
	} else if strings.Contains(cfg.IRCServer, ":+") {
		doTLS = true
		host = strings.ReplaceAll(cfg.IRCServer, ":+", ":")
	} else {
		host = cfg.IRCServer
	}
	go func() {
		var closeCh chan struct{}
		first := true
		for {
			if closeCh != nil {
				close(closeCh)
				closeCh = nil
			}
			if first {
				first = false
			} else {
				time.Sleep(10 * time.Second)
			}
			var nc net.Conn
			var err error
			if doTLS {
				nc, err = tls.Dial("tcp", host, nil)
			} else {
				nc, err = net.Dial("tcp", host)
			}
			if err != nil {
				log.Printf("connecting to irc: %v", err)
				continue
			}
			c := irc.NewClient(nc, irc.ClientConfig{
				Nick:      cfg.Nick,
				User:      cfg.Nick,
				Name:      cfg.Nick,
				SendLimit: 500 * time.Millisecond,
				SendBurst: 4,
				Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
					switch m.Command {
					case "001":
						c.WriteMessage(&irc.Message{
							Command: "JOIN",
							Params:  []string{cfg.Channel},
						})
					case "JOIN":
						if m.Name == c.CurrentNick() {
							closeCh = make(chan struct{}, 1)
							go func() {
								for {
									select {
									case line := <-lines:
										c.WriteMessage(&irc.Message{
											Command: "PRIVMSG",
											Params:  []string{cfg.Channel, line},
										})
									case <-closeCh:
										return
									}
								}
							}()
						}
					}
				}),
			})
			if err := c.Run(); err != nil {
				log.Printf("running irc: %v", err)
			}
		}
	}()
	return nil
}

func main() {
	path := flag.String("config", "bankirc.yaml", "bankirc configuration file path")
	flag.Parse()

	var err error
	cfg, err = bankirc.ReadConfig(*path)
	if err != nil {
		log.Fatal(err)
	}

	if err := runBank(); err != nil {
		log.Fatal(err)
	}
	if err := runIRC(); err != nil {
		log.Fatal(err)
	}

	select {}
}
