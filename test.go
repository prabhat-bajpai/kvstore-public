package main

import (
		"net"
		"os"
		"fmt"
		"time"
		"math/rand"
		)

func main() {

			service := "localhost:1200"
			rand.Seed(time.Now().UTC().UnixNano())							
			tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
			checkError(err)
			
			for j := 1; j <= 30; j++ {
									    go func(j string) {
											// we choose randomly what operation we do like set get delete 
											v := rand.Intn(3)
											
											// randomly set get or delete a value in in range 0 to 9
											v1 := string(rand.Intn(10)+48)
											mapping :=[3]string{"set","get","delete"}
											
											conn, err := net.DialTCP("tcp", nil, tcpAddr)
											checkError(err)
											
											if v == 0{	
											_, err = conn.Write([]byte("set"+" "+ v1+" "+ v1))
											} else {
											_, err = conn.Write([]byte(mapping[v]+" "+ v1))											
											}
											
											checkError(err)
												
											var buf [512]byte
											operation := ""
											n, err := conn.Read(buf[0:])
											operation += string(buf[0:n])											
											fmt.Println(operation)
									   
									   }(string(j+47))
									}
										time.Sleep(time.Second*10)
			}

func checkError(err error) {
							if err != nil {
											fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
											os.Exit(1)
										}
							}
							
							
							
							
