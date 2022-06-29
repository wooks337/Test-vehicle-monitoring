package handler

import (
	"encoding/json"
	"net/http"
	"test-vehcile-monitoring/message"
)

type VehicleServiceHandler BaseHandler

func (h *VehicleServiceHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	requestData := message.RequestData{}
	if err := parseHttpRequestParameter(r, &requestData); err != nil {
		//replyError(w. )
	}

	type Message struct {
		Result string
	}

	msg := Message{
		Result: "OK",
	}

	json.NewEncoder(w).Encode(&msg)
}
