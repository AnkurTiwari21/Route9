package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/AnkurTiwari21/app/models"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Ensures gofmt doesn't remove the "net" import in stage 1 (feel free to remove this!)
var _ = net.ListenUDP

func main() {

	//setting up a custom resolver
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(100000),
			}
			return d.DialContext(ctx, network, os.Args[2])
		},
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		logrus.Infoln("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logrus.Infoln("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	//after succesfull setup make a redis connection
	err = godotenv.Load(".env")
	if err != nil {
		logrus.Error("Error loading .env file")
		return
	}
	addr := os.Getenv("REDIS_ADDRESS")
	pass := os.Getenv("REDIS_PASSWORD")
	redisClient := models.InitRedisClient(addr, pass)

	route9 := `
	
██████╗░░█████╗░██╗░░░██╗████████╗███████╗░█████╗░
██╔══██╗██╔══██╗██║░░░██║╚══██╔══╝██╔════╝██╔══██╗
██████╔╝██║░░██║██║░░░██║░░░██║░░░█████╗░░╚██████║
██╔══██╗██║░░██║██║░░░██║░░░██║░░░██╔══╝░░░╚═══██║
██║░░██║╚█████╔╝╚██████╔╝░░░██║░░░███████╗░█████╔╝
╚═╝░░╚═╝░╚════╝░░╚═════╝░░░░╚═╝░░░╚══════╝░╚════╝░
	`

	fmt.Println(route9)
	logrus.Infoln("Logs from your program will appear here!")
	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			logrus.Infoln("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		logrus.Info(buf[:size])
		logrus.Infof("Received %d bytes from %s: %s\n", size, source, receivedData)

		// Create an empty response
		response := models.Message{
			Question: models.Question{},
		}

		//creating header from request from client

		//setting up flag
		initialPos := 12
		domain := []string{}
		questionBytes := []byte{}
		answerBytes := []byte{}

		questionCountByte := buf[4:6]
		questionCount := binary.BigEndian.Uint16(questionCountByte)

		for initialPos < size && questionCount > 0 {
			domainNameBytes, pos := DecodeDNSName(buf[:size], uint16(initialPos))

			intermediateQuestionBytes := response.Question.SetAllDataAndReturnQuestionBytes(string(domainNameBytes), 1, 1)
			questionBytes = append(questionBytes, intermediateQuestionBytes...)

			requestToResolver := models.Message{}
			headerForResolver := requestToResolver.Header.SetRemainingDataAndReturnBytes(buf[:size], int(1))

			finalQueryToSendToResolver := headerForResolver
			finalQueryToSendToResolver = append(finalQueryToSendToResolver, intermediateQuestionBytes...)

			//before doing lookup first query the cache to see if the record is present
			ip := ""

			cachedIp, err := redisClient.Get(context.Background(), string(domainNameBytes)).Result()
			if err != nil && err != redis.Nil {
				logrus.Error("error getting data from redis | err ", err)
			}
			ip = cachedIp //set ip to the cached ip

			if cachedIp != "" {
				//update the ttl for this key value pair in redis
				redisClient.Set(context.Background(), string(domainNameBytes), cachedIp, time.Second*60)
			}

			if err == redis.Nil {
				ips, err := resolver.LookupIP(context.Background(), "ip4", string(domainNameBytes))
				if err != nil {
					logrus.Error("error in ip lookup | err ", err)
					return
				}
				ip = ips[0].To4().String()
				redisClient.Set(context.Background(), string(domainNameBytes), ip, time.Second*60)
			}

			logrus.Infof("domain %s : %+v \n", string(domainNameBytes), ip)

			domain = append(domain, string(domainNameBytes))
			answerBytes = append(answerBytes, response.Answer.FillAnswerAndReturnBytes(string(domainNameBytes), 1, 1, 60, 4, ip)...)

			initialPos = int(pos)
			initialPos += 4
			questionCount--
		}

		headerBytes := response.Header.SetRemainingDataAndReturnBytes(buf[:size], int(len(domain))) //sending remaining data and getting header bytes
		responseBytes := response.Bytes(headerBytes)

		responseBytes = append(responseBytes, questionBytes...)
		responseBytes = append(responseBytes, answerBytes...)

		_, err = udpConn.WriteToUDP(responseBytes, source)
		if err != nil {
			logrus.Infoln("Failed to send response:", err)
		}
	}
}

func DecodeDNSName(data []byte, start uint16) (string, uint16) {
	var name bytes.Buffer
	pos := start

	for {
		// logrus.Info("pos is ", pos)
		// logrus.Info(" ")
		length := uint16(data[pos])
		if length == 0 {
			// End of the name (0x00)
			pos++
			break
		}
		//if msb 2 bits are set then add a dot and make start as the number with msb removed
		if ((length & (uint16(1) << 7)) != 0) && ((length & (uint16(1) << 6)) != 0) {
			//msb 2 bits are simultaneously set
			bufferTransport := []byte{}
			bufferTransport = append(bufferTransport, data[pos])
			bufferTransport = append(bufferTransport, data[pos+1])

			transfer := (binary.BigEndian.Uint16(bufferTransport) & 0x3FFF)
			str, _ := DecodeDNSName(data, transfer)
			pos += 2
			return name.String() + str, pos
		}
		pos++
		name.Write(data[pos : pos+length])
		pos += length
		if data[pos] != 0 {
			name.WriteByte('.')
		}
	}
	return name.String(), pos
}
