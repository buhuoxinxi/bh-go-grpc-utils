package main

import (
	_ "buhuoxinxi/bh-go-grpc-utils/testdata"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"time"

	"buhuoxinxi/bh-go-grpc-utils"
	ecpb "google.golang.org/grpc/examples/features/proto/echo"
)

func main() {
	conn := bhgrpcutils.NewClient()

	client := ecpb.NewEchoClient(conn)

	ticker := time.NewTicker(time.Second)

	for {

		resp, err := client.UnaryEcho(
			context.Background(),
			&ecpb.EchoRequest{Message: "this is client request msg"},
		)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Printf("client.UnaryEcho resp : %v", resp)

		<-ticker.C
	}
}
