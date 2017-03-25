package main

import (
	"fastapi2"

	"fastapi2/fastapi2_toy/services/module1"
	"log"
)

func main() {
	app := fastapi2.New()
	app.Register(&module1.Service{})

	server, err := app.Listen("tcp", "0.0.0.0:0", nil)
	if err != nil {
		log.Fatal("setup server failed:", err)
	}

	go server.Serve()

	client, err := app.Dial("tcp", server.Listener().Addr().String())
	if err != nil {
		log.Fatal("setup client failed:", err)
	}

	for i := 0; i < 10; i++ {
		err := client.Send(&module1.AddReq{i, i})
		if err != nil {
			log.Fatal("send failed:", err)
		}

		rsp, err := client.Receive()
		if err != nil {
			log.Fatal("recv failed:", err)
		}

		log.Printf("AddRsp: %d", rsp.(*module1.AddRsp).C)
	}

	server.Stop()

	log.Printf("============")
}
