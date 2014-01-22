
package main

import (
		"fmt"
		"net"
		"os"
		s "strings"
		"sync"
		)
func main() {

			data := make(map[string]string)
			var mutex = &sync.Mutex{}
			
			service := "localhost:1200"

			tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
			checkError(err)

			listener, err := net.ListenTCP("tcp", tcpAddr)
			checkError(err)
			
			for {
				conn, err := listener.Accept()
				if err != nil {
								continue
							  }
				go func() { 		
							var buf [512]byte
							operation := ""
							
								n, err := conn.Read(buf[0:])
								operation += string(buf[0:n])							
							
							
							a := s.Split(operation, " ")							
							if a[0] == "set" {
											mutex.Lock()
                							data[a[1]] = a[2]
											mutex.Unlock()
											_, err = conn.Write([]byte("Value set successfully"))
											checkError(err)
											}
							if a[0] == "get" {
											    val, ok := data[a[1]] 
 												if ok {
												_, err = conn.Write([]byte(data[a[1]]+ " : "+ val))
												}else {
												_, err = conn.Write([]byte("Key not exist"))
											}}
							if a[0] == "delete" {
											mutex.Lock()
											delete(data,a[1])
											mutex.Unlock()
											_, err = conn.Write([]byte("Value delete Successfully"))
											}				
							fmt.Println(data)
						}()
				}
			}
						
func checkError(err error) {
							if err != nil {
											fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
											os.Exit(1)
										}
							}
