package pubsub

type Message struct {
	Type            string                 `json:"type"`
	DateTime        int64                  `json:"dt,omitempty"`
	Payload         map[string]interface{} `json:"payload"`
	ToAllRecipients bool                   `json:"toAllRecipients"`
	FromIP          string                 `json:"fromIP,omitempty"`
	UserIDs         []string               `json:"userIDs,omitempty"`
}
