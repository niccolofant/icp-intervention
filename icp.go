package intervention

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
)

func LoadIntentity() (identity.Identity, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	pem, err := os.ReadFile(filepath.Join(wd, "identity.pem"))
	if err != nil {
		return nil, fmt.Errorf("failed to read identity file: %w", err)
	}

	iid, err := identity.NewSecp256k1IdentityFromPEMWithoutParameters(pem)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity from pem: %w", err)
	}

	return iid, nil
}

// GetAgent returns an agent for the Internet Computer, using the given identity.
func GetAgent(id identity.Identity) (*agent.Agent, error) {
	agent, err := agent.New(agent.Config{
		ClientConfig: []agent.ClientOption{},
		Identity:     id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}
