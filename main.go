package main

import (
        "bufio"
        "crypto/tls"
        "flag"
        "fmt"
        "log"
        "net"
        "os"
        "sync"
        "time"
)

func processDomain(domainChan <-chan string, wg *sync.WaitGroup) {
        defer wg.Done()

        for domainName := range domainChan {
                conn, err := tls.DialWithDialer(&net.Dialer{
                        Timeout: 2 * time.Second,
                }, "tcp", domainName, &tls.Config{
                        InsecureSkipVerify: true,
                })
                if err != nil {
                        continue
                }
                defer conn.Close()

                state := conn.ConnectionState()
                cert := state.PeerCertificates[0]

                fmt.Println(domainName + " " + cert.Subject.CommonName + " " + cert.Issuer.CommonName)
        }
}

func main() {
        concurrency := flag.Int("p", 5, "number of concurrent connections")
        flag.Parse()

        domainChan := make(chan string, *concurrency)
        wg := &sync.WaitGroup{}

        scanner := bufio.NewScanner(os.Stdin)
        for i := 0; i < *concurrency; i++ {
                wg.Add(1)
                go processDomain(domainChan, wg)
        }

        for scanner.Scan() {
                domainName := scanner.Text()
                domainChan <- domainName
        }

        close(domainChan)
        wg.Wait()

        if err := scanner.Err(); err != nil {
                log.Fatal(err)
        }
}