package httpclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"goldclient/models"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

type CustomClient struct {
	conn    *net.TCPConn
	host    string
	d       net.Dialer
	tcpAddr *net.TCPAddr
}

func NewCustomClient(hostname string) CustomClient {
	result, err := net.LookupHost(hostname)
	if err != nil {
		panic(err)
	}
	addr, err := net.ResolveTCPAddr("tcp4", hostname+":8000")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "lookup (len, first): %d %s\n", len(result), result[0])
	return CustomClient{
		conn: nil,
		host: result[0],
		d: net.Dialer{
			KeepAlive: -1,
		},
		tcpAddr: addr,
	}
}

func (c *CustomClient) RawExplore(hostname string, url string, content string) models.AreaResponse {
	conn, err := net.DialTCP("tcp", nil, c.tcpAddr)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("POST " + url + " HTTP/1.0\nHost: " + hostname + "\nContent-Type: application/json\nContent-Length: " + strconv.Itoa(len(content)) + "\n\n" + content))

	response, err := ioutil.ReadAll(conn)
	if err != nil {
		panic(err)
	}

	start := 0
	t := false
	for i := len(response) - 1; i >= 0; i-- {
		if response[i] == '{' {
			if t == true {
				start = i
				break
			}
			t = true
		}
	}
	conn.Close()

	var result models.AreaResponse
	json.Unmarshal(response[start:], &result)

	return result
}

func (c *CustomClient) RawCash(hostname, url, content string) []int {
	conn, err := net.DialTCP("tcp", nil, c.tcpAddr)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("POST " + url + " HTTP/1.0\nHost: " + hostname + "\nContent-Type: application/json\nContent-Length: " + strconv.Itoa(len(content)) + "\n\n" + content))

	reader := bufio.NewReader(conn)
	firstLine, _, _ := reader.ReadLine()
	statusCode := getResponseCode(string(firstLine))

	if statusCode != 200 {
		return nil
	}

	response, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	start := 0
	for i := len(response) - 1; i >= 0; i-- {
		if response[i] == '[' {
			start = i
			break
		}
	}

	var result []int
	json.Unmarshal(response[start:], &result)
	return result
}

func (c *CustomClient) RawLicense(hostname, url, content string) models.Licence {
	conn, err := net.DialTCP("tcp", nil, c.tcpAddr)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("POST " + url + " HTTP/1.0\nHost: " + hostname + "\nContent-Type: application/json\nContent-Length: " + strconv.Itoa(len(content)) + "\n\n" + content))

	response, err := ioutil.ReadAll(conn)
	if err != nil {
		panic(err)
	}
	start := 0
	for i := len(response) - 1; i >= 0; i-- {
		if response[i] == '{' {
			start = i
			break
		}
	}

	conn.Close()

	var result models.Licence
	json.Unmarshal(response[start:], &result)
	return result
}

func (c *CustomClient) RawDig(hostname string, url string, content string, depth int) []string {
	conn, err := net.DialTCP("tcp", nil, c.tcpAddr)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("POST " + url + " HTTP/1.0\nHost: " + hostname + "\nContent-Type: application/json\nContent-Length: " + strconv.Itoa(len(content)) + "\n\n" + content))

	reader := bufio.NewReader(conn)
	firstLine, _, _ := reader.ReadLine()
	statusCode := getResponseCode(string(firstLine))

	if statusCode == 429 {
		fmt.Fprintf(os.Stderr, "ERROR 429 in dig \n")
	}
	if statusCode == 404 {
		conn.Close()
		return nil
	}

	response, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	start := 0
	for i := len(response) - 1; i >= 0; i-- {
		if response[i] == '[' {
			start = i
			break
		}
	}

	conn.Close()
	var result []string
	json.Unmarshal(response[start:], &result)
	return result
}

func getResponseCode(line string) int {
	i := strings.IndexByte(line, ' ')
	status := strings.TrimLeft(line[i+1:], " ")
	i = strings.IndexByte(status, ' ')
	statusCode := status[:i]
	statusCodeInt, _ := strconv.Atoi(statusCode)
	return statusCodeInt
}
