package bobpad

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/principal"
)

type BobPad struct {
	agent      *agent.Agent
	canisterID principal.Principal
}

func New(agent *agent.Agent) *BobPad {
	return &BobPad{
		agent:      agent,
		canisterID: principal.MustDecode("cau4v-ziaaa-aaaas-amqta-cai"),
	}
}

