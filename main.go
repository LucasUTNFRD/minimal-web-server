package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const CLRF = "\r\n"

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	log.Println("Starting server on :4221")
	if err != nil {
		log.Fatalf("Failed to bind to port 4221: %v", err)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn) // Handle each connection in a separate goroutine
	}
}

type HTTPRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    string
}

type Request struct {
	HttpRequest
	RequestHeaders
	RequestBody string
	Accept      string
}

type HttpRequest struct {
	Method  string
	Target  string
	Version string
}

type RequestHeaders struct {
	Host      string
	UserAgent string
}

type ResponseHeaders struct {
	ContentType   string
	ContentLength int
}

type ResponseBody struct {
	Message string
}

type Response struct {
	Version       string
	StatusCode    int
	StatusMessage string
	ResponseHeaders
	ResponseBody
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("Handling new connection")

	request := NewRequest(conn)
	log.Printf("Parsed request: %+v\n", request)

	response := request.Parse()
	log.Printf("Generated response: %+v\n", response)

	resp := request.WriteResponse(response)
	log.Printf("Response to write: %s\n", string(resp))

	_, err := conn.Write(resp)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func NewRequest(conn net.Conn) Request {
	var httpReq HttpRequest
	var reqHeaders RequestHeaders
	var requestBody string
	l := bufio.NewReader(conn)
	for {
		line, err := l.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			break
		}
		log.Printf("Read line: %s", line)

		if line == CLRF {
			break
		}
		if strings.HasPrefix(line, "User-Agent:") {
			s := strings.Split(line, " ")[1]
			reqHeaders.UserAgent = s
		}
		if strings.HasPrefix(line, "Host:") {
			s := strings.Split(line, " ")[1]
			reqHeaders.Host = s
		}
		if strings.Contains(line, "HTTP/1.1") {
			s1 := strings.Split(line, " ")[0]
			s2 := strings.Split(line, " ")[1]
			s3 := strings.Split(line, " ")[2]
			httpReq.Version = strings.TrimSuffix(s3, CLRF)
			httpReq.Method = s1
			httpReq.Target = s2
		}
	}

	return Request{HttpRequest: httpReq, RequestHeaders: reqHeaders, RequestBody: requestBody, Accept: "*/*"}
}

func (r *Request) Parse() Response {
	switch {
	case r.Target == "/":
		return Response{
			Version:       r.Version,
			StatusCode:    200,
			StatusMessage: "OK",
			ResponseHeaders: ResponseHeaders{
				ContentType:   "text/plain",
				ContentLength: 0,
			},
			ResponseBody: ResponseBody{
				Message: "",
			},
		}
	case strings.Contains(r.Target, "/echo/"):
		return r.handleEcho()
	case strings.Contains(r.Target, "/user-agent"):
		return r.handleUserAgent()
	}

	return Response{
		Version:       r.Version,
		StatusCode:    404,
		StatusMessage: "Not Found",
		ResponseHeaders: ResponseHeaders{
			ContentType:   "text/plain",
			ContentLength: 0,
		},
		ResponseBody: ResponseBody{
			Message: "",
		},
	}
}

func (r *Request) handleUserAgent() Response {

	//should check if it uses the get method and not any other HTTP method
	str := strings.TrimSpace(r.RequestHeaders.UserAgent)
	stringLenght := len(str)
	return Response{
		Version:       r.Version,
		StatusCode:    200,
		StatusMessage: "OK",
		ResponseHeaders: ResponseHeaders{
			ContentType:   "text/plain",
			ContentLength: stringLenght,
		},
		ResponseBody: ResponseBody{
			Message: str,
		},
	}
}
func (r *Request) handleEcho() Response {
	str := strings.TrimPrefix(r.Target, "/echo/")
	stringLenght := len(str)
	return Response{
		Version:       r.Version,
		StatusCode:    200,
		StatusMessage: "OK",
		ResponseHeaders: ResponseHeaders{
			ContentType:   "text/plain",
			ContentLength: stringLenght,
		},
		ResponseBody: ResponseBody{
			Message: str,
		},
	}
}

func (r *Request) WriteResponse(resp Response) []byte {
	newResp := fmt.Sprintf(
		"%s %d %s\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s",
		resp.Version,
		resp.StatusCode,
		resp.StatusMessage,
		resp.ResponseHeaders.ContentType,
		resp.ResponseHeaders.ContentLength,
		resp.ResponseBody.Message,
	)
	log.Printf("Formatted response: %s\n", newResp)
	return []byte(newResp)
}
